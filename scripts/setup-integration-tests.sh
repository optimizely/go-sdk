#!/bin/bash

. ~/.nvm/nvm.sh
mkdir -p $FSC_PATH
pushd $FSC_PATH && git init && git fetch --depth=1 https://$CI_USER_TOKEN@github.com/optimizely/fullstack-sdk-compatibility-suite ${FSC_BRANCH:-master} && git checkout FETCH_HEAD
ln -s features/support/datafiles/ public 
pushd services/datafile && nvm install && nvm use && npm install && popd
node services/datafile/ &> /dev/null &
popd
