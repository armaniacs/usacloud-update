#!/bin/bash
# Mixed errors test file - combination of different error types

echo "Testing mixed error scenarios..."

# Valid commands
usacloud server list
usacloud disk list --output-type json

# Typos (invalid commands)
usacloud serv list
usacloud lst
usacloud Server list

# Deprecated commands
usacloud iso-image list
usacloud startup-script create
usacloud object-storage list

# Invalid subcommands
usacloud server lst
usacloud disk invalid-subcommand
usacloud database wrong-action

# Invalid options
usacloud server list --invalid-flag
usacloud disk create --wrong-option

# Malformed syntax
usacloud server --incomplete-command
usacloud 
usacloud disk

# Mixed valid and invalid in sequence
usacloud server list --output-type json
usacloud serv lst --invalid-flag
usacloud disk list
usacloud iso-image create --name test

# Complex mixed errors
usacloud product-server lst --output-type csv --zone = all
usacloud startup-script invalid-sub --selector "Name=test"

echo "Mixed errors test completed"