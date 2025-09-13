#!/bin/bash
# Generated scenario 3 - Comprehensive transformation testing

echo "Generated scenario test 3 - Comprehensive transformations"

# All transformation types in one script

# 1. Output format transformations
usacloud server list --output-type csv
usacloud disk list --output-type tsv
usacloud database list --output-type table

# 2. Selector removals
usacloud server list --selector "Name=production-*"
usacloud disk create --selector "Zone=tk1a" --name "test-disk"
usacloud database list --selector "Status=running"

# 3. Resource name changes  
usacloud iso-image list
usacloud iso-image create --name "custom-iso"
usacloud startup-script list
usacloud startup-script create --name "init-script"
usacloud ipv4 list

# 4. Product alias transformations
usacloud product-server list
usacloud product-disk list --zone=all
usacloud product-database list --output-type csv

# 5. Zone normalization
usacloud server list --zone = all
usacloud disk list --zone= tk1a
usacloud database create --zone =is1a --name "test-db"

# 6. Summary command (discontinued)
usacloud summary
usacloud summary server
usacloud summary --help

# 7. Object storage (discontinued)
usacloud object-storage list
usacloud object-storage create --name "backup-storage"

# 8. Complex combinations
usacloud iso-image list --selector "Size>500" --output-type csv --zone = all
usacloud product-server list --selector "Name~standard*" --output-type tsv
usacloud startup-script create --name "deploy" --selector "Zone=tk1a" --output-type json

# 9. Nested in shell constructs
if usacloud server list --output-type csv | grep -q "running"; then
    echo "Servers are running"
fi

for zone in tk1a tk1b; do
    usacloud iso-image list --zone = $zone --output-type csv
done

# 10. Pipe operations
usacloud product-server list --output-type csv | head -5
usacloud startup-script list --selector "Name~deploy*" | jq '.[] | .name'

echo "Generated scenario 3 complete - all transformation types covered"