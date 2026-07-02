#!/bin/sh
# install.sh — build and install the dev CLI, with friendly dependency errors.
#
#   ./install.sh                     build from this clone into ~/.local/bin
#   ./install.sh --check             verify the Go toolchain only (Makefile hook)
#   PREFIX=/usr/local ./install.sh   install under a different prefix
#
# When Go is missing (or older than go.mod requires) the script fails with the
# install command for the detected OS, plus the list for every other OS.

set -eu

MODULE="github.com/duclqDev99/custom-cli"
GO_MIN_FALLBACK="1.22" # only used when go.mod is not next to this script
PREFIX="${PREFIX:-$HOME/.local}"
BIN_DIR="$PREFIX/bin"
REPO_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" 2>/dev/null && pwd) || REPO_DIR="$PWD"

# ---- ui (mirrors internal/ui) ------------------------------------------

if [ -t 1 ] && [ -z "${NO_COLOR:-}" ] && [ "${TERM:-}" != "dumb" ]; then
	RESET="\033[0m" BOLD="\033[1m" RED="\033[31m" GREEN="\033[32m" YELLOW="\033[33m" CYAN="\033[36m"
else
	RESET="" BOLD="" RED="" GREEN="" YELLOW="" CYAN=""
fi

ok() { printf "  ${GREEN}✓${RESET} %s\n" "$1"; }
warn() { printf "  ${YELLOW}⚠${RESET} %s\n" "$1"; }
fail() { printf "  ${RED}✗${RESET} %s\n" "$1" >&2; }
step() { printf "${CYAN}→${RESET} %s\n" "$1"; }

# ---- os detection --------------------------------------------------------

# detect_os prints one of: darwin debian fedora arch alpine suse windows linux other
detect_os() {
	case "$(uname -s)" in
	Darwin) echo darwin ;;
	Linux)
		if [ -r /etc/os-release ]; then
			. /etc/os-release
			case "${ID:-} ${ID_LIKE:-}" in
			*debian* | *ubuntu*) echo debian ;;
			*fedora* | *rhel* | *centos*) echo fedora ;;
			*arch* | *manjaro*) echo arch ;;
			*alpine*) echo alpine ;;
			*suse*) echo suse ;;
			*) echo linux ;;
			esac
		else
			echo linux
		fi
		;;
	MINGW* | MSYS* | CYGWIN*) echo windows ;;
	*) echo other ;;
	esac
}

os_label() {
	case "$1" in
	darwin) echo "macOS" ;;
	debian) echo "Debian/Ubuntu" ;;
	fedora) echo "Fedora/RHEL" ;;
	arch) echo "Arch Linux" ;;
	alpine) echo "Alpine" ;;
	suse) echo "openSUSE" ;;
	windows) echo "Windows" ;;
	linux) echo "Linux" ;;
	*) echo "this machine" ;;
	esac
}

# go_install_cmd prints the Go install command for a detect_os value.
go_install_cmd() {
	case "$1" in
	darwin) echo "brew install go" ;;
	debian) echo "sudo apt-get install -y golang-go" ;;
	fedora) echo "sudo dnf install -y golang" ;;
	arch) echo "sudo pacman -S go" ;;
	alpine) echo "sudo apk add go" ;;
	suse) echo "sudo zypper install go" ;;
	windows) echo "winget install GoLang.Go" ;;
	*) echo "download from https://go.dev/dl/" ;;
	esac
}

# go_install_help prints install commands: detected OS first, then every OS.
go_install_help() {
	os=$(detect_os)
	printf "\n"
	printf "    ${BOLD}install Go on %s:${RESET}\n\n" "$(os_label "$os")"
	printf "      ${BOLD}${CYAN}%s${RESET}\n\n" "$(go_install_cmd "$os")"
	printf "    ${BOLD}other operating systems:${RESET}\n\n"
	printf "      %-16s%s\n" "macOS" "brew install go"
	printf "      %-16s%s\n" "Debian/Ubuntu" "sudo apt-get install -y golang-go"
	printf "      %-16s%s\n" "Fedora/RHEL" "sudo dnf install -y golang"
	printf "      %-16s%s\n" "Arch Linux" "sudo pacman -S go"
	printf "      %-16s%s\n" "Alpine" "sudo apk add go"
	printf "      %-16s%s\n" "openSUSE" "sudo zypper install go"
	printf "      %-16s%s\n" "Windows" "winget install GoLang.Go   (or: choco install golang / scoop install go)"
	printf "      %-16s%s\n" "any OS" "download from https://go.dev/dl/"
	printf "\n"
	printf "    then re-run the install.\n\n"
}

# ---- checks ---------------------------------------------------------------

# check_go verifies the Go toolchain exists and satisfies go.mod's minimum.
check_go() {
	if ! command -v go >/dev/null 2>&1; then
		fail "Go is not installed — dev is built from Go source"
		go_install_help >&2
		return 1
	fi

	gover=$(go env GOVERSION 2>/dev/null || true)
	if [ -z "$gover" ]; then
		gover=$(go version 2>/dev/null | awk '{print $3}')
	fi

	min="$GO_MIN_FALLBACK"
	if [ -f "$REPO_DIR/go.mod" ]; then
		parsed=$(awk '$1 == "go" {print $2; exit}' "$REPO_DIR/go.mod")
		if [ -n "$parsed" ]; then
			min="$parsed"
		fi
	fi

	have=$(printf '%s' "$gover" | sed 's/^go1\.//; s/[^0-9].*$//')
	want=$(printf '%s' "$min" | sed 's/^1\.//; s/[^0-9].*$//')
	if [ -n "$have" ] && [ -n "$want" ] && [ "$have" -lt "$want" ]; then
		fail "Go $gover is too old — dev needs Go >= $min"
		go_install_help >&2
		return 1
	fi

	ok "Go ${gover:-found} ($(command -v go))"
}

# ---- install ----------------------------------------------------------------

install_dev() {
	mkdir -p "$BIN_DIR"
	if [ -f "$REPO_DIR/go.mod" ] && [ -d "$REPO_DIR/cmd/dev" ]; then
		version=$(git -C "$REPO_DIR" describe --tags --always --dirty 2>/dev/null || echo 0.1.0)
		(cd "$REPO_DIR" && go build -ldflags "-s -w -X main.version=$version" -o "$BIN_DIR/dev" ./cmd/dev)
		ok "installed $BIN_DIR/dev ($version)"
	else
		# not run from a clone — build the published module instead
		GOBIN="$BIN_DIR" go install "$MODULE/cmd/dev@latest"
		ok "installed $BIN_DIR/dev (latest)"
	fi
}

# path_hint warns when BIN_DIR is not on PATH and prints the fix for the shell.
path_hint() {
	case ":$PATH:" in
	*":$BIN_DIR:"*)
		ok "$BIN_DIR is on your PATH"
		return 0
		;;
	esac

	rc="$HOME/.profile"
	case "$(basename "${SHELL:-/bin/sh}")" in
	zsh) rc="$HOME/.zshrc" ;;
	bash) rc="$HOME/.bashrc" ;;
	fish) rc="" ;;
	esac

	warn "$BIN_DIR is not on your PATH — add it permanently:"
	if [ -z "$rc" ]; then
		printf "      ${BOLD}fish_add_path %s${RESET}\n" "$BIN_DIR"
	else
		printf "      ${BOLD}echo 'export PATH=\"%s:\$PATH\"' >> %s${RESET}\n" "$BIN_DIR" "$rc"
		printf "      ${BOLD}source %s${RESET}\n" "$rc"
	fi
}

usage() {
	printf "${BOLD}install.sh${RESET} — build & install the dev CLI\n\n"
	printf "  ./install.sh            check Go, build, install to %s\n" "$BIN_DIR"
	printf "  ./install.sh --check    verify the Go toolchain only\n"
	printf "  PREFIX=/usr/local ./install.sh\n"
}

# ---- main -------------------------------------------------------------------

case "${1:-}" in
--check | check)
	check_go
	exit 0
	;;
-h | --help | help)
	usage
	exit 0
	;;
"") ;;
*)
	fail "unknown option: $1"
	usage >&2
	exit 1
	;;
esac

step "checking the Go toolchain"
check_go
step "building dev"
install_dev
path_hint
printf "\n"
ok "done — try: ${BOLD}dev init${RESET}"
