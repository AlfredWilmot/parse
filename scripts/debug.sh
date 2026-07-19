DLV_PORT=4444
STDIN_STR="test"
FILE_UNDER_TEST=main.go

# install go debugger (dlv) if missing -- assumes ${HOME}/go/bin is in ${PATH}
if ! command -v dlv &> /dev/null ; then
  echo "test"
  go install github.com/go-delve/delve/cmd/dlv@latest
fi

case "${1}" in
  # create a dlv server from a main.go file, for a client to interact with from a separate terminal
  server)
    shift
    echo "${STDIN_STR}" | dlv debug --headless -r /dev/stdin -l "127.0.0.1:${DLV_PORT}" "${FILE_UNDER_TEST}" -- "${@}"
    ;;
  # establish a client connection to a dlv server running in a separate terminal
  client)
    dlv connect ":${DLV_PORT}"
    ;;
  *)
    echo "Usage: ${0} [server|client]" 1>&2
    ;;
esac
