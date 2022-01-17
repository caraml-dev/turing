#!/usr/bin/env bash

set -o errexit
set -o pipefail

DEFAULT_SWAGGER_UI_VERSION=v3.52.3
SWAGGER_CONFIG_FILENAME=swagger-config.yaml
REQUIRED_FILES=(
  "swagger-ui-bundle.js"
  "swagger-ui-standalone-preset.js"
  "swagger-ui.css"
  "favicon-16x16.png"
  "favicon-32x32.png"
)

show_help() {
  cat <<EOF
Usage: $(basename "$0") <options>
    -h, --help               Display help
    -s, --spec-url           URL containing the API definition
    -o, --output             The output path for the generated Swagger UI files
    -v, --version            The Swagger-UI version to use (default: DEFAULT_SWAGGER_UI_VERSION)"
EOF
}

main() {
  local version="$DEFAULT_SWAGGER_UI_VERSION"
  local output=
  local spec_file=

  parse_command_line "$@"

  download_swagger_ui

  make_swagger_config

  make_index_file

  echo "Done. Static Swagger UI distribution can be served from the $output folder"
}

parse_command_line() {
  while :; do
    case "${1:-}" in
      -h | --help)
        show_help
        exit
        ;;
      -v | --version)
        if [[ -n "${2:-}" ]]; then
          version="$2"
          shift
        else
          echo "ERROR: '-v|--version' cannot be empty." >&2
          show_help
          exit 1
        fi
        ;;
      -o | --output)
        if [[ -n "${2:-}" ]]; then
          output="$2"
          shift
        else
          echo "ERROR: '-o|--output' cannot be empty." >&2
          show_help
          exit 1
        fi
        ;;
      -s | --spec-url)
        if [[ -n "${2:-}" ]]; then
          spec_url="$2"
          shift
        else
          echo "ERROR: '-s|--spec-url' cannot be empty." >&2
          show_help
          exit 1
        fi
        ;;
      *)
        break
        ;;
    esac

    shift
  done

  if [[ -z "$spec_url" ]]; then
    echo "ERROR: '-s|--spec-url' is required." >&2
    show_help
    exit 1
  fi

  if [[ -z "$output" ]]; then
    echo "ERROR: '-o|--output' is required." >&2
    show_help
    exit 1
  fi
}

download_swagger_ui(){
  local tarball_name=swagger-ui.tar.gz
  local tmp_dir=swagger-ui-temp
  tarball_url="https://api.github.com/repos/swagger-api/swagger-ui/tarball/$version"

  echo "Downloading Swagger UI release..."
  local wget_code=0
  wget -qO "$tarball_name" "$tarball_url" || wget_code=$?

  if [ "$wget_code" -ne "0" ]; then
    echo "ERROR: Swagger UI release $version not found." >&2
    exit 1
  else
    echo "Swagger UI distribution has been downloaded."
  fi

  mkdir -p "$tmp_dir"

  echo "Unpacking..."
  tar -xzf "$tarball_name" --strip-components=1 -C "$tmp_dir"
  rm -f "$tarball_name"

  mkdir -p "$output"

  echo "Copying required files..."
  for file in "${REQUIRED_FILES[@]}"; do
    echo "$output/$file"
    cp "$tmp_dir/dist/$file" "$output/$file"
  done
  rm -rf "$tmp_dir"
}

make_swagger_config(){
  echo "Generating Swagger UI configuration..."
  cat <<EOF > "$output/$SWAGGER_CONFIG_FILENAME"
deepLinking: true
url: "$spec_url"
EOF
}

make_index_file(){
  echo "Making swagger-ui index.html..."
  template_index_file=$(dirname "$0")"/dist/index.html"
  sed "s|<swaggerConfig>|$SWAGGER_CONFIG_FILENAME|" "$template_index_file" > "$output/index.html"
}

main "$@"
