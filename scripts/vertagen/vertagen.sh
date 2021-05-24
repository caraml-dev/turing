#!/bin/bash
#
# If run inside a git repository will return a valid semver based on
# the semver formatted tags. For example if the current HEAD is tagged
# at 0.0.1, then the version echoed will simply be 0.0.1. However, if
# the tag is say, 3 patches behind, the tag will be in the form
# `0.0.1-build.3+0ace960`. This is basically, the current tag a
# monotonically increasing commit (the number of commits since the
# tag, and then a git short ref to identify the commit.
#
# You may also pass this script the a `release` argument. If that is
# the case it will exit with a non-zero value if the current head is
# not tagged.
#
set -e

format=
version_prefix=
commit_count=
main_branch_name="main"
version_candidate="0.0.0"
version_tag=""
vsn=

function getVersionCandidate() {
    local commit_hash_regex='^([0-9a-f]+)'
    local latest_tag_line=`git log --oneline --decorate  |  fgrep "tag: " | head -n 1`

    if [[ $latest_tag_line =~ $commit_hash_regex ]]; then
        local tagged_commit_hash=${BASH_REMATCH[1]}
        # For annotated tags, --sort=-taggerdate will ensure that if, as an example, v0.0.5 and
        # v0.0.5-rc1 point to the same commit, the newer tag (v0.0.5) is used for the version number
        # generation instead of the alphabetically latest version (v0.0.5-rc1).
        local latest_tag=`git tag --sort=-taggerdate --points-at ${tagged_commit_hash} | head -n 1`
        # This script expects the tags to be in one of the following formats:
        # – <component_prefix>/v0.0.0-<suffix> – for releases of individual components of this repo
        #   Example: batch-ensembler/v0.0.1, router/v0.4.1-rc1 etc
        # – (v)0.0.0-<suffix> – for releases of the entire repo
        #   Example: v2.3.1, v0.0.1-alpha1, 10.0.1
        local version_regex='((.+/)?(v([^/]+)|([0-9]+(\.[0-9]+)*)))'
        if [[ $latest_tag =~ $version_regex ]]; then
            if [[ ${BASH_REMATCH[3]:0:1} = "v" ]]; then
                version_tag=${BASH_REMATCH[1]}
                version_candidate=${BASH_REMATCH[4]}
            else
                version_tag=${BASH_REMATCH[1]}
                version_candidate=${BASH_REMATCH[5]}
            fi
        fi
    fi
}

function getCommitCount() {
    if [[ $version_tag = "" ]]; then
        commit_count=`git rev-list HEAD | wc -l`
    else
        commit_count=`git rev-list ${version_tag}..HEAD | wc -l`
    fi
    commit_count=`echo $commit_count | tr -d ' 't`
}

function buildVersion() {
    if [[ $commit_count = 0 ]]; then
        vsn=$version_candidate
    else
        local ref=`git log -n 1 --pretty=format:'%h'`
        local current_branch=`git branch --show-current`
        if [[ $format == "docker" ]]; then
          vsn="${version_prefix}${version_candidate}-build.${commit_count}-${ref}"
        elif [[ $format == "pypi" ]]; then
          local suffix=""
          if [[ $current_branch != "$main_branch_name" ]]; then
            suffix=".dev"
          fi
          vsn="${version_prefix}${version_candidate}.post${commit_count}${suffix}"
        else
          vsn="${version_prefix}${version_candidate}-build.${commit_count}+${ref}"
        fi
    fi
}


function usage() {
  echo "usage: vertagen [[-p <version_prefix>] | [-m <main_branch_name>] | [-f docker] ] | [-h]]"
}

while [ "$1" != "" ]; do
    case $1 in
        -p | --prefix )           shift
                                  version_prefix=$1
                                  ;;
        -f | --format )           shift
                                  format=$1
                                  ;;
        -m | --main-branch-name ) shift
                                  main_branch_name=$1
                                  ;;
        -h | --help )             usage
                                  exit
                                  ;;
        * )                       usage
                                  exit 1
    esac
    shift
done

getVersionCandidate
getCommitCount
buildVersion

echo $vsn
