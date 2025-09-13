#!/bin/bash
# Beginner workflow scenario - common tasks for new users

echo "Starting beginner workflow..."

# Basic server operations - what beginners typically do first
usacloud server list
usacloud server list --output-type csv
usacloud server create --name "my-first-server"
usacloud server read 123456789

# Disk management - second most common operations
usacloud disk list
usacloud disk create --name "my-disk" --size 20
usacloud disk connect --server-id 123456789

# Network setup - often confusing for beginners
usacloud switch list
usacloud switch create --name "my-switch"

# Archive operations - backup and restore
usacloud archive list
usacloud archive create --name "backup-$(date +%Y%m%d)"

# SSH key management - security basics
usacloud ssh-key list
usacloud ssh-key create --name "my-key"

# Database setup - advanced operations
usacloud database list
usacloud database create --name "my-db"

# Load balancer - for scaling
usacloud loadbalancer list

# Monitoring and management
usacloud server monitor --server-id 123456789
usacloud disk monitor --disk-id 123456789

# Cleanup operations
usacloud server shutdown --server-id 123456789
usacloud disk disconnect --server-id 123456789

echo "Beginner workflow completed"