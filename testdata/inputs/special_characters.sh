#!/bin/bash
# Special characters test file - for testing unicode, symbols, and encoding

echo "Testing special characters in usacloud commands"

# Commands with Unicode characters in names and selectors
usacloud server list --selector "Name=サーバー-テスト" --output-type csv
usacloud disk create --name "データ-ディスク-日本語" --output-type json  
usacloud database list --selector "Tags.環境=本番" --zone = all

# Commands with various symbols and punctuation
usacloud server list --selector "Name~web-[0-9]+" --output-type tsv
usacloud disk list --selector "Tags.Project=e-commerce@2024" --zone= tk1a
usacloud database list --selector "Size>=100GB && Status!=stopped"

# Deprecated commands with special characters
usacloud iso-image list --selector "Name~バックアップ-*" --output-type csv
usacloud startup-script create --name "初期化スクリプト#1" 
usacloud ipv4 list --selector "Address~192.168.*"

# Product commands with symbols
usacloud product-server list --selector "Price<¥10000" --output-type json
usacloud product-disk list --selector "Type=SSD & Speed>=1000IOPS"

# Commands with escaped characters
usacloud server list --selector "Description=\"High-performance server (>= 16GB RAM)\"" --output-type csv
usacloud disk create --name "backup_$(date +%Y-%m-%d_%H:%M:%S)" --output-type json

# Mixed encoding scenarios
usacloud server list --selector "Tags.Owner=山田太郎" --zone = "東京1a" --output-type csv
usacloud database list --selector "Name~テスト-データベース-[α-ω]+" --output-type json

# Special shell characters that need careful handling
usacloud server list --selector 'Name="test|prod"' --output-type csv
usacloud disk list --selector "Size>100 && Name!~temp*" --output-type json
usacloud database list --selector "Tags.Environment in [prod,staging]" --zone = all

# Mathematical and technical symbols
usacloud server list --selector "CPU>=2 ∧ Memory>=4096MB" --output-type csv
usacloud disk list --selector "IOPS≥1000 ∨ Size≥500GB" --output-type json

# Emoji and modern Unicode (edge case)
usacloud server create --name "🚀production-server-2024" --tags "Environment=prod 🏷️"
usacloud disk list --selector "Tags.Status=✅running" --output-type json

echo "Special characters test completed ✅"