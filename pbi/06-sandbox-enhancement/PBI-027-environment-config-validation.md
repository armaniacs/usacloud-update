# PBI-027: サンドボックス環境設定バリデーション

## 概要
サンドボックス実行に必要な環境設定の事前検証機能を実装します。usacloud CLI、APIキー、ネットワーク接続などの前提条件を自動チェックし、問題があれば分かりやすいガイダンスを提供します。

## 受け入れ条件
- [ ] usacloud CLIのインストール状況を自動検証できる
- [ ] APIキーの有効性を事前確認できる
- [ ] ネットワーク接続・プロキシ設定を検証できる
- [ ] 必要な権限（ゾーンアクセス等）を確認できる
- [ ] 設定不備時に修正手順を詳細に案内できる

## 技術仕様

### 1. 環境バリデータ設計
```go
type EnvironmentValidator struct {
    checks []ValidationCheck
    config *Config
}

type ValidationCheck interface {
    Name() string
    Description() string
    Validate() *ValidationResult
    Fix() error
}

type ValidationResult struct {
    Passed      bool
    Message     string
    Severity    Severity
    FixAction   string
    HelpURL     string
}

type Severity int

const (
    SeverityInfo Severity = iota
    SeverityWarning
    SeverityError
    SeverityCritical
)
```

### 2. USACloud CLI検証
```go
type USACloudCLICheck struct {
    requiredVersion string
}

func (c *USACloudCLICheck) Validate() *ValidationResult {
    // usacloud --version を実行
    cmd := exec.Command("usacloud", "--version")
    output, err := cmd.Output()
    if err != nil {
        return &ValidationResult{
            Passed:    false,
            Message:   "usacloud CLIが見つかりません",
            Severity:  SeverityCritical,
            FixAction: "usacloud CLIをインストールしてください",
            HelpURL:   "https://docs.usacloud.jp/installation/",
        }
    }
    
    // バージョン確認
    version := strings.TrimSpace(string(output))
    if !c.isVersionCompatible(version) {
        return &ValidationResult{
            Passed:    false,
            Message:   fmt.Sprintf("usacloudのバージョンが古すぎます（現在: %s, 必要: %s以上）", version, c.requiredVersion),
            Severity:  SeverityError,
            FixAction: "usacloud CLIを最新版にアップデートしてください",
            HelpURL:   "https://docs.usacloud.jp/installation/",
        }
    }
    
    return &ValidationResult{
        Passed:   true,
        Message:  fmt.Sprintf("usacloud CLI %s が利用可能です", version),
        Severity: SeverityInfo,
    }
}
```

### 3. APIキー検証
```go
type APIKeyCheck struct {
    accessToken     string
    accessTokenSecret string
    zone           string
}

func (c *APIKeyCheck) Validate() *ValidationResult {
    if c.accessToken == "" || c.accessTokenSecret == "" {
        return &ValidationResult{
            Passed:    false,
            Message:   "APIキーが設定されていません",
            Severity:  SeverityCritical,
            FixAction: "設定ファイルにAPIキーを設定してください",
            HelpURL:   "https://docs.usacloud.jp/configuration/",
        }
    }
    
    // 簡単なAPI呼び出しでキーの有効性をテスト
    cmd := exec.Command("usacloud", "auth-status", "--zone", c.zone)
    cmd.Env = append(os.Environ(),
        fmt.Sprintf("SAKURACLOUD_ACCESS_TOKEN=%s", c.accessToken),
        fmt.Sprintf("SAKURACLOUD_ACCESS_TOKEN_SECRET=%s", c.accessTokenSecret),
    )
    
    output, err := cmd.Output()
    if err != nil {
        return &ValidationResult{
            Passed:    false,
            Message:   "APIキーが無効です",
            Severity:  SeverityError,
            FixAction: "正しいAPIキーを設定してください",
            HelpURL:   "https://docs.usacloud.jp/configuration/",
        }
    }
    
    return &ValidationResult{
        Passed:   true,
        Message:  "APIキーは有効です",
        Severity: SeverityInfo,
    }
}
```

### 4. ネットワーク接続検証
```go
type NetworkCheck struct {
    endpoints []string
    timeout   time.Duration
}

func (c *NetworkCheck) Validate() *ValidationResult {
    for _, endpoint := range c.endpoints {
        if !c.testConnection(endpoint) {
            return &ValidationResult{
                Passed:    false,
                Message:   fmt.Sprintf("%s への接続に失敗しました", endpoint),
                Severity:  SeverityError,
                FixAction: "ネットワーク接続またはプロキシ設定を確認してください",
                HelpURL:   "https://docs.usacloud.jp/troubleshooting/network/",
            }
        }
    }
    
    return &ValidationResult{
        Passed:   true,
        Message:  "ネットワーク接続は正常です",
        Severity: SeverityInfo,
    }
}

func (c *NetworkCheck) testConnection(endpoint string) bool {
    client := &http.Client{
        Timeout: c.timeout,
    }
    
    _, err := client.Get(endpoint)
    return err == nil
}
```

### 5. 統合バリデーション実行
```go
func (ev *EnvironmentValidator) RunAllChecks() []*ValidationResult {
    var results []*ValidationResult
    
    for _, check := range ev.checks {
        result := check.Validate()
        result.CheckName = check.Name()
        results = append(results, result)
    }
    
    return results
}

func (ev *EnvironmentValidator) HasCriticalErrors(results []*ValidationResult) bool {
    for _, result := range results {
        if !result.Passed && result.Severity >= SeverityError {
            return true
        }
    }
    return false
}

func (ev *EnvironmentValidator) GenerateReport(results []*ValidationResult) string {
    var report strings.Builder
    
    report.WriteString("🔍 サンドボックス環境検証結果\n")
    report.WriteString("================================\n\n")
    
    for _, result := range results {
        icon := "✅"
        if !result.Passed {
            switch result.Severity {
            case SeverityWarning:
                icon = "⚠️"
            case SeverityError:
                icon = "❌"
            case SeverityCritical:
                icon = "🚫"
            }
        }
        
        report.WriteString(fmt.Sprintf("%s %s: %s\n", icon, result.CheckName, result.Message))
        
        if !result.Passed && result.FixAction != "" {
            report.WriteString(fmt.Sprintf("   💡 対処方法: %s\n", result.FixAction))
            if result.HelpURL != "" {
                report.WriteString(fmt.Sprintf("   📖 詳細: %s\n", result.HelpURL))
            }
        }
        report.WriteString("\n")
    }
    
    return report.String()
}
```

## テスト戦略
- **モックテスト**: 外部コマンド・API呼び出しをモック化
- **環境別テスト**: 様々な環境設定での動作確認
- **エラー条件テスト**: 各種設定エラーパターンの検証
- **ユーザビリティテスト**: エラーメッセージの分かりやすさ確認

## 依存関係
- 前提PBI: なし（独立した検証機能）
- 関連PBI: PBI-024（エラーハンドリング）、PBI-028（設定管理統合）
- 既存コード: internal/config/

## 見積もり
- 開発工数: 9時間
  - バリデータフレームワーク設計・実装: 3時間
  - 各種検証チェック実装: 4時間
  - レポート生成・UI統合: 2時間

## 完了の定義
- [ ] 全ての必要な環境設定項目を検証できる
- [ ] エラー時に適切な修正案内を提供する
- [ ] レポート出力が見やすく分かりやすい
- [ ] 自動修復機能が可能な範囲で動作する
- [ ] 検証処理の性能が十分である

## 備考
- 検証は軽量で高速に実行されるよう最適化
- 将来的には設定の自動修復機能も考慮
- セキュリティ情報（APIキー等）のログ出力に注意

## 実装状況
❌ **PBI-027は未実装** (2025-09-11)

### 現在の状況
- サンドボックス環境設定の自動バリデーション機能は未実装
- usacloud CLI、APIキー、ネットワーク接続の事前検証システムなし
- 環境設定エラー時の詳細ガイダンス機能なし
- 自動修復機能なし

### 実装すべき要素
1. **EnvironmentValidator フレームワーク**
   - ValidationCheck インターフェースの実装
   - ValidationResult 構造体と重要度管理
   - 統合バリデーション実行エンジン

2. **個別検証チェック**
   - USACloudCLICheck: usacloud CLIの存在・バージョン確認
   - APIKeyCheck: APIキーの有効性検証
   - NetworkCheck: エンドポイント接続テスト
   - ZoneAccessCheck: ゾーン権限確認

3. **レポート生成システム**
   - 視覚的に分かりやすいレポート出力
   - エラー別の修正手順ガイダンス
   - ヘルプURLリンク提供

4. **CLI統合**
   - --validate フラグによる事前検証実行
   - サンドボックス実行前の自動バリデーション
   - 設定ファイル作成時の検証統合

### 次のステップ
1. `internal/validator/` パッケージの作成
2. 基本的なValidationCheckインターフェースの実装
3. usacloud CLI存在確認機能の実装
4. APIキー有効性テスト機能の追加
5. ネットワーク接続テストの実装
6. 統合レポート生成機能の作成
7. メインCLIへの統合
8. 包括的なテストケース作成

### 関連ファイル
- 実装予定: `internal/validator/environment.go`
- 実装予定: `internal/validator/checks.go` 
- 実装予定: `internal/validator/report.go`
- 統合対象: `cmd/usacloud-update/main.go`
- 設定連携: `internal/config/`

---

## 実装方針変更 (2025-09-11)

🔴 **当PBIは機能拡張のため実装を延期します**

### 延期理由
- 現在のシステム安定化を優先
- 既存機能の修復・改善が急務
- リソースをコア機能の品質向上に集中
- 環境設定バリデーションよりも既存テストの安定化が緊急

### 再検討時期
- v2.0.0安定版リリース後
- テスト安定化完了後（PBI-024〜030）
- 基幹機能の品質確保完了後
- 既存サンドボックス機能の安定化完了後

### 現在の優先度
**低優先度** - 将来のロードマップで再評価予定

### 注記
- 既存の環境設定機能は引き続き保守・改善
- 環境設定バリデーション機能の実装は延期
- 現在のサンドボックス基盤の安定化を最優先