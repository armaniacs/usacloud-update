# Windows向けusacloud-updateビルドガイド

## 概要
Windows環境でusacloud-updateをビルドするための手順を説明します。
このガイドでは、Windows 10/11でのGo開発環境のセットアップからビルドまでを段階的に説明します。

## 前提条件
- **Windows 10/11** (64-bit)
- **Go 1.24.1以上**
- **Git for Windows**
- **Make** (オプション)
- **インターネット接続** (依存関係のダウンロード用)

## 環境セットアップ

### 1. Goのインストール

#### 公式インストーラーを使用する方法（推奨）
1. [Go公式サイト](https://golang.org/dl/)を開く
2. **"go1.24.1.windows-amd64.msi"** をダウンロード
3. MSIインストーラーを実行し、指示に従ってインストール
4. インストール完了後、新しいコマンドプロンプトを開く

#### 環境変数の確認
```cmd
go version
```
上記コマンドで`go version go1.24.1 windows/amd64`のような出力が表示されれば成功です。

### 2. Git for Windowsのインストール

1. [Git for Windows](https://gitforwindows.org/)をダウンロード
2. インストーラーを実行
3. 推奨設定：
   - **"Git Bash Here"** を有効化
   - **"Git from the command line and also from 3rd-party software"** を選択
   - **改行コード**: "Checkout Windows-style, commit Unix-style line endings"

### 3. Make（オプション）

#### Chocolateyを使用する方法
```powershell
# PowerShellを管理者権限で実行
Set-ExecutionPolicy Bypass -Scope Process -Force
[System.Net.ServicePointManager]::SecurityProtocol = [System.Net.ServicePointManager]::SecurityProtocol -bor 3072
iex ((New-Object System.Net.WebClient).DownloadString('https://community.chocolatey.org/install.ps1'))

# Makeのインストール
choco install make
```

#### 手動インストール
1. [GnuWin32 Make](http://gnuwin32.sourceforge.net/packages/make.htm)をダウンロード
2. インストール後、`C:\Program Files (x86)\GnuWin32\bin`をPATHに追加

## ビルド手順

### Git Bashを使用する場合（推奨）:
```bash
# リポジトリのクローン
git clone https://github.com/armaniacs/usacloud-update.git
cd usacloud-update

# 依存関係の解決
go mod tidy

# ビルド実行
go build -o bin/usacloud-update.exe ./cmd/usacloud-update
```

### PowerShellを使用する場合:
```powershell
# リポジトリのクローン
git clone https://github.com/armaniacs/usacloud-update.git
cd usacloud-update

# 依存関係の解決
go mod tidy

# ビルド実行
go build -o bin\usacloud-update.exe .\cmd\usacloud-update
```

### Makeが利用可能な場合:
```bash
# 簡単ビルド（推奨）
make build

# その他の便利コマンド
make test           # テスト実行
make vet            # 静的解析
make clean          # ビルド成果物の削除
```

## 実行確認

### 動作テスト
```cmd
# バージョン確認
.\bin\usacloud-update.exe --version

# ヘルプ表示
.\bin\usacloud-update.exe --help

# サンプル実行（テストデータ使用）
.\bin\usacloud-update.exe --in testdata\sample_v0_v1_mixed.sh --out output.sh
```

## インストール

### システム全体で利用可能にする
```powershell
# 任意のディレクトリ（例: C:\Tools\）にコピー
copy bin\usacloud-update.exe C:\Tools\

# 環境変数PATHにC:\Toolsを追加
# システムのプロパティ > 詳細設定 > 環境変数 から設定
```

### Goツールとしてインストール
```bash
go install ./cmd/usacloud-update
```

## トラブルシューティング

### よくある問題と解決方法

#### 1. 文字化けが発生する
**症状**: 日本語が正しく表示されない
**解決方法**:
- Git Bashまたは **PowerShell 7+** を使用
- コマンドプロンプトの場合: `chcp 65001` で UTF-8 に設定

#### 2. パス区切り文字エラー
**症状**: `no such file or directory` エラー
**解決方法**:
- Windows環境では `\` を使用
- Git Bashでは `/` も使用可能

#### 3. 改行コードの問題
**症状**: テストファイルで差分が発生
**解決方法**:
```bash
git config --global core.autocrlf true
```

#### 4. ファイアウォールエラー
**症状**: `go mod tidy` でダウンロードエラー
**解決方法**:
- Windows Defenderでファイアウォール例外を追加
- プロキシ環境の場合: `GOPROXY` 環境変数を設定

#### 5. 権限エラー
**症状**: 実行時に権限エラー
**解決方法**:
```powershell
# PowerShellの実行ポリシー変更
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
```

#### 6. メモリ不足エラー
**症状**: ビルド時にメモリエラー
**解決方法**:
```bash
# Goのメモリ制限を増やす
set GOMEMLIMIT=2GiB
go build -o bin/usacloud-update.exe ./cmd/usacloud-update
```

## パフォーマンス最適化

### ビルド速度向上
```powershell
# Windows Defenderの除外設定にプロジェクトフォルダを追加
# Windows セキュリティ > ウイルスと脅威の防止 > 除外 > フォルダーを追加
```

### 並列ビルド
```bash
# CPUコア数を指定（例: 4コア）
set GOMAXPROCS=4
make build
```

## 追加情報

### WSL2を使用したLinuxビルド
Windows上でLinux環境を使用してビルドすることも可能です：
```bash
# WSL2でUbuntuを起動
wsl -d Ubuntu

# Linux向けの手順に従ってビルド
# 詳細は README-Build-Linux.md を参照
```

### クロスコンパイル
```bash
# Linux向けバイナリを作成
set GOOS=linux
set GOARCH=amd64
go build -o bin/usacloud-update-linux ./cmd/usacloud-update

# macOS向けバイナリを作成
set GOOS=darwin
set GOARCH=amd64
go build -o bin/usacloud-update-macos ./cmd/usacloud-update
```

### Visual Studio Code設定
`.vscode/settings.json`:
```json
{
    "go.toolsManagement.checkForUpdates": "local",
    "go.useLanguageServer": true,
    "go.formatTool": "goimports",
    "go.lintTool": "golangci-lint"
}
```

## 関連ドキュメント
- [Linux向けビルドガイド](README-Build-Linux.md)
- [macOS向けビルドガイド](README-Build-macOS.md)
- [使用方法詳細](README-Usage.md)
- [開発者向けガイド](ref/detailed-implementation-reference.md)

---

**更新日**: 2025-09-18
**対応バージョン**: usacloud-update v1.9.6
**Go要件**: 1.24.1+