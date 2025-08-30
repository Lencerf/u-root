#!/bin/bash

set -x
set -e

TMP_DIR=$(mktemp -d)
trap "rm -rf '$TMP_DIR'" EXIT

convert_arch() {
    case ${1} in
        arm64 | aarch64)
            echo "aarch64"
            ;;
        amd64 | x86_64)
            echo "x86_64"
            ;;
    esac
}

convert_goarch() {
    case ${1} in
        arm64 | aarch64)
            echo "arm64"
            ;;
        amd64 | x86_64)
            echo "amd64"
            ;;
    esac
}

ARCH=${ARCH:=$(uname -m)}
GOARCH=$(convert_goarch ${ARCH})
FISHARCH=$(convert_arch ${ARCH})
unset ARCH

mkdir -p bin

if [ ! -f bin/fish_${FISHARCH} ]; then
    wget https://github.com/fish-shell/fish-shell/releases/download/4.1.2/fish-4.1.2-linux-${FISHARCH}.tar.xz \
        -O bin/fish-${FISHARCH}.tar.xz
    tar x -C bin -f bin/fish-${FISHARCH}.tar.xz
    mv bin/fish bin/fish_${FISHARCH}
fi

GOARCH="" GOOS="" go build

EXTRA_FILES=(
    -files bin/fish_${FISHARCH}:bin/fish \
    -files bash_history.txt:root/.bash_history \
)

if [ $GOARCH = $(convert_goarch $(uname -m)) ] && [ $(uname -s) = "Linux" ]; then
    EXTRA_FILES+=(
        -files $(which hexdump):bin/hexdump \
        -files $(which lspci):bin/lspci \
        -files $(which iperf3):bin/iperf3 \
    )
fi

GOARCH=${GOARCH} GOOS=linux ./u-root -defaultsh="" \
    ${EXTRA_FILES[@]} \
    -o $HOME/data/initramfs.linux_${GOARCH}.cpio \
    core ./cmds/exp/modprobe
