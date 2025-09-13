#!/bin/bash
# Generated scenario 1 - Mixed version commands for testing

echo "Generated scenario test 1"

# Mix of v0 and v1.0 commands that need transformation
usacloud server list --output-type csv
usacloud server list --selector "Name=test"
usacloud disk list --output-type tsv
usacloud database list --zone = all

# Deprecated commands that need warnings
usacloud iso-image list
usacloud startup-script list
usacloud ipv4 list

# Product alias transformations
usacloud product-server list
usacloud product-disk list
usacloud product-database list

# Typo-prone commands for suggestion testing
usacloud serv list
usacloud lst
usacloud Server list

# Complex transformations
usacloud server create --selector "Zone=tk1a" --output-type csv
usacloud disk connect --server-selector "Name=web*" --zone = all

# Summary and object-storage (discontinued)
usacloud summary server
usacloud object-storage list

# Mixed correct and incorrect in same script
usacloud server list
usacloud disk lst  
usacloud database list --output-type json
usacloud iso-image create --name "test"

echo "Generated scenario 1 complete"