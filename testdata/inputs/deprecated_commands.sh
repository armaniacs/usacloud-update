#!/bin/bash
# 廃止コマンドテスト用サンプル
# 全ての廃止コマンドパターンを含むスクリプト

# === v1.0で廃止されたコマンド ===

# iso-image → cdrom
usacloud iso-image list
usacloud iso-image read 123456789
usacloud iso-image create --name "test-iso"
usacloud iso-image update 123456789 --name "updated-iso"
usacloud iso-image delete 123456789

# startup-script → note  
usacloud startup-script list
usacloud startup-script read 123456789
usacloud startup-script create --name "startup.sh" --content "#!/bin/bash\necho hello"
usacloud startup-script update 123456789 --name "updated.sh"
usacloud startup-script delete 123456789

# ipv4 → ipaddress
usacloud ipv4 list
usacloud ipv4 read 123456789
usacloud ipv4 create --prefix 28
usacloud ipv4 update 123456789 --hostname "test.example.com"
usacloud ipv4 delete 123456789

# === プロダクトエイリアスの廃止 ===

# product-disk → disk-plan
usacloud product-disk list
usacloud product-disk read disk-ssd-100gb

# product-server → server-plan  
usacloud product-server list
usacloud product-server read is1a-cpu1-1gb

# product-database → database-plan
usacloud product-database list
usacloud product-database read postgresql-s

# product-ipv4 → ipaddress-plan (存在しない場合)
usacloud product-ipv4 list

# === 完全に廃止されたコマンド ===

# summary コマンド（v1.1で廃止）
usacloud summary
usacloud summary --zone is1a
usacloud summary --output-type json

# object-storage コマンド（v1.1で廃止）
usacloud object-storage list
usacloud object-storage read bucket-name
usacloud object-storage create --name "test-bucket"
usacloud object-storage delete bucket-name

# === 廃止コマンドと非推奨オプションの組み合わせ ===

# 廃止コマンド + csv出力（二重の問題）
usacloud iso-image list --output-type csv
usacloud startup-script list --output-type tsv
usacloud ipv4 list --output-type csv

# 廃止コマンド + ゾーン指定問題
usacloud iso-image list --zone = all
usacloud startup-script list --zone = is1a

# === 部分的に廃止されたサブコマンド ===

# 一部のサブコマンドが廃止された場合の例
usacloud server deprecated-action 123456789
usacloud disk old-subcommand 123456789
usacloud database legacy-operation 123456789

# === 廃止されたオプション ===

# 古いセレクタ形式（仮想的）
usacloud server list --selector "Name=web"
usacloud disk list --selector "Zone=is1a"

# 古いフィルタ形式（仮想的）
usacloud server list --filter name=web
usacloud database list --filter type=postgresql

# === 移行対象となる複雑なケース ===

# 複数の廃止要素を含む
usacloud iso-image list --output-type csv --zone = all
usacloud startup-script read 123 --selector "Tags=production" --output-type tsv

# パイプラインでの使用
usacloud iso-image list | grep "ubuntu"
usacloud product-disk list | head -10
echo "backup" | usacloud startup-script create --name "backup.sh" --content -

# 環境変数との組み合わせ
ZONE=is1a usacloud iso-image list
OUTPUT_TYPE=csv usacloud startup-script list

# 条件分岐での使用
if usacloud iso-image read 123456789 >/dev/null 2>&1; then
    echo "ISO exists"
fi

# ループでの使用
for iso in $(usacloud iso-image list --output-type csv | tail -n +2 | cut -d, -f1); do
    usacloud iso-image read $iso
done