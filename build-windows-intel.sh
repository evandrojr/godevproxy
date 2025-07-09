#!/bin/bash
# Compila o servidor para MacOS Intel (amd64)
set -e
GOOS=windows GOARCH=amd64 go build -o godevproxy-windows-intel.exe

echo "Binário gerado: godevproxy-windows-intel"
