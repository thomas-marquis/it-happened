#!/bin/bash

set -e

failed_example=""

for example in examples/*/main.go; do
    if [ -f "$example" ]; then
        dir=$(dirname "$example")
        echo "Running example: $dir"
        echo "---"
        if ! (cd "$dir" && go run main.go); then
            failed_example="$dir"
            break
        fi
        echo ""
    fi
done

if [ -n "$failed_example" ]; then
    echo ""
    echo "ERROR: Example failed: $failed_example"
    exit 1
else
    echo "SUCCESS: All examples executed successfully"
fi