#!/bin/bash

source ./vars.sh

rm -rf $GOPATH/src/softwareupgrade/vendor/github.com
rm -rf $GOPATH/src/softwareupgrade/vendor/golang.org
rm -rf $GOPATH/pkg
rm -rf $GOPATH/bin

