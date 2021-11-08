#!/usr/bin/env bash

set -e

show_help() {
  cat <<EOF
Usage: $(basename "$0") <options>
    -h, --help               Display help
    -p, --prefix             (Optional) Prefix used to filter git tags. Default: (empty)
    -b, --main-branch        (Optional) Repository's main branch. Default: 'main'
    -y, --pypi-compatible    Generate PyPi compatible version. Default: (disabled)
EOF
}

parse_command_line(){
  while :; do
    case "${1:-}" in
      -h | --help)
        show_help
        exit
        ;;
      -p | --prefix)
        if [[ -n "${2:-}" ]]; then
          tag_prefix="$2"
          shift
        else
          echo "ERROR: '-p|--prefix' cannot be empty." >&2
          show_help
          exit 1
        fi
        ;;
      -b | --main-branch)
        if [[ -n "${2:-}" ]]; then
          main_branch="$2"
          shift
        else
          echo "ERROR: '-b|--main-branch' cannot be empty." >&2
          show_help
          exit 1
        fi
        ;;
      -y | --pypi-compatible)
        pypi_compatible=true
        ;;
      *)
        break
        ;;
    esac

    shift
  done
}

get_latest_version_tag(){
  local tag_regex="^${tag_prefix}v?([0-9]+(\.[0-9]+)*(-.*)?)$"
  local latest_tag=$(git tag -l --sort=-creatordate | grep -E "$tag_regex" | head -n 1)
  if [[ ${latest_tag} =~ ${tag_regex} ]]; then
    version_tag=${latest_tag}
    version_candidate=${BASH_REMATCH[1]}
  fi
}

get_commit_count(){
  if [[ -z "$version_tag" ]]; then
  commit_count=$(git rev-list HEAD | wc -l)
  else
      commit_count=$(git rev-list "${version_tag}"..HEAD | wc -l)
  fi
  commit_count=$(echo "${commit_count}" | tr -d "[:space:]")
}

main(){
  local tag_prefix=
  local main_branch=main
  local pypi_compatible=false
  local version_tag=
  local version_candidate=0.0.0
  local commit_count=

  parse_command_line "$@"

  get_latest_version_tag
  get_commit_count


  if [[ $commit_count = 0 ]]; then
    # if the current commit is the same that was tagged with a version tag
    # then return just use this version tag value as a version
    echo "$version_candidate"
    exit 0
  fi

  local version=
  if [ "${pypi_compatible}" == true ]; then
    version="${version_candidate}.post${commit_count}"
    if [[ $(git branch --show-current) != "$main_branch" ]]; then
      version="${version}.dev"
    fi
  else
    local ref=$(git log -n 1 --pretty=format:'%h')
    version="${version_candidate}-build.${commit_count}-${ref}"
  fi

  echo "$version"
}

main "$@"