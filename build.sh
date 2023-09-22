#!/bin/bash

set -e

if [[ -z $1 ]]; then
    echo "provide build option"
    exit 1
fi

case $1 in
    virtual-machine)
        cd $1
        go build
        mv vm ..
        ;;
    assembler)
        cd $1
        go build
        mv asm ..
        ;;
    linker)
        cd $1
        go build
        mv ln ..
        ;;
    all)
        ./build.sh virtual-machine
        ./build.sh assembler
        ./build.sh linker
        ;;
    *)
        echo "unknown build option $1"
        exit 1
        ;;
esac
