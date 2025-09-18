# Linux向けusacloud-updateビルドガイド

## 概要
Linux環境でusacloud-updateをビルドするための手順を説明します。
このガイドでは、主要なLinuxディストリビューション（Ubuntu、CentOS、Debian、Fedora等）でのビルド方法を詳しく説明します。

## 前提条件
- **Linux** (Ubuntu 20.04+, CentOS 7+, Debian 10+, Fedora 30+など)
- **Go 1.24.1以上**
- **Git**
- **Make**
- **gcc/clang** (CGOが必要な依存関係用)
- **インターネット接続** (依存関係のダウンロード用)

## 環境セットアップ

### 1. パッケージマネージャの更新

#### Ubuntu/Debian系
```bash
sudo apt update && sudo apt upgrade -y
```

#### CentOS/RHEL 7系
```bash
sudo yum update -y
```

#### CentOS/RHEL 8+ / Fedora
```bash
sudo dnf update -y
```

#### Arch Linux
```bash
sudo pacman -Syu
```

### 2. 必要なパッケージのインストール

#### Ubuntu/Debian系
```bash
sudo apt install -y git make build-essential curl wget
```

#### CentOS/RHEL 7系
```bash
sudo yum groupinstall -y "Development Tools"
sudo yum install -y git make gcc curl wget
```

#### CentOS/RHEL 8+ / Fedora
```bash
sudo dnf groupinstall -y "Development Tools"
sudo dnf install -y git make gcc curl wget
```

#### Arch Linux
```bash
sudo pacman -S git make gcc curl wget
```

### 3. Goのインストール

#### 公式バイナリを使用する方法（推奨）
```bash
# 古いGoの削除（既にインストールされている場合）
sudo rm -rf /usr/local/go

# 最新版のダウンロードとインストール
wget https://go.dev/dl/go1.24.1.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.24.1.linux-amd64.tar.gz

# PATHの設定
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
echo 'export GOPATH=$HOME/go' >> ~/.bashrc
echo 'export PATH=$PATH:$GOPATH/bin' >> ~/.bashrc

# 設定の反映
source ~/.bashrc

# 動作確認
go version
```

#### パッケージマネージャーを使用する方法

**Ubuntu/Debian（Snapを使用）**:
```bash
sudo snap install go --classic
```

**Fedora**:
```bash
sudo dnf install golang
```

**Arch Linux**:
```bash
sudo pacman -S go
```

**注意**: パッケージマネージャー版のGoは古い場合があります。1.24.1未満の場合は公式バイナリを使用してください。

## ビルド手順

### 基本的なビルド
```bash
# リポジトリのクローン
git clone https://github.com/armaniacs/usacloud-update.git
cd usacloud-update

# 依存関係の解決
go mod tidy

# Makeを使用したビルド（推奨）
make build

# または直接Goコマンドでビルド
go build -o bin/usacloud-update ./cmd/usacloud-update
```

### 最適化ビルド
```bash
# 静的リンクビルド（他のLinuxシステムでも動作）
CGO_ENABLED=0 go build -ldflags "-s -w" -o bin/usacloud-update ./cmd/usacloud-update

# デバッグ情報付きビルド
go build -gcflags="all=-N -l" -o bin/usacloud-update ./cmd/usacloud-update
```

## 実行確認

### 動作テスト
```bash
# バージョン確認
./bin/usacloud-update --version

# ヘルプ表示
./bin/usacloud-update --help

# サンプル実行（テストデータ使用）
./bin/usacloud-update --in testdata/sample_v0_v1_mixed.sh --out output.sh

# 実行権限の付与（必要に応じて）
chmod +x bin/usacloud-update
```

### テスト実行
```bash
# 単体テスト実行
make test

# BDDテスト実行
make bdd

# 長時間テスト実行
make test-long

# 静的解析
make vet
```

## インストール

### ユーザーのGOPATHにインストール
```bash
make install
# または
go install ./cmd/usacloud-update
```

### システム全体にインストール
```bash
# /usr/local/binにインストール（要管理者権限）
sudo cp bin/usacloud-update /usr/local/bin/

# または/usr/binにインストール
sudo cp bin/usacloud-update /usr/bin/

# 実行権限の確認
sudo chmod +x /usr/local/bin/usacloud-update
```

### Debian/Ubuntuパッケージ作成（上級者向け）
```bash
# fpm（Effing Package Management）を使用
gem install fpm

# .debパッケージ作成
fpm -s dir -t deb -n usacloud-update -v 1.9.6 \
    --description "usacloud script migration tool" \
    --url "https://github.com/armaniacs/usacloud-update" \
    --license "MIT" \
    bin/usacloud-update=/usr/local/bin/usacloud-update
```

## トラブルシューティング

### よくある問題と解決方法

#### 1. 権限エラー
**症状**: `permission denied` エラー
**解決方法**:
```bash
# 実行権限の付与
chmod +x bin/usacloud-update

# sudoでのインストール
sudo cp bin/usacloud-update /usr/local/bin/
```

#### 2. Goパスエラー
**症状**: `GOPATH not set` または `module not found`
**解決方法**:
```bash
# Go環境の確認
go env

# GOPATHの設定（必要に応じて）
export GOPATH=$HOME/go
export PATH=$PATH:$GOPATH/bin
```

#### 3. 依存関係エラー
**症状**: `build constraints exclude all Go files`
**解決方法**:
```bash
# Goのバージョン確認
go version

# モジュールキャッシュのクリア
go clean -modcache
go mod tidy
```

#### 4. ネットワークエラー
**症状**: `dial tcp: lookup proxy.golang.org` エラー
**解決方法**:
```bash
# プロキシ設定の確認
go env GOPROXY

# 直接ダウンロードの設定
export GOPROXY=direct
export GOSUMDB=off
```

#### 5. メモリ不足エラー
**症状**: ビルド時にメモリ不足
**解決方法**:
```bash
# スワップファイルの作成（一時的）
sudo fallocate -l 2G /swapfile
sudo chmod 600 /swapfile
sudo mkswap /swapfile
sudo swapon /swapfile

# ビルド後にスワップを無効化
sudo swapoff /swapfile
sudo rm /swapfile
```

#### 6. CGOエラー
**症状**: `cgo: exec clang: no such file or directory`
**解決方法**:
```bash
# Ubuntu/Debian
sudo apt install build-essential

# CentOS/RHEL/Fedora
sudo dnf install gcc

# CGOを無効化（可能な場合）
CGO_ENABLED=0 go build -o bin/usacloud-update ./cmd/usacloud-update
```

## パフォーマンス最適化

### 並列ビルド
```bash
# CPUコア数を指定
export GOMAXPROCS=$(nproc)
make build
```

### ビルドキャッシュの活用
```bash
# ビルドキャッシュの場所確認
go env GOCACHE

# キャッシュサイズの確認
du -sh $(go env GOCACHE)
```

## クロスコンパイル

### 他のプラットフォーム向けビルド
```bash
# Windows向け
GOOS=windows GOARCH=amd64 go build -o bin/usacloud-update.exe ./cmd/usacloud-update

# macOS向け
GOOS=darwin GOARCH=amd64 go build -o bin/usacloud-update-macos ./cmd/usacloud-update

# ARM64 Linux向け（Raspberry Pi等）
GOOS=linux GOARCH=arm64 go build -o bin/usacloud-update-arm64 ./cmd/usacloud-update

# 32bit Linux向け
GOOS=linux GOARCH=386 go build -o bin/usacloud-update-386 ./cmd/usacloud-update
```

## Docker環境でのビルド

### Dockerfileの例
```dockerfile
FROM golang:1.24.1-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -ldflags "-s -w" -o usacloud-update ./cmd/usacloud-update

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/usacloud-update .
CMD ["./usacloud-update"]
```

### Docker使用例
```bash
# イメージのビルド
docker build -t usacloud-update .

# コンテナでの実行
docker run --rm -v $(pwd):/workspace usacloud-update --help
```

## システム固有の注意事項

### Ubuntu/Debian
- Snapパッケージを使用する場合、`/snap/bin`がPATHに含まれていることを確認
- 古いバージョンの場合、PPA を使用して最新のGoを取得

### CentOS/RHEL
- EPEL リポジトリの有効化が必要な場合があります
- SELinux が有効な場合、適切なコンテキストの設定が必要

### Arch Linux
- AUR（Arch User Repository）を使用してGoの最新版を取得可能

## 追加情報

### 開発環境セットアップ
```bash
# VS Codeのインストール（Snap使用）
sudo snap install code --classic

# Go拡張機能の設定
code --install-extension golang.go

# delve（デバッガー）のインストール
go install github.com/go-delve/delve/cmd/dlv@latest
```

### プロファイリング
```bash
# CPUプロファイル付きビルド
go build -cpuprofile=cpu.prof -o bin/usacloud-update ./cmd/usacloud-update

# メモリプロファイル付きビルド
go build -memprofile=mem.prof -o bin/usacloud-update ./cmd/usacloud-update
```

## 関連ドキュメント
- [Windows向けビルドガイド](README-Build-Windows.md)
- [macOS向けビルドガイド](README-Build-macOS.md)
- [使用方法詳細](README-Usage.md)
- [開発者向けガイド](ref/detailed-implementation-reference.md)

---

**更新日**: 2025-09-18
**対応バージョン**: usacloud-update v1.9.6
**Go要件**: 1.24.1+
**対応ディストリビューション**: Ubuntu 20.04+, CentOS 7+, Debian 10+, Fedora 30+, Arch Linux