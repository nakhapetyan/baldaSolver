#!/bin/bash
cd "$(dirname "$0")"

pkill -9 -f "baldaServer"
nohup /usr/lib/go-1.9/bin/go run main.go -http 9000 baldaServer > errors.log &
ps auxf | grep baldaServer | grep -v grep