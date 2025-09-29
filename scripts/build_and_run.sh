#!/bin/bash

# Build and run the MQTT subscriber program

echo "Tidying up Go modules..."
go mod tidy

echo "Building the program..."
mkdir -p ../build
go build -o ../build/main ../cmd

if [ $? -eq 0 ]; then
    echo "Build successful. Running the program..."
    cd ..
    ./build/main
else
    echo "Build failed."
    exit 1
fi
