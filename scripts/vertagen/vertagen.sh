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
version_candidate=
version_tag=
vsn=

function getVersionCandidate() {
    local ver_regex='tag: (v([^,\)]+)|([0-9]+(\.[0-9]+)*))'
    local tag_lines=`git log --oneline --decorate  |  fgrep "tag: "`

    if [[ $tag_lines =~ $ver_regex ]]; then
        if [[ ${BASH_REMATCH[1]:0:1} = "v" ]]; then
            version_tag=${BASH_REMATCH[1]}
            version_candidate=${BASH_REMATCH[2]}
        else
            version_tag=${BASH_REMATCH[3]}
            version_candidate=${BASH_REMATCH[3]}
        fi
    else
        version_tag=""
        version_candidate="0.0.0"
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
        if [[ $format == "docker" ]]; then
          vsn="${version_prefix}${version_candidate}-build.${commit_count}-${ref}"
        else
          vsn="${version_prefix}${version_candidate}-build.${commit_count}+${ref}"
        fi
    fi
}


function usage() {
  echo "usage: vertagen [[-p <version_prefix>] | [-f docker] ] | [-h]]"
}

while [ "$1" != "" ]; do
    case $1 in
        -p | --prefix )             shift
                                version_prefix=$1
                                ;;
        -f | --format )         shift
                                format=$1
                                ;;
        -h | --help )           usage
                                exit
                                ;;
        * )                     usage
                                exit 1
    esac
    shift
done

getVersionCandidate
getCommitCount
buildVersion

echo $vsn
