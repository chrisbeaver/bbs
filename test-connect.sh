#!/bin/bash

# Simple test script to connect to the BBS
# This script demonstrates how to connect to the SSH BBS server

echo "Searchlight BBS Test Client"
echo "=========================="
echo ""
echo "This script will connect to the BBS server running on localhost:2323"
echo ""
echo "Available test accounts:"
echo "  Username: sysop, Password: password (Full access)"
echo "  Username: test,  Password: test     (Regular user)"
echo ""
echo "To connect manually, use:"
echo "  ssh -p 2323 sysop@localhost"
echo "  ssh -p 2323 test@localhost"
echo ""
echo "Press Ctrl+C to disconnect once connected."
echo ""

# Check if the BBS server is running
if ! nc -z localhost 2323 2>/dev/null; then
    echo "Error: BBS server is not running on port 2323"
    echo "Please start the server first with: go run main.go"
    exit 1
fi

echo "BBS server is running. Connecting as 'sysop'..."
echo ""

# Connect to the BBS
ssh -p 2323 -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null sysop@localhost
