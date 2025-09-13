#!/bin/bash
# Generated scenario 2 - Complex transformation edge cases

echo "Generated scenario test 2 - Edge cases"

# Selector transformation edge cases
usacloud server list --selector "Tags.Environment=prod AND Status=running"
usacloud disk list --selector "Size>100 OR Name~backup*"
usacloud database list --selector "Plan~premium AND Zone=tk1a"

# Output format edge cases
usacloud server list --output-type=csv
usacloud disk create --name test --output-type=tsv  
usacloud database list --output-type=json

# Zone format variations
usacloud server list --zone=tk1a
usacloud server list --zone =  tk1a
usacloud server list --zone= tk1a
usacloud server list --zone = tk1a

# Multiple deprecated commands in sequence
usacloud iso-image list --output-type csv
usacloud startup-script create --name "test"
usacloud ipv4 list --zone=tk1a

# Product alias variations
usacloud product-server list --zone=all
usacloud product-disk list --output-type csv
usacloud product-database list --selector "Name~small*"

# Complex mixed transformations
usacloud server create --selector "Zone=tk1a" --output-type=csv --zone = all
usacloud iso-image list --selector "Size>1000" --output-type=tsv --zone =tk1a

# Nested command variations
usacloud server list | grep running | wc -l
usacloud disk list --output-type csv | sort | head -10
usacloud database list --zone=all | jq '.[] | select(.status=="running")'

# Quoted arguments with transformations
usacloud server list --selector 'Tags.Environment="production"' --output-type csv
usacloud disk create --name "backup-$(date +%Y%m%d)" --output-type=tsv

# Comments mixed with commands
# This should transform
usacloud iso-image list --output-type csv
# This is a comment
usacloud server list --selector "Name=test"
# Another transformation needed
usacloud product-server list --zone = all

echo "Generated scenario 2 complete"