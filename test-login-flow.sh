#!/bin/bash

echo "Testing the fixed login flow:"
echo "1. Should prompt for username (enter: sysop)"
echo "2. Should prompt for password (enter: password)"  
echo "3. Should show bulletins"
echo "4. Should show main menu with arrow navigation"
echo ""
echo "Starting test in 3 seconds..."
sleep 3

# Use expect if available, otherwise manual test
if command -v expect >/dev/null 2>&1; then
    expect << 'EOF'
spawn ./bbs -l
expect "Username: "
send "sysop\r"
expect "Password: "
send "password\r"
expect "Press any key to continue..."
send " "
expect eof
EOF
else
    echo "Running manual test (expect not available):"
    ./bbs -l
fi
