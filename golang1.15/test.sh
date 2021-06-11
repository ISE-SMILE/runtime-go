#!/usr/bin/env bash
set -e
set -x

BASE_DIR="$(dirname $0)"
cd $BASE_DIR

ID=${1:?target dir}
COMPILE=$2

BUILD_DIR=/tmp/action_build

#move $ID to build folder
rm -Rvf $BUILD_DIR >/dev/null
mkdir -p $BUILD_DIR
cp -R $BASE_DIR/tests/$ID/* $BUILD_DIR

#compile and zip or just zip
cd $BUILD_DIR
if [ -z "$COMPILE" ]
then
  zip -j main.zip $BUILD_DIR/*
else
  CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o exec .
  echo "openwhisk/action-golang-v1.15" >> exec.env
  zip main.zip exec exec.env
  rm exec exec.env
fi

#generate payload.json
echo -n "{\"value\":{\"code\":\"" > payload.json
base64 -w 0 main.zip >> payload.json
echo -n "\",\"Binary\":true" >> payload.json

#for non-compiled code set the Main function
if [ -z "$COMPILE" ]
then
  echo -n ",\"Main\":\"Main\"" >> payload.json
fi
echo -n "}}" >> payload.json

#send init command using payload.json
curl -i -X POST -H "Content-Type: application/json" -d @payload.json http://localhost:8080/init
sleep 2

#smoke-test the function
echo -n "{\"msg\":\"hello\"}" > msg.json
curl -i -X POST -H "Content-Type: application/json" -d @msg.json http://localhost:8080/run


cd $BUILD_DIR