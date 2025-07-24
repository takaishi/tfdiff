#!/bin/bash

# tfdiff Demo Script
# This script demonstrates different usage patterns of tfdiff

BINARY="../dist/tfdiff"
LEFT_DIR="basic/left"
RIGHT_DIR="basic/right"

if [ ! -f "$BINARY" ]; then
    echo "Error: tfdiff binary not found at $BINARY"
    echo "Please build the binary first: go build -o tfdiff cmd/tfdiff/main.go"
    exit 1
fi

echo "🚀 tfdiff Demo - Different Usage Patterns"
echo "========================================"
echo

echo "📁 Example modules:"
echo "  Left:  $LEFT_DIR (Production configuration)"
echo "  Right: $RIGHT_DIR (Staging configuration with updates)"
echo

# Demo 1: Basic comparison
echo "1️⃣  Basic Comparison (Default: module_calls, outputs, resources, data_sources)"
echo "Command: $BINARY $LEFT_DIR $RIGHT_DIR"
echo "----------------------------------------"
$BINARY $LEFT_DIR $RIGHT_DIR
echo

read -p "Press Enter to continue..."

# Demo 2: Resources only
echo "2️⃣  Resources Only"
echo "Command: $BINARY $LEFT_DIR $RIGHT_DIR --level resources"
echo "----------------------------------------"
$BINARY $LEFT_DIR $RIGHT_DIR --level resources
echo

read -p "Press Enter to continue..."

# Demo 3: Variables and outputs only
echo "3️⃣  Variables and Outputs Only"
echo "Command: $BINARY $LEFT_DIR $RIGHT_DIR --level variables,outputs"
echo "----------------------------------------"
$BINARY $LEFT_DIR $RIGHT_DIR --level variables,outputs
echo

read -p "Press Enter to continue..."

# Demo 4: All levels
echo "4️⃣  All Levels"
echo "Command: $BINARY $LEFT_DIR $RIGHT_DIR --level all"
echo "----------------------------------------"
$BINARY $LEFT_DIR $RIGHT_DIR --level all
echo

read -p "Press Enter to continue..."

# Demo 5: JSON output
echo "5️⃣  JSON Output"
echo "Command: $BINARY $LEFT_DIR $RIGHT_DIR --output json"
echo "----------------------------------------"
$BINARY $LEFT_DIR $RIGHT_DIR --output json | head -20
echo "... (truncated for demo)"
echo

read -p "Press Enter to continue..."

# Demo 6: Show descriptions
echo "6️⃣  Include Description Differences"
echo "Command: $BINARY $LEFT_DIR $RIGHT_DIR --ignore-descriptions=false"
echo "----------------------------------------"
$BINARY $LEFT_DIR $RIGHT_DIR --ignore-descriptions=false
echo

read -p "Press Enter to continue..."

# Demo 7: Show help
echo "7️⃣  Help Information"
echo "Command: $BINARY --help"
echo "----------------------------------------"
$BINARY --help
echo

echo "✅ Demo completed!"
echo "Try running the commands yourself to explore different options."
