
#!/bin/bash

set -e
PKG=$1
NAME=$2

if [ eq $PKG "" ]
then
    echo "PKG is required"
fi

if [ eq $NAME "" ]
then
    echo "NAME is required"
fi

cmd="mockery -dir=$PKG -name=$NAME -output=./pkg/mocks -case underscore"
echo "Mocking..."
echo $cmd
eval $cmd