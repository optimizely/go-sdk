#!/bin/bash
# This script fetches Full stack compatibility suite and copies all the feature files into root/tests/integration/features directory
# and copies all the datafiles into fsc-datafiles folder.
set -e
FSC_PATH=tmp/fsc-repo
rm -rf $FSC_PATH
mkdir -p $FSC_PATH
 
pushd $FSC_PATH && git init && git fetch --depth=1 https://$CI_USER_TOKEN@github.com/optimizely/fullstack-sdk-compatibility-suite ${FSC_BRANCH:-master} && git checkout FETCH_HEAD && popd
mkdir -p ./tests/integration/features
cp -r ./$FSC_PATH/features/* ./tests/integration/features
mkdir -p ./fsc-datafiles
cp -r ./$FSC_PATH/features/support/datafiles/*.json ./fsc-datafiles
export DATAFILES_DIR=$TRAVIS_BUILD_DIR/fsc-datafiles
echo "Ready for testing."
