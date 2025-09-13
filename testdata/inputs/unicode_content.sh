#!/bin/bash
# Unicode content test file - comprehensive Unicode and internationalization testing

echo "Unicode content testing - 多言語対応テスト"

# Japanese (Hiragana, Katakana, Kanji)
usacloud server list --selector "Name=ウェブサーバー" --output-type csv
usacloud disk create --name "データベース用ディスク" --output-type json
usacloud database list --selector "Tags.環境=本番環境" --zone = all

# Chinese (Simplified and Traditional)
usacloud server list --selector "Name=数据库服务器" --output-type tsv
usacloud disk list --selector "Tags.項目=電子商務" --zone= tk1a
usacloud database create --name "用戶數據庫" --zone = is1a

# Korean
usacloud server list --selector "Name=웹서버" --output-type csv
usacloud disk create --name "백업디스크" --output-type json
usacloud database list --selector "Tags.환경=운영환경"

# European languages with accents
usacloud server list --selector "Name=Serveur-Européen" --output-type csv
usacloud disk create --name "Almacén-de-Datos-Español" --output-type json
usacloud database list --selector "Tags.Região=São-Paulo"

# Cyrillic script
usacloud server list --selector "Name=Сервер-База-Данных" --output-type csv
usacloud disk create --name "Диск-Резервного-Копирования" --output-type json

# Arabic (RTL script)
usacloud server list --selector "Name=خادم-قاعدة-البيانات" --output-type csv
usacloud database create --name "قاعدة-بيانات-المستخدمين"

# Mixed scripts in single command
usacloud server list --selector "Tags.Owner=田中-Smith-김철수" --output-type json
usacloud disk create --name "データ-Data-데이터-$(date +%Y%m%d)" --zone = all

# Deprecated commands with Unicode
usacloud iso-image list --selector "Name~イメージ-*" --output-type csv
usacloud startup-script create --name "初期化-스크립트-Инициализация" 
usacloud ipv4 list --selector "Description~ネットワーク-*"

# Product commands with international content
usacloud product-server list --selector "Description~高性能-高性能-고성능" --output-type json
usacloud product-disk list --selector "Type~SSD-固体硬盘-솔리드스테이트"

# Mathematical and scientific Unicode symbols
usacloud server list --selector "CPU≥2 ∧ Memory≥4096MB ∧ πr²>100" --output-type csv
usacloud disk list --selector "IOPS≥1000 ∨ 容量≥500GB ∨ 크기≥1TB" --output-type json

# Currency symbols from different countries
usacloud server list --selector "Cost<¥10000 ∨ Price<€100 ∨ Value<$150" --output-type csv
usacloud database list --selector "Budget<₩100000 ∨ Amount<₹5000"

# Emoji and pictographs (modern Unicode)  
usacloud server create --name "🌐web-server-🚀2024" --tags "Status=🟢running"
usacloud disk list --selector "Tags.Backup=💾daily" --output-type json
usacloud database create --name "📊analytics-db-📈trending"

# Zero-width and combining characters (edge cases)
usacloud server list --selector "Name=Test‌Server" --output-type csv  # Zero-width non-joiner
usacloud disk create --name "Café́-disk" --output-type json  # Combining acute accent

echo "Unicode content test completed ✅ - テスト完了 - 测试完成 - 테스트 완료"