# macOS向けusacloud-updateビルドガイド

## 概要
macOS環境でusacloud-updateをビルドするための手順を説明します。
このガイドでは、Intel Mac・Apple Silicon Mac両方での開発環境セットアップからビルドまでを詳しく説明します。

## 前提条件
- **macOS 11 (Big Sur) 以降**
- **Go 1.24.1以上**
- **Xcode Command Line Tools**
- **Homebrew**（推奨）
- **インターネット接続** (依存関係のダウンロード用)

## 環境セットアップ

### 1. Xcode Command Line Toolsのインストール

```bash
# Command Line Toolsのインストール
xcode-select --install

# インストール確認
xcode-select -p
# 出力例: /Applications/Xcode.app/Contents/Developer
```

### 2. Homebrewのインストール

#### 新規インストール
```bash
# Homebrewのインストール
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"

# Apple Silicon Macの場合、PATHを設定
echo 'eval "$(/opt/homebrew/bin/brew shellenv)"' >> ~/.zshrc
source ~/.zshrc

# Intel Macの場合は自動的にPATHが設定されます
```

#### インストール確認
```bash
brew --version
# 出力例: Homebrew 4.1.11
```

### 3. Goのインストール

#### Homebrewを使用する方法（推奨）
```bash
# 最新版Goのインストール
brew install go

# 特定バージョンの指定（必要に応じて）
brew install go@1.24

# バージョン確認
go version
# 出力例: go version go1.24.1 darwin/arm64
```

#### 公式インストーラーを使用する方法
1. [Go公式サイト](https://golang.org/dl/)を開く
2. **macOS用パッケージ** をダウンロード
   - Intel Mac: `go1.24.1.darwin-amd64.pkg`
   - Apple Silicon Mac: `go1.24.1.darwin-arm64.pkg`
3. ダウンロードしたパッケージを実行してインストール

### 4. 開発ツールのインストール

```bash
# Git（通常はCommand Line Toolsに含まれる）
git --version

# Make（通常はCommand Line Toolsに含まれる）
make --version

# 追加の便利ツール
brew install wget curl jq
```

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

### Apple Silicon向け最適化ビルド
```bash
# Apple Silicon Macでのネイティブビルド
GOARCH=arm64 go build -o bin/usacloud-update ./cmd/usacloud-update

# Intel Macでのネイティブビルド
GOARCH=amd64 go build -o bin/usacloud-update ./cmd/usacloud-update
```

### Universal Binary作成
```bash
# Intel版とApple Silicon版を個別にビルド
GOOS=darwin GOARCH=amd64 go build -o bin/usacloud-update-amd64 ./cmd/usacloud-update
GOOS=darwin GOARCH=arm64 go build -o bin/usacloud-update-arm64 ./cmd/usacloud-update

# Universal Binaryの作成
lipo -create -output bin/usacloud-update bin/usacloud-update-amd64 bin/usacloud-update-arm64

# アーキテクチャの確認
lipo -info bin/usacloud-update
# 出力例: Architectures in the fat file: bin/usacloud-update are: x86_64 arm64
```

## 実行確認

### 動作テスト
```bash
# 実行権限の付与
chmod +x bin/usacloud-update

# バージョン確認
./bin/usacloud-update --version

# ヘルプ表示
./bin/usacloud-update --help

# サンプル実行（テストデータ使用）
./bin/usacloud-update --in testdata/sample_v0_v1_mixed.sh --out output.sh

# アーキテクチャ確認
file bin/usacloud-update
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

### Homebrewでのインストール（将来的な対応）
```bash
# 将来的には以下でインストール可能にする予定
# brew tap armaniacs/tap
# brew install usacloud-update
```

### ユーザーのGOPATHにインストール
```bash
make install
# または
go install ./cmd/usacloud-update
```

### システム全体にインストール
```bash
# /usr/local/binにインストール
sudo cp bin/usacloud-update /usr/local/bin/

# または~/binに配置（個人用）
mkdir -p ~/bin
cp bin/usacloud-update ~/bin/
echo 'export PATH=$PATH:$HOME/bin' >> ~/.zshrc
source ~/.zshrc
```

### アプリケーションバンドル作成（上級者向け）
```bash
# アプリケーションバンドルの作成
mkdir -p usacloud-update.app/Contents/MacOS
cp bin/usacloud-update usacloud-update.app/Contents/MacOS/

# Info.plistの作成
cat > usacloud-update.app/Contents/Info.plist << EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>CFBundleExecutable</key>
    <string>usacloud-update</string>
    <key>CFBundleIdentifier</key>
    <string>com.armaniacs.usacloud-update</string>
    <key>CFBundleName</key>
    <string>usacloud-update</string>
    <key>CFBundleVersion</key>
    <string>1.9.6</string>
</dict>
</plist>
EOF
```

## セキュリティとコード署名

### Gatekeeper対応
macOS Catalina以降では、未署名のバイナリに対してセキュリティ警告が表示されます。

#### 開発者向け設定
```bash
# 特定のアプリを許可（実行時の警告後）
sudo spctl --add --label "usacloud-update" bin/usacloud-update

# 一時的にGatekeeperを無効化（非推奨）
sudo spctl --master-disable
```

#### コード署名（Apple Developer Program必要）
```bash
# 開発者証明書でのコード署名
codesign -s "Developer ID Application: Your Name (TEAM_ID)" bin/usacloud-update

# 署名の確認
codesign -vv bin/usacloud-update

# 公証（Notarization）用のdmg作成
hdiutil create -volname "usacloud-update" -srcfolder bin/ -ov -format UDZO usacloud-update.dmg
```

### Xcode設定
プロジェクトをXcodeで管理する場合：

```bash
# go.workファイルの作成
go work init .
go work use .

# VS Code with Go拡張のインストール
brew install --cask visual-studio-code
code --install-extension golang.go
```

## トラブルシューティング

### よくある問題と解決方法

#### 1. 「開発元が未確認のため開けません」エラー
**症状**: アプリケーション実行時にセキュリティ警告
**解決方法**:
```bash
# Controlキーを押しながらクリックして実行
# または
sudo xattr -r -d com.apple.quarantine bin/usacloud-update

# システム環境設定でも許可可能
# システム環境設定 > セキュリティとプライバシー > 「このまま開く」
```

#### 2. 権限エラー
**症状**: `permission denied` エラー
**解決方法**:
```bash
# 実行権限の付与
chmod +x bin/usacloud-update

# 完全な権限設定
chmod 755 bin/usacloud-update
```

#### 3. Rosetta 2関連エラー（Apple Silicon Mac）
**症状**: Intel版バイナリ実行時のエラー
**解決方法**:
```bash
# Rosetta 2のインストール
sudo softwareupdate --install-rosetta

# または対話的インストール
sudo softwareupdate --install-rosetta --agree-to-license
```

#### 4. パス設定エラー
**症状**: `command not found` エラー
**解決方法**:
```bash
# PATHの確認
echo $PATH

# Go PATHの設定（~/.zshrc に追加）
echo 'export PATH=$PATH:$(go env GOPATH)/bin' >> ~/.zshrc
source ~/.zshrc
```

#### 5. Homebrew関連エラー
**症状**: `brew command not found`
**解決方法**:
```bash
# Apple Silicon Macの場合
echo 'eval "$(/opt/homebrew/bin/brew shellenv)"' >> ~/.zshrc

# Intel Macの場合
echo 'eval "$(/usr/local/bin/brew shellenv)"' >> ~/.zshrc

source ~/.zshrc
```

#### 6. CGOエラー
**症状**: `clang: command not found`
**解決方法**:
```bash
# Xcode Command Line Toolsの再インストール
sudo xcode-select --install

# CGOを無効化（可能な場合）
CGO_ENABLED=0 go build -o bin/usacloud-update ./cmd/usacloud-update
```

## パフォーマンス最適化

### ビルド最適化
```bash
# 最適化ビルド
go build -ldflags="-s -w" -o bin/usacloud-update ./cmd/usacloud-update

# 並列ビルド
GOMAXPROCS=$(sysctl -n hw.ncpu) make build

# プロファイル指向最適化（PGO）
go build -pgo=default -o bin/usacloud-update ./cmd/usacloud-update
```

### メモリ使用量最適化
```bash
# メモリ制限の設定
export GOMEMLIMIT=2GiB
go build -o bin/usacloud-update ./cmd/usacloud-update
```

## クロスコンパイル

### 他のプラットフォーム向けビルド
```bash
# Windows向け
GOOS=windows GOARCH=amd64 go build -o bin/usacloud-update.exe ./cmd/usacloud-update

# Linux向け（AMD64）
GOOS=linux GOARCH=amd64 go build -o bin/usacloud-update-linux ./cmd/usacloud-update

# Linux向け（ARM64）
GOOS=linux GOARCH=arm64 go build -o bin/usacloud-update-linux-arm64 ./cmd/usacloud-update

# FreeBSD向け
GOOS=freebsd GOARCH=amd64 go build -o bin/usacloud-update-freebsd ./cmd/usacloud-update
```

## 追加情報

### 開発環境の推奨設定

#### VS Code設定（.vscode/settings.json）
```json
{
    "go.toolsManagement.checkForUpdates": "local",
    "go.useLanguageServer": true,
    "go.formatTool": "goimports",
    "go.lintTool": "golangci-lint",
    "go.buildOnSave": "package",
    "go.vetOnSave": "package",
    "go.testOnSave": true
}
```

#### Git設定
```bash
# 改行コード設定（macOS推奨）
git config --global core.autocrlf input

# 大文字小文字を区別
git config --global core.ignorecase false
```

### パッケージマネージャー対応

#### MacPortsの場合
```bash
# MacPortsでのGoインストール
sudo port install go

# MacPortsでのGitインストール
sudo port install git
```

### デバッグ環境
```bash
# delve（Goデバッガー）のインストール
go install github.com/go-delve/delve/cmd/dlv@latest

# デバッグ付きビルド
go build -gcflags="all=-N -l" -o bin/usacloud-update ./cmd/usacloud-update

# デバッガーの起動
dlv exec bin/usacloud-update
```

### CI/CD環境（GitHub Actions例）
```yaml
# .github/workflows/build-macos.yml
name: Build macOS
on: [push, pull_request]
jobs:
  build:
    runs-on: macos-latest
    steps:
    - uses: actions/checkout@v4
    - uses: actions/setup-go@v4
      with:
        go-version: '1.24.1'
    - run: make build
    - run: make test
```

## 関連ドキュメント
- [Windows向けビルドガイド](README-Build-Windows.md)
- [Linux向けビルドガイド](README-Build-Linux.md)
- [使用方法詳細](README-Usage.md)
- [開発者向けガイド](ref/detailed-implementation-reference.md)

---

**更新日**: 2025-09-18
**対応バージョン**: usacloud-update v1.9.6
**Go要件**: 1.24.1+
**対応OS**: macOS 11 (Big Sur) 以降
**対応アーキテクチャ**: Intel (x86_64), Apple Silicon (arm64)