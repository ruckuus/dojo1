#!/usr/bin/env bash
export PATH=$PATH:/usr/local/go/bin
export GOPATH=/root/go
go version
go get -u github.com/golang/dep/cmd/dep
cd /root/go/src/github.com/ruckuus/dojo1/ && \
    dep ensure && \
    go build && \
    sudo service lenslocked.com stop && \
    cp dojo1 /root/app/dojo1 && \
    cp -r views /root/app/ && \
    cp -r public /root/app/ && \
    service lenslocked.com restart