#!/usr/bin/env bash
export PATH=$PATH:/usr/local/go/bin
export GOPATH=/root/go
go version
cd /root/go/src/github.com/ruckuus/dojo1/ && \
    govendor sync && \
    go build && \
    sudo service lenslocked.com stop && \
    cp dojo1 /root/app/dojo1 && \
    cp -r views /root/app/ && \
    cp -r public /root/app/ && \
    service lenslocked.com restart