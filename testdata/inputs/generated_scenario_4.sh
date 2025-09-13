#!/bin/bash
# Generated scenario 4 - Error-prone patterns and edge cases

echo "Generated scenario test 4 - Error patterns and edge cases"

# Whitespace and formatting variations
usacloud server list    --output-type   csv
usacloud disk list --output-type=tsv
usacloud database list  --zone =  all   

# Mixed quotes and escaping
usacloud server list --selector 'Name="test-server"'
usacloud disk create --name 'backup-disk' --selector "Zone=tk1a"
usacloud database list --selector "Tags.Environment='production'"

# Command fragments and incomplete commands
# These should be handled gracefully
usacloud server
usacloud disk --name test
usacloud database --selector

# Multiple transformations per line
usacloud iso-image list --selector "Size>100" --output-type csv --zone = all
usacloud product-server list --output-type tsv --selector "Name~web*" --zone= tk1a

# Comments with transformable content
# usacloud iso-image list --output-type csv (commented command)
echo "Running: usacloud startup-script list --output-type tsv"
# Zone format: usacloud server list --zone = all

# Conditional execution with transformations
[ -f /tmp/test ] && usacloud product-disk list --output-type csv
[ -n "$ZONE" ] && usacloud iso-image list --zone = $ZONE

# Variable substitution with transformations  
ZONE="tk1a"
OUTPUT="csv"
usacloud startup-script list --zone = $ZONE --output-type $OUTPUT
usacloud product-server list --output-type=${OUTPUT}

# Function definitions containing transformations
deploy_function() {
    local zone=$1
    usacloud server list --zone = $zone --output-type csv
    usacloud iso-image list --selector "Name~deploy*"
}

# Heredoc with transformations (should not be transformed)
cat << 'EOF'
usacloud iso-image list --output-type csv
usacloud product-server list --zone = all
EOF

# Array and loop constructs
zones=(tk1a tk1b is1a)
for zone in "${zones[@]}"; do
    usacloud startup-script list --zone = $zone --output-type csv
done

# Case statements with transformations
case "$ACTION" in
    list)
        usacloud product-disk list --output-type csv
        ;;
    create)
        usacloud iso-image create --name "$NAME"
        ;;
esac

echo "Generated scenario 4 complete - edge cases covered"