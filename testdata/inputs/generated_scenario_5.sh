#!/bin/bash
# Generated scenario 5 - Final comprehensive test

echo "Generated scenario test 5 - Final comprehensive validation"

# Real-world deployment script with all transformation types

set -e

echo "=== Phase 1: Environment Validation ==="
# Output format transformations
usacloud server list --output-type csv | grep -q "running"
usacloud disk list --output-type tsv > /tmp/disk_inventory.tsv
usacloud database list --output-type json | jq '.[] | select(.status=="running")'

echo "=== Phase 2: Legacy Resource Management ==="
# Deprecated resource handling
usacloud iso-image list --output-type csv > /tmp/iso_images.csv
usacloud startup-script list | grep -E "(deploy|init)"
usacloud ipv4 list --output-type json

echo "=== Phase 3: Product Planning ==="
# Product alias transformations
pricing=$(usacloud product-server list --output-type csv)
disk_plans=$(usacloud product-disk list --zone=all --output-type tsv)
db_options=$(usacloud product-database list --output-type json)

echo "=== Phase 4: Resource Discovery ==="
# Selector-based queries (to be transformed)
web_servers=$(usacloud server list --selector "Tags.Role=web" --output-type csv)
backup_disks=$(usacloud disk list --selector "Name~backup-*" --output-type json)
prod_dbs=$(usacloud database list --selector "Tags.Environment=production")

echo "=== Phase 5: Multi-Zone Operations ==="
# Zone normalization
for zone in tk1a tk1b is1a; do
    echo "Processing zone: $zone"
    usacloud server list --zone = $zone --output-type csv
    usacloud disk list --zone= $zone --output-type tsv
    usacloud database list --zone =$zone --output-type json
done

echo "=== Phase 6: Complex Queries ==="
# Multiple transformations per command
usacloud iso-image list --selector "Size>1000" --output-type csv --zone = all
usacloud product-server list --selector "Name~standard*" --output-type tsv --zone=tk1a
usacloud startup-script list --selector "Tags.Type=deployment" --output-type json

echo "=== Phase 7: Error-Prone Patterns ==="
# Commands that commonly have issues
usacloud summary server 2>/dev/null || echo "Summary deprecated"
usacloud object-storage list 2>/dev/null || echo "Object storage discontinued"

echo "=== Phase 8: Conditional Operations ==="
# Conditional execution with transformations
if [ "$ENVIRONMENT" = "production" ]; then
    usacloud server list --selector "Tags.Environment=production" --output-type json
    usacloud iso-image list --zone = tk1a --output-type csv
fi

echo "=== Phase 9: Batch Operations ==="
# Bulk operations
declare -a operations=(
    "usacloud product-disk list --output-type csv"
    "usacloud startup-script list --selector 'Name~batch*'"
    "usacloud ipv4 list --zone = all --output-type json"
)

for op in "${operations[@]}"; do
    eval "$op" || echo "Operation failed: $op"
done

echo "=== Phase 10: Final Verification ==="
# Final checks with mixed patterns
usacloud server list --output-type json | \
    jq '.[] | select(.status=="running")' | \
    jq 'length'

usacloud iso-image list --selector "Created>2024-01-01" --output-type csv | wc -l
usacloud product-server list --zone = all | grep -c "standard"

echo "Generated scenario 5 complete - comprehensive test finished"