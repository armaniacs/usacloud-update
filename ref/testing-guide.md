# テストガイド

## テスト戦略

sacloud-update は**ゴールデンファイルテスト**を採用し、変換結果の一貫性を保証します。

## テストファイル構成

```
testdata/
├── sample_v0_v1_mixed.sh    # 入力サンプル（v0とv1が混在）
└── expected_v1_1.sh         # 期待出力（v1.1統一後）
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
        {"matching pattern", "replacement # sacloud-update: test reason (https://example.com)", true},
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
   ./bin/sacloud-update --in testdata/sample_v0_v1_mixed.sh --stats
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