#!/bin/bash
# Expert workflow scenario - advanced operations for experienced users

echo "Starting expert workflow..."

# Advanced server management with complex selectors
usacloud server list --selector "Tags.Environment=production" --output-type json
usacloud server power-on --selector "Name~web-*" --zone=tk1a
usacloud server update --selector "Tags.Role=database" --memory 16

# Complex disk operations with batch processing
usacloud disk list --selector "Size>=100" --output-type csv
usacloud disk snapshot --selector "Name~backup-*" 
usacloud disk create --plan ssd --size 500 --name "high-performance-disk"

# Advanced database configurations
usacloud database list --selector "Plan~premium*"
usacloud database create --plan premium --backup-rotate 30 --name "production-db"
usacloud database clone --source-id 123456 --name "staging-clone"

# Load balancer with SSL and health checks
usacloud loadbalancer create --plan standard --name "lb-production"
usacloud loadbalancer ssl-certificate-add --cert-file /path/to/cert.pem
usacloud loadbalancer health-check --path /health --interval 10

# Advanced networking with multiple zones
usacloud internet create --name "multi-zone-router" --bandwidth 1000
usacloud switch create --name "production-switch" 
usacloud bridge create --name "inter-zone-bridge"

# Automation and scripting operations
usacloud note create --name "deploy-script" --class shell
usacloud note update --id 123456 --content "$(cat deploy.sh)"

# Monitoring and alerts setup
usacloud server monitor --id 123456 --interval 60
usacloud database monitor --id 789012 --interval 300

# Bulk operations with complex transformations
for zone in tk1a tk1b is1a is1b; do
    usacloud server list --zone $zone --selector "Tags.Environment=staging"
    usacloud disk cleanup --zone $zone --older-than "30d"
done

# Archive management with versioning
usacloud archive create --name "backup-$(date +%Y%m%d-%H%M%S)"
usacloud archive share --id 456789 --password "$(openssl rand -base64 12)"

# Security and compliance operations
usacloud ssh-key rotate --name "production-key"
usacloud server security-group-add --rules "tcp:443:0.0.0.0/0"

echo "Expert workflow completed"