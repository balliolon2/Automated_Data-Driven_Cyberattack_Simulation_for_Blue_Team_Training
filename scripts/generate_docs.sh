#!/usr/bin/env bash
set -e

echo "Checking for swag CLI..."
if ! command -v swag &> /dev/null; then
    GOPATH=$(go env GOPATH)
    if [ -z "$GOPATH" ]; then
        GOPATH="$HOME/go"
    fi
    if [ ! -f "$GOPATH/bin/swag" ]; then
        echo "Installing swaggo/swag CLI..."
        go install github.com/swaggo/swag/cmd/swag@v1.16.4
    fi
    SWAG="$GOPATH/bin/swag"
else
    SWAG="swag"
fi

echo "Generating Swagger/OpenAPI documentation..."
$SWAG init -g main.go --parseDependency --parseInternal

echo "Documentation generated successfully in docs/ (docs.go, swagger.json, swagger.yaml)"
