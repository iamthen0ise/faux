#!/bin/bash

set -e

# Detect the architecture of the machine
ARCH=$(uname -m)

# Determine the appropriate binary to use based on the architecture
if [ "$ARCH" = "x86_64" ]; then
    BINARY="faux-amd64"
elif [ "$ARCH" = "aarch64" ]; then
    BINARY="faux-arm64"
else
    echo "Unsupported architecture: $ARCH"
    exit 1
fi


# If no arguments were passed, execute the binary
exec "./bin/$BINARY"
