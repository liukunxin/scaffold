package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func runMono(args []string) error {
	if len(args) < 1 {
		return errors.New("usage: go-infra-cli mono add app <name> [--dir <project-root>] | go-infra-cli mono add domain <name> [--dir <project-root>]")
	}
	switch args[0] {
	case "add":
		return runMonoAdd(args[1:])
	default:
		return fmt.Errorf("unsupported mono subcommand %q", args[0])
	}
}

func runMonoAdd(args []string) error {
	fs := flag.NewFlagSet("mono add", flag.ContinueOnError)
	projectDir := fs.String("dir", "", "monorepo root directory (default: auto-detect from current directory)")

	if len(args) < 2 {
		return errors.New("usage: go-infra-cli mono add app <name> [--dir <project-root>] | go-infra-cli mono add domain <name> [--dir <project-root>]")
	}
	kind := strings.TrimSpace(strings.ToLower(args[0]))
	name := strings.TrimSpace(args[1])
	if err := validateMonoName(name); err != nil {
		return err
	}
	if err := fs.Parse(args[2:]); err != nil {
		return err
	}

	root, err := resolveMonorepoDir(*projectDir)
	if err != nil {
		return err
	}
	modulePath, err := readModulePath(root)
	if err != nil {
		return err
	}

	switch kind {
	case "app":
		if err := addMonoApp(root, modulePath, name); err != nil {
			return err
		}
		fmt.Printf("monorepo app added: apps/%s\n", name)
	case "domain":
		if err := addMonoDomain(root, name); err != nil {
			return err
		}
		fmt.Printf("monorepo domain added: domains/%s\n", name)
	default:
		return fmt.Errorf("unsupported mono add target %q: only app|domain are allowed", kind)
	}
	return nil
}

func validateMonoName(name string) error {
	ok, err := regexp.MatchString(`^[a-z][a-z0-9-]*$`, name)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("invalid name %q: must match ^[a-z][a-z0-9-]*$", name)
	}
	return nil
}

func resolveMonorepoDir(explicit string) (string, error) {
	if explicit != "" {
		abs, err := filepath.Abs(explicit)
		if err != nil {
			return "", err
		}
		if err := validateMonorepoRoot(abs); err != nil {
			return "", err
		}
		return abs, nil
	}
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if err := validateMonorepoRoot(wd); err == nil {
			return wd, nil
		}
		parent := filepath.Dir(wd)
		if parent == wd {
			break
		}
		wd = parent
	}
	return "", errors.New("cannot find monorepo root (need go.work + internal/event + internal/runtime + apps + domains); use --dir")
}

func validateMonorepoRoot(dir string) error {
	required := []string{
		"go.work",
		"go.mod",
		"internal/event/envelope.go",
		"internal/runtime/engine.go",
		"apps",
		"domains",
	}
	for _, rel := range required {
		if !fileOrDirExists(filepath.Join(dir, filepath.FromSlash(rel))) {
			return fmt.Errorf("invalid monorepo root: missing %s", rel)
		}
	}
	return nil
}

func addMonoApp(root, rootModule, name string) error {
	appDir := filepath.Join(root, "apps", name)
	if fileOrDirExists(appDir) {
		return fmt.Errorf("app already exists: %s", appDir)
	}
	if err := os.MkdirAll(appDir, 0o755); err != nil {
		return err
	}

	files := map[string]string{
		"go.mod": fmt.Sprintf(`module %s/apps/%s

go 1.25

require (
	github.com/gin-gonic/gin v1.11.0
	github.com/liukunxin/go-infra v0.0.0
	%s v0.0.0
)

replace %s => ../..
`, rootModule, name, rootModule, rootModule),
		"cmd/main.go": fmt.Sprintf(`package main

import (
	"log"

	"%s/apps/%s/bootstrap"
)

func main() {
	app, err := bootstrap.New()
	if err != nil {
		log.Fatalf("bootstrap init failed: %%v", err)
	}
	defer app.Close()

	if err = app.Run(); err != nil {
		log.Fatalf("server run failed: %%v", err)
	}
}
`, rootModule, name),
		"bootstrap/config.go": "package bootstrap\n\n" +
			"import (\n" +
			"\tkconfig \"github.com/liukunxin/go-infra/pkg/base/config\"\n" +
			"\tklog \"github.com/liukunxin/go-infra/pkg/base/log\"\n" +
			"\t\"github.com/liukunxin/go-infra/pkg/base/trace\"\n" +
			")\n\n" +
			"type Config struct {\n" +
			"\tAppName string `yaml:\"app_name\"`\n" +
			"\tServer  struct {\n" +
			"\t\tAddress string `yaml:\"address\"`\n" +
			"\t} `yaml:\"server\"`\n" +
			"\tLog   klog.Config  `yaml:\"log\"`\n" +
			"\tTrace trace.Config `yaml:\"trace\"`\n" +
			"}\n\n" +
			"func loadConfig() (*Config, error) {\n" +
			"\treturn kconfig.Load[Config](kconfig.WithValidate(false))\n" +
			"}\n",
		"bootstrap/app.go": fmt.Sprintf(`package bootstrap

import (
	"os"

	"github.com/gin-gonic/gin"
	"github.com/liukunxin/go-infra/pkg/base/env"
	"github.com/liukunxin/go-infra/pkg/base/log"
	"github.com/liukunxin/go-infra/pkg/base/trace"
	"%s/internal/event"
	httptransport "%s/apps/%s/transport/http"
	runtimecore "%s/internal/runtime"
)

type App struct {
	cfg  *Config
	http *httptransport.Server
}

func New() (*App, error) {
	cfg, err := loadConfig()
	if err != nil {
		return nil, err
	}

	env.SetName(cfg.AppName)
	env.SetEnv(os.Getenv("env"))
	if err = log.Init(cfg.Log); err != nil {
		return nil, err
	}
	traceCfg := cfg.Trace
	serviceName := cfg.AppName
	if traceCfg.ServiceName != nil && *traceCfg.ServiceName != "" {
		serviceName = *traceCfg.ServiceName
	}
	traceCfg.ServiceName = &serviceName
	if err = trace.Init(trace.WithConfig(&traceCfg)); err != nil {
		return nil, err
	}

	dispatcher := event.NewDispatcher()
	engine := runtimecore.NewEngine(runtimecore.NewRouter(dispatcher))
	httpSrv := httptransport.New(gin.New(), engine)
	return &App{cfg: cfg, http: httpSrv}, nil
}

func (a *App) Run() error {
	return a.http.Run(a.cfg.Server.Address)
}

func (a *App) Close() {
	trace.Flush()
	log.Close()
}
`, rootModule, rootModule, name, rootModule),
		"transport/http/server.go": fmt.Sprintf(`package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/liukunxin/go-infra/pkg/biz/middlewares"
	runtimecore "%s/internal/runtime"
)

type Server struct {
	router *gin.Engine
	engine *runtimecore.Engine
}

func New(router *gin.Engine, engine *runtimecore.Engine) *Server {
	s := &Server{router: router, engine: engine}
	s.setupRoutes()
	return s
}

func (s *Server) setupRoutes() {
	s.router.Use(gin.Recovery(), middlewares.GinTraceMiddleware(), middlewares.HttpLogRecord())
	s.router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
}

func (s *Server) Run(addr string) error {
	return s.router.Run(addr)
}
`, rootModule),
		"transport/grpc/server.go":          "package grpc\n\ntype Server struct{}\n",
		"transport/ws/server.go":            "package ws\n\ntype Server struct{}\n",
		"internal/service/service.go":       "package service\n\n// Reserved for app-specific orchestration.\n",
		"internal/handler/handler.go":       "package handler\n\n// Reserved for app-specific transport handlers.\n",
		"internal/repository/repository.go": "package repository\n\n// Reserved for app-side persistence implementation.\n",
		"internal/model/model.go":           "package model\n\n// Reserved for app-level view models.\n",
		"wiring/di.go":                      "package wiring\n\n// Reserved for dependency graph wiring.\n",
		"README.md":                         fmt.Sprintf("# %s\n\nGenerated app skeleton for monorepo.\n", name),
	}

	if err := writeFiles(appDir, files); err != nil {
		return err
	}
	return appendGoWorkUse(root, "./apps/"+name)
}

func addMonoDomain(root, name string) error {
	domainDir := filepath.Join(root, "domains", name)
	if fileOrDirExists(domainDir) {
		return fmt.Errorf("domain already exists: %s", domainDir)
	}
	if err := os.MkdirAll(domainDir, 0o755); err != nil {
		return err
	}

	eventName := strings.ReplaceAll(name, "-", ".")
	files := map[string]string{
		"api/event.go": fmt.Sprintf(`package api

const EventDefault = "domain.%s.default"
`, eventName),
		"api/contract.go": `package api

import "context"

type Contract interface {
	Handle(ctx context.Context, payload map[string]any) (map[string]any, error)
}
`,
		"service/core.go": `package service

import "context"

type Core struct{}

func NewCore() *Core { return &Core{} }

func (c *Core) Handle(_ context.Context, payload map[string]any) (map[string]any, error) {
	return payload, nil
}
`,
		"service/processor.go":    "package service\n\n// Processor is reserved for richer domain orchestration.\ntype Processor struct{}\n",
		"model/entity.go":         "package model\n\ntype Entity struct{}\n",
		"adapter/external_api.go": "package adapter\n\n// Reserved for domain-level external integrations.\n",
		"runtime/handler.go": `package runtime

import (
	"context"

	"REPLACE_MODULE/internal/event"
	domainapi "REPLACE_MODULE/domains/REPLACE_NAME/api"
)

func NewHandler(contract domainapi.Contract) func(ctx context.Context, evt event.Envelope) (event.Envelope, error) {
	return func(ctx context.Context, evt event.Envelope) (event.Envelope, error) {
		out, err := contract.Handle(ctx, evt.Payload)
		if err != nil {
			return event.Envelope{}, err
		}
		return event.Envelope{
			EventID:   evt.EventID,
			EventType: domainapi.EventDefault + ".result",
			SessionID: evt.SessionID,
			Seq:       evt.Seq,
			Timestamp: evt.Timestamp,
			Payload:   out,
		}, nil
	}
}
`,
		"README.md": fmt.Sprintf("# %s\n\nGenerated domain skeleton for monorepo.\n", name),
	}

	modulePath, err := readModulePath(root)
	if err != nil {
		return err
	}
	files["runtime/handler.go"] = strings.ReplaceAll(files["runtime/handler.go"], "REPLACE_MODULE", modulePath)
	files["runtime/handler.go"] = strings.ReplaceAll(files["runtime/handler.go"], "REPLACE_NAME", name)

	return writeFiles(domainDir, files)
}

func writeFiles(base string, files map[string]string) error {
	for rel, content := range files {
		path := filepath.Join(base, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			return err
		}
	}
	return nil
}

func appendGoWorkUse(root, usePath string) error {
	path := filepath.Join(root, "go.work")
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	content := string(data)
	if strings.Contains(content, "\n\t"+usePath+"\n") || strings.Contains(content, "\n\t"+usePath+"\r\n") {
		return nil
	}

	lines := strings.Split(content, "\n")
	out := make([]string, 0, len(lines)+1)
	inUseBlock := false
	inserted := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "use (") {
			inUseBlock = true
			out = append(out, line)
			continue
		}
		if inUseBlock && trimmed == ")" && !inserted {
			out = append(out, "\t"+usePath)
			inserted = true
		}
		if inUseBlock && trimmed == ")" {
			inUseBlock = false
		}
		out = append(out, line)
	}
	if !inserted {
		out = append(out, "use (", "\t"+usePath, ")")
	}
	return os.WriteFile(path, []byte(strings.Join(out, "\n")), 0o644)
}
