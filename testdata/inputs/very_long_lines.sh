#!/bin/bash
# Very long lines test file - for testing buffer limits and line processing

echo "Testing very long lines with usacloud commands"

# Extremely long command with many parameters and transformations
usacloud server list --output-type csv --selector "Tags.Environment=production AND Tags.Role=web-server AND Tags.Region=tokyo AND Tags.Team=engineering AND Tags.Project=ecommerce AND Tags.Owner=john.doe@example.com AND Status=running AND CPU>=2 AND Memory>=4096 AND Disk>=100" --zone = all --format table --include-deleted false --include-stopped false --sort-by name --sort-order asc --limit 1000 --offset 0

# Another very long line with deprecated commands
usacloud iso-image list --output-type tsv --selector "Name~production-* OR Name~staging-* OR Name~development-* OR Name~testing-* OR Name~backup-*" --zone = tk1a --created-after 2024-01-01 --created-before 2024-12-31 --size-min 1000 --size-max 50000 --include-public true --include-private true --sort-by created --sort-order desc

# Product command with extensive selector
usacloud product-server list --output-type json --selector "CPU>=1 AND CPU<=32 AND Memory>=1024 AND Memory<=65536 AND Disk>=20 AND Disk<=2000 AND Name~standard* OR Name~premium* OR Name~optimized*" --zone = all --include-discontinued false --category compute --subcategory virtual-machine

# Very long startup script command  
usacloud startup-script create --name "very-long-deployment-script-for-production-environment-with-comprehensive-configuration-and-monitoring-setup" --content "$(cat /very/long/path/to/deployment/script/with/many/directories/production-deploy-$(date +%Y%m%d-%H%M%S).sh)" --selector "Zone=tk1a AND Environment=production"

# Chain of commands in one line
usacloud server list --output-type csv | grep production | cut -d',' -f1,2,3 | while read server; do usacloud disk list --output-type json --selector "ServerId=$server" | jq '.[] | select(.size > 100)' >> /tmp/large-disks.json; done

echo "Very long lines test completed"