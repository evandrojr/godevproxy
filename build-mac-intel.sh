#!/bin/bash
# Compila o servidor para MacOS Intel (amd64)
set -e
GOOS=darwin GOARCH=amd64 go build -o socks5-server-mac-intel

echo "Binário gerado: socks5-server-mac-intel"
