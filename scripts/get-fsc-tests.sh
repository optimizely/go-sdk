#!/bin/bash
# This script fetches Full stack compatibility suite and copy feature files in to root/features directory
# and copy datafiles from Full stack compatibility suite in to fsc-datafiles folder
set -e
FSC_PATH=tmp/fsc-repo
rm -rf $FSC_PATH
mkdir -p $FSC_PATH
 
pushd $FSC_PATH && git init && git fetch --depth=1 https://$CI_USER_TOKEN@github.com/optimizely/fullstack-sdk-compatibility-suite ${FSC_BRANCH:-master} && git checkout FETCH_HEAD && popd
mkdir -p ./features
cp -r ./$FSC_PATH/features/support/datafiles/*.json ./features
ls ./features
mkdir -p ./fsc-datafiles
cp -r ./$FSC_PATH/features/* ./fsc-datafiles
export DATAFILES_DIR=$TRAVIS_BUILD_DIR/fsc-datafiles
# TODO: Need to delete it. 
ls $DATAFILES_DIR
echo "Ready for testing."
