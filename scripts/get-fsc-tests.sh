#!/bin/bash

# This script fetches Full stack compatibility suite and copies all the feature files and datafiles to the given paths.

# expects defined in environment or job settings
# CI_USER_TOKEN

# inputs:
# FSC_BRANCH - which branch to use (default: master)
# FEATURES_PATH - destination path to copy feature files (required)
# DATAFILES_PATH - destination path to copy datafiles (required)

FSC_BRANCH="${FSC_BRANCH:-master}"
FEATURES_PATH="${FEATURES_PATH:-}"
DATAFILES_PATH="${DATAFILES_PATH:-}"

usage() { echo "Usage: $0 -h [-b <string>] [-f <string>] [-d <string>]" 1>&2; }

show_example() { cat <<EOF
Example: $0 -b feature_branch -f tests/integration/features -d fsc-datafiles
EOF
}

while getopts ":b:f:d:h" o; do
  case "${o}" in
    b)
      FSC_BRANCH=${OPTARG}
      ;;
    f)
      FEATURES_PATH=${OPTARG}
      ;;
    d)
      DATAFILES_PATH=${OPTARG}
      ;;
    h)
      usage
      echo
      show_example
      exit 1
      ;;
    *)
      usage
      exit 1
      ;;
  esac

done
shift $((OPTIND-1))


if [ -z "$FEATURES_PATH" ]; then
  echo
  echo "-f is a required argument"
  echo
  show_example
  exit 1
fi

if [ -z "$DATAFILES_PATH" ]; then
  echo
  echo "-d is a required argument"
  echo
  show_example
  exit 1
fi

set -e
FSC_PATH=tmp/fsc-repo
rm -rf $FSC_PATH
mkdir -p $FSC_PATH
 
pushd $FSC_PATH && git init && git fetch --depth=1 https://$CI_USER_TOKEN@github.com/optimizely/fullstack-sdk-compatibility-suite ${FSC_BRANCH} && git checkout FETCH_HEAD && popd
mkdir -p ./${FEATURES_PATH}
cp -r ./$FSC_PATH/features/* ./${FEATURES_PATH}
mkdir -p ./${DATAFILES_PATH}
cp -r ./$FSC_PATH/features/support/datafiles/*.json ./${DATAFILES_PATH}

echo "Ready for testing."
