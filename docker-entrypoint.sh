#!/bin/bash
set -e

CMD=()
TURING_UI_CONFIG_PATH=
TURING_UI_DIST_DIR=${TURING_UI_DIST_DIR:-}
TURING_UI_DIST_CONFIG_FILE=${TURING_UI_DIST_DIR}/app.config.js
TURING_API_BIN=${TURING_API_BIN:?"ERROR: TURING_API_BIN is not specified"}

show_help() {
  cat <<EOF
Usage: $(basename "$0") <options> <...>
    -ui-config               JSON file containing configuration of Turing UI
EOF
}

main(){
  parse_command_line "$@"

  if [[ -n "$TURING_UI_CONFIG_PATH" ]]; then
    echo "Turing UI config found at ${TURING_UI_CONFIG_PATH}..."
    if [[ -n "$TURING_UI_DIST_DIR" ]]; then
      echo "Overriding UI config at $TURING_UI_DIST_CONFIG_FILE"

      echo "var config = $(cat $TURING_UI_CONFIG_PATH);" > "$TURING_UI_DIST_CONFIG_FILE"

      echo "Done."
    else
      echo "TURING_UI_DIST_DIR: Turing UI static build directory not provided. Skipping."
    fi
  else
    echo "Turing UI config is not provided. Skipping."
  fi
}

parse_command_line(){
  while [[ $# -gt 0 ]]; do
    case "$1" in
      -ui-config)
        if [[ -n "$2" ]]; then
          TURING_UI_CONFIG_PATH="$2"
          shift
        else
          echo "ERROR: '-ui-config' cannot be empty." >&2
          show_help
          exit 1
        fi
        ;;
      *)
        CMD+=("$1")
        ;;
    esac

    shift
  done

  if [[ -n "$TURING_UI_CONFIG_PATH" ]]; then
    if [ ! -f "$TURING_UI_CONFIG_PATH" ]; then
      echo "ERROR: config file $TURING_UI_CONFIG_PATH does not exist." >&2
      show_help
      exit 1
    fi
  fi
}

main "$@"

echo "Launching turing-api server: " "$TURING_API_BIN" "${CMD[@]}"
exec "$TURING_API_BIN" "${CMD[@]}"

