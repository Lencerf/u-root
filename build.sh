set -x
set -e

go build

./u-root --go-build-tags=goshliner -defaultsh=gosh \
    -files $(which hexdump):bin/hexdump \
    -files $(which lspci):bin/lspci \
    -files $(which ping):bin/ping \
    -files $(which bash):bin/bash \
    -files $(which iperf3):bin/iperf3 \
    -files $(which netclient):bin/netclient \
    -files $(which netserver):bin/netserver \
    -files bash_history.txt:root/.bash_history \
    -o $HOME/data/initramfs.linux_amd64.cpio \
    core ./cmds/exp/modprobe
