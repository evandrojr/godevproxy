#!/bin/bash
# Compila o servidor para MacOS Intel (amd64)
set -e
GOOS=darwin GOARCH=amd64 go build -o godevproxy-mac-intel

echo "Binário gerado: godevproxy-mac-intel"
