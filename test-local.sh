#!/bin/bash

echo "Testing Searchlight BBS Local Mode"
echo "=================================="
echo ""
echo "Available test accounts:"
echo "  Username: sysop    Password: password"
echo "  Username: test     Password: test"
echo ""
echo "Navigation:"
echo "  ↑↓ Arrow keys to navigate"
echo "  Enter to select"
echo "  Q to go back (on submenus)"
echo "  G to logout"
echo ""
echo "Starting BBS in 3 seconds..."
sleep 3

./bbs -l
