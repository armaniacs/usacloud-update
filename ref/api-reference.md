# API リファレンス

## コア型定義

### `Change` 構造体

```go
type Change struct {
    RuleName string  // 適用されたルール名
    Before   string  // 変更前のテキストフラグメント
    After    string  // 変更後のテキストフラグメント
}
```

**用途**: 個別の変換記録を表現。統計出力やデバッグに使用。

### `Result` 構造体

```go
type Result struct {
    Line    string    // 変換後の行全体
    Changed bool      // 変更が発生したかのフラグ
    Changes []Change  // 適用された全変更のリスト
}
```

**用途**: 単一行の変換結果と変更履歴を包含。

### `Rule` インターフェース

```go
type Rule interface {
    Name() string
    Apply(line string) (newLine string, changed bool, beforeFrag string, afterFrag string)
}
```

**メソッド仕様**:
- `Name()`: ルールの識別名を返す
- `Apply()`: 変換処理の実行
  - `newLine`: 変換後の行全体
  - `changed`: 変更有無のブール値
  - `beforeFrag`: 変更前のテキスト片
  - `afterFrag`: 変更後のテキスト片

### `Engine` 構造体

```go
type Engine struct {
    rules []Rule
}
```

**メソッド**:
- `Apply(line string) Result`: メイン変換処理

## 公開関数

### `NewDefaultEngine() *Engine`

**戻り値**: デフォルトルールセットで初期化されたエンジン

**用途**: 標準的な v0→v1.1 変換エンジンの作成

### `GeneratedHeader() string`

**戻り値**: 生成ファイル用の標準ヘッダー文字列

**内容**: `"# Updated for usacloud v1.1 by sacloud-update — DO NOT EDIT ABOVE THIS LINE"`

### `DefaultRules() []Rule`

**戻り値**: 全ての標準変換ルールのスライス

**ルール構成** (適用順):
1. `output-type-csv-tsv`
2. `selector-to-arg`
3. `iso-image-to-cdrom`
4. `startup-script-to-note`
5. `ipv4-to-ipaddress`
6. `product-alias-*` (3つのエイリアス)
7. `summary-removed`
8. `object-storage-removed-*` (2つのエイリアス)
9. `zone-all-normalize`

## 内部実装

### `simpleRule` 構造体

```go
type simpleRule struct {
    name   string
    re     *regexp.Regexp
    repl   func([]string) string
    reason string
    url    string
}
```

**フィールド**:
- `name`: ルール識別名
- `re`: コンパイル済み正規表現
- `repl`: マッチグループから置換文字列を生成する関数
- `reason`: 変更理由（コメント用）
- `url`: 参考URL（コメント用）

### ヘルパー関数

#### `mk(name, pattern string, repl func([]string) string, reason, url string) Rule`

**引数**:
- `name`: ルール名
- `pattern`: 正規表現パターン文字列
- `repl`: 置換関数 (マッチ配列 → 置換文字列)
- `reason`: 変更理由
- `url`: 参考ドキュメントURL

**戻り値**: `simpleRule` を `Rule` インターフェースとして返却

**用途**: ルール定義の簡潔な記述を提供

## 変換処理のフロー

### `Engine.Apply()` の内部処理

1. **前処理**: 空行・コメント行のスキップ判定
2. **ルール適用**: 全ルールの順次実行
3. **変更追跡**: `Change` レコードの蓄積
4. **結果構築**: `Result` オブジェクトの組み立て

### `simpleRule.Apply()` の内部処理  

1. **パターンマッチ**: 正規表現による行の検査
2. **置換処理**: `repl` 関数による変換文字列生成
3. **コメント付与**: 変更理由とURLの自動追加
4. **フラグメント抽出**: 変更箇所の before/after 情報生成

## 使用例

### カスタムエンジンの作成

```go
// カスタムルールの定義
customRule := mk(
    "my-custom-rule",
    `old-pattern`,
    func(m []string) string {
        return strings.Replace(m[0], "old", "new", -1)
    },
    "カスタム変換の説明",
    "https://example.com/docs",
)

// カスタムエンジンの構築
engine := &Engine{
    rules: []Rule{customRule},
}

// 変換の実行
result := engine.Apply("old-pattern を含む行")
```

### 変換結果の活用

```go
result := engine.Apply(inputLine)

if result.Changed {
    fmt.Printf("変更された行: %s\n", result.Line)
    for _, change := range result.Changes {
        fmt.Printf("ルール: %s, %s → %s\n", 
            change.RuleName, change.Before, change.After)
    }
} else {
    fmt.Printf("変更なし: %s\n", result.Line)
}
```

## 拡張ガイド

### 新しい `Rule` 実装

```go
type MyCustomRule struct {
    // カスタムフィールド
}

func (r *MyCustomRule) Name() string {
    return "my-custom-rule"
}

func (r *MyCustomRule) Apply(line string) (string, bool, string, string) {
    // カスタムロジック
    if 条件 {
        return 変換後行, true, 変更前片, 変更後片
    }
    return line, false, "", ""
}
```

### 複雑な変換パターン

- 複数行にわたる変換
- 文脈依存の処理
- 状態を持つ変換ロジック

これらは `Rule` インターフェースのカスタム実装で対応可能です。