package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func runAdd(args []string) error {
	return runConfigFeatureToggle(args, true)
}

func runRemove(args []string) error {
	return runConfigFeatureToggle(args, false)
}

func runConfigFeatureToggle(args []string, enable bool) error {
	cmdName := "add"
	if !enable {
		cmdName = "remove"
	}

	fs := flag.NewFlagSet(cmdName, flag.ContinueOnError)
	projectDir := fs.String("dir", "", "project root directory (default: auto-detect from current directory)")

	var featuresArg string
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		featuresArg = strings.TrimSpace(args[0])
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
	} else {
		if err := fs.Parse(args); err != nil {
			return err
		}
		if fs.NArg() < 1 {
			return fmt.Errorf("feature list is required, e.g. go-infra-cli %s redis", cmdName)
		}
		featuresArg = strings.TrimSpace(fs.Arg(0))
	}

	features, err := parseConfigFeaturesArg(featuresArg)
	if err != nil {
		return err
	}

	root, err := resolveProjectDir(*projectDir)
	if err != nil {
		return err
	}

	anyChanged := false
	for _, feature := range features {
		artifactChanged, err := syncConfigFeatureArtifacts(root, feature, enable)
		if err != nil {
			return fmt.Errorf("%s: %w", feature, err)
		}
		if artifactChanged {
			anyChanged = true
			verb := "installed"
			if !enable {
				verb = "removed"
			}
			fmt.Printf("%s: %s\n", feature, verb)
		} else {
			if enable && isConfigFeatureInstalled(root, feature) {
				fmt.Printf("%s: already installed (skipped)\n", feature)
			}
			if !enable && !isConfigFeatureInstalled(root, feature) {
				fmt.Printf("%s: already removed (skipped)\n", feature)
			}
		}
		if enable {
			printAddFeatureHints(feature, root)
		}
	}

	if modulePath, err := readModulePath(root); err == nil {
		if err = replaceStarterStrings(root, modulePath); err != nil {
			return fmt.Errorf("rewrite module paths: %w", err)
		}
	}

	if !anyChanged {
		fmt.Println("no file changes (already in target state)")
	} else {
		fmt.Printf("\nproject: %s\nnext: adjust configs if needed and restart the service\n", root)
	}
	return nil
}

func resolveProjectDir(explicit string) (string, error) {
	if explicit != "" {
		abs, err := filepath.Abs(explicit)
		if err != nil {
			return "", err
		}
		if err := validateScaffoldProject(abs); err != nil {
			return "", err
		}
		return abs, nil
	}
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if err := validateScaffoldProject(wd); err == nil {
			return wd, nil
		}
		parent := filepath.Dir(wd)
		if parent == wd {
			break
		}
		wd = parent
	}
	return "", errors.New("cannot find a go-infra-starter style project (need go.mod, configs/config.yml, and standard internal layout); use --dir")
}

func validateScaffoldProject(dir string) error {
	if err := validateProjectRoot(dir); err != nil {
		return err
	}
	return validateStarterLayout(dir)
}

func validateProjectRoot(dir string) error {
	if !fileExists(filepath.Join(dir, "go.mod")) {
		return errors.New("missing go.mod")
	}
	if !fileExists(filepath.Join(dir, "configs", "config.yml")) {
		return errors.New("missing configs/config.yml")
	}
	return nil
}

func printAddFeatureHints(feature, projectDir string) {
	switch feature {
	case "mysql":
		configPath := filepath.Join(projectDir, "configs", "config.yml")
		data, err := os.ReadFile(configPath)
		if err != nil {
			fmt.Println("mysql: set mysql.dsn in configs/config.yml before use")
			return
		}
		if strings.Contains(string(data), `dsn: ""`) {
			fmt.Println("mysql: set mysql.dsn in configs/config.yml (currently empty)")
		}
	case "redis":
		fmt.Println("redis: verify redis.addresses/password in configs/config.yml")
	}
}
