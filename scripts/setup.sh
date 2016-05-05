#!/bin/bash

here=$(cd "$(dirname "${BASH_SOURCE}")"; pwd -P)
. $here/_env.sh

rm -f $GOPATH/src/${REPO_PATH}
mkdir -p $GOPATH/src/${ORG_PATH}
ln -s ${PWD} $GOPATH/src/${REPO_PATH}

ln -s ~/Development/OSS/Kel/kel-go gopath/src/github.com/kelproject/kel-go

glide install
