#!/bin/bash
# Compila o servidor para Linux AMD64
set -e
GOOS=linux GOARCH=amd64 go build -o godevproxy-linux-amd64

echo "Binário gerado: godevproxy-linux-amd64"
