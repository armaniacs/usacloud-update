# 開発ワークフロー

## 基本開発サイクル

### 1. セットアップ

```bash
# 依存関係の確認
go mod tidy

# 初回ビルド
make build
```

### 2. 開発 → テスト → 修正の反復

```bash
# コード修正
# ↓
# 形式チェックとビルド
make build  # tidy, fmt も含む

# テスト実行
make test

# 必要に応じてゴールデンファイル更新
make golden

# サンプル実行で手動確認
make verify-sample
```

### 3. 品質チェック

```bash
# 静的解析
make vet

# 全体チェック
make tidy fmt vet test
```

## 新機能開発のワークフロー

### ステップ1: ルール設計

1. **要件の明確化**
   - 変換対象の usacloud コマンド特定
   - v0 → v1.1 の変更点調査
   - 公式ドキュメントのURL収集

2. **正規表現パターンの検討**
   - マッチング条件の定義
   - エッジケースの洗い出し
   - 置換ロジックの設計

### ステップ2: 実装

1. **`ruledefs.go` へのルール追加**:
   ```go
   rules = append(rules, mk(
       "new-rule-name",
       `正規表現パターン`,
       func(m []string) string {
           // 置換ロジック
           return replacementString
       },
       "変更理由の説明",
       "https://公式ドキュメントURL",
   ))
   ```

2. **テストデータの準備**:
   - `testdata/sample_v0_v1_mixed.sh` に新パターン追加
   - または新しいテストファイルの作成

### ステップ3: テストと検証

1. **ゴールデンファイルの更新**:
   ```bash
   make golden
   ```

2. **変換結果の確認**:
   ```bash
   # 統計出力付きで実行
   make run
   
   # 期待値との比較
   make verify-sample
   ```

3. **回帰テストの実行**:
   ```bash
   make test
   ```

### ステップ4: リファクタリング

```bash
# コードの整理
make fmt

# 静的解析
make vet

# 最終テスト
make test
```

## 複雑な変換ルールの開発

### カスタム `Rule` 実装

```go
type ComplexRule struct {
    // 状態管理が必要な場合のフィールド
    context map[string]interface{}
}

func (r *ComplexRule) Name() string {
    return "complex-transformation"
}

func (r *ComplexRule) Apply(line string) (string, bool, string, string) {
    // 複雑な変換ロジック
    // - 複数行にわたる処理
    // - 文脈依存の変換
    // - 条件分岐の多い処理
    
    if complexCondition(line) {
        transformed := complexTransformation(line)
        return transformed, true, extractBefore(line), extractAfter(transformed)
    }
    
    return line, false, "", ""
}
```

### デバッグ支援

1. **ログ出力の追加**:
   ```go
   import "log"
   
   func debugRule(name, input, output string) {
       if os.Getenv("DEBUG") != "" {
           log.Printf("[%s] %s → %s", name, input, output)
       }
   }
   ```

2. **テスト時のデバッグ**:
   ```bash
   # デバッグモードでテスト実行
   DEBUG=1 go test -v ./...
   ```

## パフォーマンス最適化

### プロファイリング

```bash
# CPU プロファイル
go test -cpuprofile=cpu.prof -bench=. ./...

# メモリプロファイル  
go test -memprofile=mem.prof -bench=. ./...

# プロファイル結果の確認
go tool pprof cpu.prof
go tool pprof mem.prof
```

### ベンチマークテスト

```go
func BenchmarkEngineApply(b *testing.B) {
    engine := NewDefaultEngine()
    testLine := "usacloud server list --output-type=csv"
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        engine.Apply(testLine)
    }
}

func BenchmarkLargeFile(b *testing.B) {
    // 大容量ファイルでの性能測定
    largeInput := strings.Repeat("usacloud server list\n", 10000)
    lines := strings.Split(largeInput, "\n")
    engine := NewDefaultEngine()
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        for _, line := range lines {
            engine.Apply(line)
        }
    }
}
```

## リリース準備

### 1. 最終チェック

```bash
# 全ての品質チェックを実行
make clean build test vet

# サンプル実行で手動確認
make verify-sample
```

### 2. ドキュメント更新

- `CLAUDE.md` の更新
- `ref/` 以下のドキュメント同期
- コミットメッセージの準備

### 3. バイナリ配布準備

```bash
# クリーンビルド
make clean build

# バイナリの動作確認
./bin/sacloud-update --help
echo "usacloud server list --output-type=csv" | ./bin/sacloud-update
```

## トラブルシューティング

### よくある開発時の問題

1. **正規表現のエスケープ**
   - Go の文字列リテラルとregexpのエスケープが重複
   - 生文字列 `` `pattern` `` の使用を推奨

2. **ルールの適用順序**
   - 早い段階で適用されたルールが後続に影響
   - `DefaultRules()` での順序調整が必要

3. **テストデータの不足**
   - エッジケースの見落とし
   - 複数パターンの組み合わせテスト不足

### デバッグ戦略

1. **段階的な確認**:
   ```bash
   # 単一行での動作確認
   echo "問題のある行" | ./bin/sacloud-update --stats
   
   # 特定ルールの動作確認
   go test -run "TestSpecificRule" -v ./...
   ```

2. **ルール無効化での切り分け**:
   ```go
   // 一時的にルールをコメントアウト
   // rules = append(rules, problematicRule)
   ```

3. **中間状態の出力**:
   ```go
   // Engine.Apply() 内でのデバッグ出力
   fmt.Printf("Before: %s\n", cur)
   after, ok, _, _ := r.Apply(cur)
   if ok {
       fmt.Printf("After %s: %s\n", r.Name(), after)
   }
   ```

## 継続的な改善

### 定期的なメンテナンス

- usacloud 公式ドキュメントの変更追跡
- 新しい非推奨項目の監視
- ユーザーフィードバックの収集と対応

### 拡張性の保持

- 新しい変換パターンへの対応準備
- テストケースの充実
- パフォーマンス監視の継続