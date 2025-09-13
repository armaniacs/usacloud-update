#!/bin/bash
# Invalid options test file for error scenarios

echo "Testing invalid options scenarios..."

# Invalid flags and options
usacloud server list --invalid-flag
usacloud disk create --wrong-option value
usacloud database list --non-existent-param

# Malformed commands
usacloud server --incomplete
usacloud disk
usacloud 

# Wrong option formats
usacloud server list --output-type=invalid-format
usacloud disk list --zone=nonexistent-zone
usacloud database create --size=-100

# Mixed valid/invalid in same command
usacloud server list --output-type json --invalid-option test

# Completely wrong syntax
usacloud server list --
usacloud disk create --name
usacloud database --invalid

echo "Invalid options test completed"