#!/usr/bin/env sh
# 渲染冒烟：不联网验证两种布局都能生成，且生成物符合目录约定。
#
# CI（.github/workflows/ci.yml 的 templates job）与本地开发共用这一份 ——
# 逻辑只写一次，避免"CI 能过、本地漏跑"。
#
# 用法：
#   sh scripts/smoke.sh
#   PYTHON=python sh scripts/smoke.sh        # 指定 Python 解释器（默认 python3）
#
# 平台说明：
# - 脚本自身是 POSIX sh（CI 上是 dash），不要用 bash 专有语法。
# - Windows 请在 WSL 里跑。Git Bash 的 MSYS 会改写传给 Windows 版 go.exe / CLI 的 POSIX 路径，
#   生成物会落到别处（现象是"跑完了但临时目录是空的"）。

set -eu

repo_root=$(cd "$(dirname "$0")/.." && pwd)
python_bin=${PYTHON:-python3}
work=$(mktemp -d)
trap 'rm -rf "$work" || true' EXIT

echo "== build cli =="
cli="$work/go-infra-cli"
(cd "$repo_root/cli" && go build -o "$cli" ./cmd/go-infra-cli)
# Windows 上 Go 会给不带扩展名的 -o 自动补 .exe（WSL / Linux 不会走到这个分支）。
if [ ! -f "$cli" ] && [ -f "$cli.exe" ]; then
	cli="$cli.exe"
fi

cd "$work"

# 断言生成物里不再残留模板名（不联网，只查文件内容）。
assert_no_template_name() {
	file=$1
	name=$2
	if grep -F "$name" "$file" >/dev/null 2>&1; then
		echo "smoke: template name $name still present in $file" >&2
		exit 1
	fi
}

echo "== init --layout single =="
"$cli" init smoke-single --module github.com/example/smoke-single --layout single --skip-tidy
test -f smoke-single/go.mod
test -f smoke-single/Makefile
test -f smoke-single/Dockerfile
assert_no_template_name smoke-single/go.mod single-starter

echo "== init --layout monorepo =="
"$cli" init smoke-mono --module github.com/example/smoke-mono --layout monorepo --skip-tidy
test -f smoke-mono/go.work
test -f smoke-mono/services/gateway/go.mod
test -f smoke-mono/services/gateway/Dockerfile
test -f smoke-mono/packages/go/contracts/go.mod
test -f smoke-mono/apps/web/package.json
assert_no_template_name smoke-mono/go.work monorepo-starter
assert_no_template_name smoke-mono/services/gateway/go.mod monorepo-starter
assert_no_template_name smoke-mono/packages/go/contracts/go.mod monorepo-starter

echo "== structure check =="
cd smoke-mono
"$python_bin" tools/lint/check-structure.py

echo "smoke: OK"
