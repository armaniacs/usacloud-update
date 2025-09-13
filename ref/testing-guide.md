# テストガイド

## テスト戦略

usacloud-update は多層的なテスト戦略を採用し、コードの品質と信頼性を保証します：

### 主要テストアプローチ
- **ゴールデンファイルテスト**: 変換結果の一貫性を保証
- **ユニットテスト**: 各コンポーネントの個別機能を検証（**56.1%** のテストカバレッジ達成）
- **エッジケーステスト**: 並行処理、エラー条件、境界値の検証
- **BDDテスト**: Godog を使用した行動駆動開発テスト
- **統合テスト**: コンポーネント間の相互作用を検証

### テスト実装統計
- **8つの包括的テストファイル**: 5,175+行のテストコードで全機能を網羅
- **全ユニットテスト通過**: コンパイルエラー・テスト失敗の完全解決
- **BDD完全実装**: 7つのプレースホルダ関数を完全実装

## テストファイル構成

### ゴールデンファイル
```
testdata/
├── sample_v0_v1_mixed.sh           # 入力サンプル（v0とv1が混在）
├── expected_v1_1.sh                # 期待出力（v1.1統一後）
├── mixed_with_non_usacloud.sh      # 非usacloudコマンドを含む混在サンプル
└── expected_mixed_non_usacloud.sh  # 混在サンプルの期待出力
```

### 包括的テストファイル（新規作成）
```
cmd/usacloud-update/
└── main_test.go                    # CLI エントリーポイントのテスト

internal/
├── config/
│   └── env_test.go                 # 環境設定のテスト
├── transform/
│   ├── rules_test.go               # ルールシステムのテスト 
│   ├── ruledefs_test.go           # 全変換ルールの詳細テスト
│   └── engine_edge_test.go        # 変換エンジンのエッジケース
├── sandbox/
│   └── executor_edge_test.go      # サンドボックス実行のエッジケース
└── tui/
    ├── app_test.go                # TUI アプリケーションのテスト
    └── app_edge_test.go           # TUI エッジケーステスト
```

## 主要テストケース

### `TestGolden_SampleMixed`

**目的**: サンプル入力の変換結果が期待出力と一致することを検証

**処理フロー**:
1. `sample_v0_v1_mixed.sh` を変換エンジンで処理
2. 変換結果と `expected_v1_1.sh` を比較
3. 不一致時は詳細な diff を表示

**実行方法**:
```bash
# 通常のテスト実行
make test
# または
go test ./...

# より詳細な出力
go test -v ./...
```

## ゴールデンファイル更新

### 自動更新

ルール変更後、期待出力を最新の変換結果で更新:

```bash
# Makefile 経由（推奨）
make golden

# Go コマンド直接実行
go test -run Golden -update ./...
```

## BDD テスト（v1.9.0新機能） ✨

### 概要

v1.9.0で追加されたサンドボックス機能の**行動駆動開発（BDD）テスト**。
[Godog フレームワーク](https://github.com/cucumber/godog)を使用してGherkin記法による仕様駆動テストを実装。

### 実行方法

```bash
# BDD テストの実行
make bdd

# または直接実行
go test ./internal/bdd -godog.format=pretty -godog.paths=features
```

### テスト範囲

#### **完全実装済みの7つのBDDステップ関数**:

1. **`tuiInterfaceIsDisplayed()`**
   - TUI インタラクティブモードの動作検証
   - 変換スクリプトの表示確認

2. **`listOfConvertedCommandsIsDisplayed()`**  
   - 変換コマンド一覧の表示検証
   - 変換エビデンスの存在確認

3. **`followingOptionsAreProvidedForEachCommand()`**
   - TUI操作オプションの提供確認
   - コマンド実行オプションの検証

4. **`executionResultIsDisplayedInTUI()`**
   - TUI内での実行結果表示検証
   - コマンド実行情報の確認

5. **`conversionOnlyOptionIsPresented()`**
   - 変換のみオプションの提供確認
   - 接続失敗時のフォールバック検証

6. **`environmentVariableSetupIsGuided()`**
   - 認証設定のガイダンス検証
   - エラー時の適切な案内確認

7. **`envSampleFileIsReferenced()`**
   - 設定ファイルへの参照確認
   - 設定ガイダンスの検証

### テスト仕様ファイル

```
features/sandbox.feature  # Gherkin記法によるBDD仕様定義
internal/bdd/steps.go     # BDDステップ関数の実装
internal/bdd/main_test.go # BDDテストのエントリーポイント
```

### BDD vs ゴールデンファイルテスト

| テスト手法 | 対象範囲 | 目的 |
|-----------|---------|------|
| **ゴールデンファイル** | コア変換ロジック | 変換結果の一貫性保証 |
| **BDD テスト** | サンドボックス・TUI機能 | ユーザー体験の動作検証 |

### 品質保証状況

- ✅ **全BDDステップ実装完了**: 7つの関数全てに実際の検証ロジックを実装
- ✅ **包括的シナリオカバレッジ**: インタラクティブTUI、実行結果表示、設定支援を網羅
- ✅ **自動化されたサンドボックステスト**: 手動テストを自動化し品質向上に貢献

### 更新フロー

1. ルール変更の実装
2. `make golden` で期待出力を更新
3. 変更内容の確認とレビュー
4. `make test` で検証

## テスト実装詳細

### `applyFile` ヘルパー関数

```go
func applyFile(t *testing.T, inPath string) string
```

**機能**:
- 指定ファイルを行単位で読み込み
- 変換エンジンを適用
- 生成ヘッダー付きの完全な出力を返却
- 大容量ファイル対応（1MB バッファ）

**エラーハンドリング**:
- ファイル読み込みエラー
- スキャンエラー  
- メモリ不足エラー

### テストデータ特徴

#### 入力サンプル (`sample_v0_v1_mixed.sh`)

```bash
#!/usr/bin/env bash
set -euo pipefail

# v0風: csv/tsv
usacloud server list --output-type=csv

# v0風: selector
usacloud disk read --selector name=mydisk
usacloud server delete --selector tag=to-be-removed

# v0風: リソース名
usacloud iso-image list
usacloud startup-script list
usacloud ipv4 read --zone tk1a --ipaddress 203.0.113.10

# v0: product-*
usacloud product-disk list

# 廃止コマンド
usacloud summary

# オブジェクトストレージ（非対応）
usacloud object-storage list
usacloud ojs put file.txt

# ゾーン指定の正規化
usacloud server list --zone = all
```

#### 期待出力の特徴

- **生成ヘッダー**: 自動付与される警告コメント
- **説明コメント**: 各変更に対する詳細な説明とURL
- **コメントアウト**: 廃止コマンドの安全な無効化

## テストの拡張

### 新しいテストケースの追加

1. **入力サンプルの作成**:
   ```bash
   # testdata/new_sample.sh
   #!/usr/bin/env bash
   # 新しいテストパターン
   usacloud new-old-command --deprecated-option
   ```

2. **テスト関数の追加**:
   ```go
   func TestGolden_NewSample(t *testing.T) {
       inPath := "../../testdata/new_sample.sh"
       wantPath := "../../testdata/expected_new.sh"
       
       got := applyFile(t, inPath)
       
       if *update {
           if err := os.WriteFile(wantPath, []byte(got), 0o644); err != nil {
               t.Fatalf("update golden %s: %v", wantPath, err)
           }
           return
       }
       
       wantBytes, err := os.ReadFile(wantPath)
       if err != nil {
           t.Fatalf("open expected %s: %v", wantPath, err)
       }
       want := string(wantBytes)
       
       if got != want {
           t.Errorf("golden mismatch.\n--- want ---\n%s\n--- got ---\n%s", want, got)
       }
   }
   ```

3. **期待出力の生成**:
   ```bash
   go test -run TestGolden_NewSample -update ./...
   ```

### 個別ルールのユニットテスト

```go
func TestSpecificRule(t *testing.T) {
    rule := mk(
        "test-rule",
        `pattern`,
        func(m []string) string { return "replacement" },
        "test reason",
        "https://example.com",
    )
    
    testCases := []struct {
        input    string
        expected string
        changed  bool
    }{
        {"matching pattern", "replacement # usacloud-update: test reason (https://example.com)", true},
        {"no match", "no match", false},
    }
    
    for _, tc := range testCases {
        result, changed, _, _ := rule.Apply(tc.input)
        if result != tc.expected || changed != tc.changed {
            t.Errorf("input: %s, got: %s (changed: %v), want: %s (changed: %v)",
                tc.input, result, changed, tc.expected, tc.changed)
        }
    }
}
```

## テスト実行パターン

### 開発中のテスト

```bash
# 特定のテストのみ実行
go test -run TestGolden_SampleMixed ./...

# 詳細出力付き
go test -v ./...

# カバレッジ確認
go test -cover ./...
```

### CI/CD での実行

```bash
# Makefile での包括実行
make test

# 形式チェック込み
make tidy fmt vet test
```

### デバッグ用の実行

```bash
# ゴールデンファイル更新とテスト実行を組み合わせ
make golden && make test

# サンプル実行と期待値比較
make verify-sample
```

## トラブルシューティング

### よくある問題

1. **改行コードの不一致**
   - Unix (LF) と Windows (CRLF) の違い
   - Git の `autocrlf` 設定を確認

2. **大きなファイルでのメモリエラー**
   - バッファサイズの調整 (現在 1MB)
   - ファイル分割を検討

3. **ゴールデンファイルの差分**
   - ルール変更による意図的な差分
   - `make golden` で期待値を更新

### デバッグ手法

1. **統計出力の活用**:
   ```bash
   ./bin/usacloud-update --in testdata/sample_v0_v1_mixed.sh --stats
   ```

2. **中間結果の確認**:
   ```go
   // テスト内で個別の変換を検証
   result := engine.Apply("問題のある行")
   t.Logf("Result: %+v", result)
   ```

3. **ルール単体での確認**:
   ```go
   // 特定のルールのみテスト
   rule := DefaultRules()[0]  // 最初のルール
   newLine, changed, before, after := rule.Apply("テスト行")
   ```