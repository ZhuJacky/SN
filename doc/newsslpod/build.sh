#!/bin/sh

set -e

# ensure go
if ! which "go" >/dev/null 2>&1; then
  echo "Not found command: go, Please install."
  exit 1
fi

# ensure program path
if [ $PWD != "$GOPATH/src/mysslee_qcloud" ]; then
  echo "Please clone into $GOPATH/src/mysslee_qcloud"
  exit 1
fi

# compile file
cd app/backend \
  && go build -o ../../deploy/services/backend \
  && cd ../..

cd app/checker \
  && go build -o ../../deploy/services/checker \
  && cd ../..

cd app/notifier \
  && go build -o ../../deploy/services/notifier \
  && cd ../..
