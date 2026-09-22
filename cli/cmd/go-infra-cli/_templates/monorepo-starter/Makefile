# Monorepo 根 Makefile。
#
# 只做跨 Project 的编排，并且**不假设所有 Project 都是 Go**。
# Project 自己的构建细节属于 Project（见 services/gateway/Makefile）。

GO_MODULES := $(patsubst %/,%,$(dir $(wildcard apps/*/go.mod services/*/go.mod packages/*/*/go.mod)))

PYTHON ?= python
NPM    ?= npm

.PHONY: help list check structure fmt go-tidy go-build go-test \
        web-install web-dev web-build py-check

help:
	@echo "Overview / checks"
	@echo "  make list        list every Go Project in this repository"
	@echo "  make structure   run tools/lint/check-structure.py (directory conventions)"
	@echo "  make check       fmt + go-tidy + go-build + go-test + structure"
	@echo ""
	@echo "Go (runs per module: the workspace root is not a module)"
	@echo "  make go-tidy     go work sync + go mod tidy in every module"
	@echo "  make go-build    go build ./... in every module"
	@echo "  make go-test     go test ./... in every module"
	@echo "  make fmt         gofmt -w in every Go module"
	@echo ""
	@echo "Non-Go Projects (only apps/web today; change these targets when you swap the stack)"
	@echo "  make web-install npm install --prefix apps/web"
	@echo "  make web-dev     npm run dev --prefix apps/web"
	@echo "  make web-build   npm run build --prefix apps/web"
	@echo "  make py-check    import-check the packages/python/* bindings"
	@echo ""
	@echo "Note: the go.work workspace root is not a module, so do not run go build ./..."
	@echo "inside it (it fails with 'directory prefix . does not contain modules')."
	@echo "The targets above cd into each module directory."

list:
	@for m in $(GO_MODULES); do echo "$$m"; done

structure:
	@$(PYTHON) tools/lint/check-structure.py

go-tidy:
	go work sync
	@for m in $(GO_MODULES); do echo "==> tidy $$m"; (cd $$m && go mod tidy) || exit 1; done

go-build:
	@for m in $(GO_MODULES); do echo "==> build $$m"; (cd $$m && go build ./...) || exit 1; done

go-test:
	@for m in $(GO_MODULES); do echo "==> test $$m"; (cd $$m && go test ./...) || exit 1; done

fmt:
	@for m in $(GO_MODULES); do echo "==> fmt $$m"; (cd $$m && gofmt -w .) || exit 1; done

web-install:
	$(NPM) install --prefix apps/web

web-dev:
	$(NPM) run dev --prefix apps/web

web-build:
	$(NPM) run build --prefix apps/web

py-check:
	@$(PYTHON) -c "import sys; sys.path.insert(0, 'packages/python/contracts/src'); import contracts; env = contracts.new_envelope('e1', 'demo.ping.completed.v1', {'message': 'pong'}); assert env.to_wire()['eventId'] == 'e1'; print('packages/python/contracts: ok')"

check: fmt go-tidy go-build go-test structure
