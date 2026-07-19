
# ---------------------------------------------------------------------------- #
# --------------------------- INSTALL MISSING DEPS --------------------------- #
# ---------------------------------------------------------------------------- #
declare -A deps=(
  ["gofumpt"]="mvdan.cc/gofumpt@latest"
  ["gopls"]="golang.org/x/tools/cmd/gopls@latest"
)

for dep in "${!deps[@]}"; do
  command -v "${dep}" &> /dev/null || go install "${deps[$dep]}"
done

# ---------------------------------------------------------------------------- #
# -------------------------------- PARSE ARGS -------------------------------- #
# ---------------------------------------------------------------------------- #
case "$1" in
  fmt)
	  go mod tidy
    gofumpt -w .
    ;;

  build)
  # Build the application into a static binary
	# https://stackoverflow.com/a/61324538/22415851
	  CGO_ENABLED=0 go build -a -ldflags '-extldflags "-static"' .
    ;;
  *)
    echo "Usage: ${0} [fmt|build]" 1>&2
    ;;
esac
