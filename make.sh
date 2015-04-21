#!/bin/sh
ScriptDir=`echo $(cd "$(dirname "$0")"; pwd)`
cd ${ScriptDir}

Workspace=`pwd`

which go > /dev/null 2>&1
if [ $? -ne 0 ]; then
    echo "you must install go."
    exit 1
fi
export GOPATH=${Workspace}:${Workspace}/Godeps/_workspace

go build
