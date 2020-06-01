#!/bin/sh -e
cd functions/source/
for d in * ; do
    n=$(echo $d| tr '[:upper:]' '[:lower:]')
    cd $d
    docker build -t $n .
    docker rm $n > /dev/null 2>&1 || true
    docker run -i --name $n $n
    mkdir -p ../../packages/$d/
    docker cp $n:/output/. ../../packages/$d/
    docker rm $n > /dev/null
    cd ../
  done
cd ../../
