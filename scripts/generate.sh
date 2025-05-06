#!/bin/bash
# Script to generate documentation for Bitspark Go packages

set -e  # Exit on error

# Check for gotree
if ! command -v gotree &> /dev/null; then
    echo "Installing gotree..."
    go install bitspark.dev/go-tree/cmd/gotree@latest
fi

# Default values
SRC_DIR="."
DOCS_DIR="./docs/json"

# Determine package name in a cross-platform way
if [[ -f "$SRC_DIR/go.mod" ]]; then
    # If go.mod exists, use the module name from there
    PACKAGE_NAME=$(grep -m 1 "^module " "$SRC_DIR/go.mod" | cut -d ' ' -f 2 | xargs basename 2>/dev/null)
    # If that fails, fall back to directory name
    if [[ -z "$PACKAGE_NAME" ]]; then
        PACKAGE_NAME=$(basename "$(cd "$SRC_DIR" && pwd)")
    fi
else
    # Otherwise just use the directory name
    PACKAGE_NAME=$(basename "$(cd "$SRC_DIR" && pwd)")
fi

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    key="$1"
    case $key in
        -src|--source)
            SRC_DIR="$2"
            shift
            shift
            ;;
        -docs-dir|--docs-dir)
            DOCS_DIR="$2"
            shift
            shift
            ;;
        -name|--package-name)
            PACKAGE_NAME="$2"
            shift
            shift
            ;;
        *)
            echo "Unknown option: $1"
            exit 1
            ;;
    esac
done

echo "Generating documentation for $PACKAGE_NAME from $SRC_DIR..."

# Create docs directory if it doesn't exist
mkdir -p "$DOCS_DIR"

# Generate the documentation JSON
gotree -src "$SRC_DIR" -json -docs-dir "$DOCS_DIR"

echo "Documentation generated at $DOCS_DIR/${PACKAGE_NAME}.json"

# Simple completion message
echo "Done! JSON documentation has been generated successfully." 