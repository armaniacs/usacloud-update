#!/bin/bash
# Medium script with 1000 lines for performance testing
# Generated script to test transformation performance

echo "Starting medium performance test - 1000 lines"

# Repeat common patterns to create 1000 lines
for i in {1..100}; do
    echo "# Batch $i"
    echo "usacloud server list --output-type csv"
    echo "usacloud disk list --output-type tsv" 
    echo "usacloud database list --zone = all"
    echo "usacloud iso-image list"
    echo "usacloud startup-script list"
    echo "usacloud product-server list"
    echo "usacloud server list --selector \"Name=test-$i\""
    echo "usacloud disk create --name \"disk-$i\""
    echo "usacloud database create --name \"db-$i\""
    echo ""
done

echo "Medium performance test completed - 1000 lines processed"