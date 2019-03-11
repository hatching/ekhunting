Exploit Kit Hunting project
===========================

This repository contains a number of helper projects for the Exploit Kit
Hunting project. In order to install it, the following steps are required:

    sudo apt install libpcap-dev
    go get github.com/hatching/ekhunting
    go install github.com/hatching/ekhunting/cmd/realtime

Furthermore some additional patches for gopacket are required, these will be
PRd shortly such that the official mirror can be used in the future.
