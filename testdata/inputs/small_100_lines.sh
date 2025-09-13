#!/bin/bash
# Small script with 100 lines for performance testing

# Server commands
usacloud server list
usacloud server list --output-type csv
usacloud server create
usacloud server delete
usacloud server read
usacloud server update
usacloud server power-on
usacloud server power-off
usacloud server reset
usacloud server shutdown

# Disk commands  
usacloud disk list
usacloud disk list --output-type tsv
usacloud disk create
usacloud disk delete
usacloud disk read
usacloud disk update
usacloud disk connect
usacloud disk disconnect
usacloud disk snapshot
usacloud disk restore

# Database commands
usacloud database list
usacloud database create
usacloud database delete
usacloud database read
usacloud database update
usacloud database backup
usacloud database restore
usacloud database clone
usacloud database monitor
usacloud database logs

# Load balancer commands
usacloud loadbalancer list
usacloud loadbalancer create
usacloud loadbalancer delete
usacloud loadbalancer read
usacloud loadbalancer update
usacloud loadbalancer monitor
usacloud loadbalancer status
usacloud loadbalancer vip-add
usacloud loadbalancer vip-delete
usacloud loadbalancer server-add

# Archive commands
usacloud archive list
usacloud archive create
usacloud archive delete
usacloud archive read
usacloud archive update
usacloud archive download
usacloud archive upload
usacloud archive share
usacloud archive unshare
usacloud archive ftp-open

# ISO image commands (deprecated)
usacloud iso-image list
usacloud iso-image create
usacloud iso-image delete
usacloud iso-image read
usacloud iso-image update

# Switch commands
usacloud switch list
usacloud switch create
usacloud switch delete
usacloud switch read
usacloud switch update
usacloud switch connect
usacloud switch disconnect
usacloud switch bridge-info
usacloud switch monitor
usacloud switch status

# Router commands
usacloud internet list
usacloud internet create
usacloud internet delete
usacloud internet read
usacloud internet update
usacloud internet monitor
usacloud internet status
usacloud internet enable-ipv6
usacloud internet disable-ipv6
usacloud internet update-bandwidth

# Bridge commands
usacloud bridge list
usacloud bridge create
usacloud bridge delete
usacloud bridge read
usacloud bridge update
usacloud bridge info
usacloud bridge monitor
usacloud bridge status
usacloud bridge connect
usacloud bridge disconnect

# Note commands (formerly startup-script)
usacloud note list
usacloud note create
usacloud note delete
usacloud note read
usacloud note update
usacloud note info
usacloud note download
usacloud note upload
usacloud note share
usacloud note unshare

# SSH key commands
usacloud ssh-key list
usacloud ssh-key create
usacloud ssh-key delete
usacloud ssh-key read
usacloud ssh-key update
usacloud ssh-key generate

# Icon commands
usacloud icon list
usacloud icon create
usacloud icon delete
usacloud icon read
usacloud icon update

# Summary command (deprecated)
usacloud summary

echo "Completed 100 line performance test"