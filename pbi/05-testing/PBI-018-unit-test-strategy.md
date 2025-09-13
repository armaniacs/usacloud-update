# PBI-018: ユニットテスト戦略

## 概要
コマンド検証・エラーフィードバックシステム全体の品質を保証するため、各コンポーネントに対する包括的なユニットテストスイートを設計・実装する。高いテストカバレッジと保守性を両立し、継続的な品質改善を支援するテスト戦略を構築する。

## 受け入れ条件
- [ ] 全新規コンポーネントのテストカバレッジが90%以上である
- [ ] テストが高速で信頼性が高く、CI/CDパイプラインで実行可能である
- [ ] 各テストが独立して実行可能でテスト間の依存がない
- [ ] テストデータとモックが適切に設計され再利用可能である
- [ ] テストの可読性が高く、ドキュメントとしても機能する

## 技術仕様

### テスト構造とアーキテクチャ

#### 1. テストディレクトリ構造
```
internal/
├── validation/
│   ├── main_command_validator.go
│   ├── main_command_validator_test.go
│   ├── subcommand_validator.go
│   ├── subcommand_validator_test.go
│   ├── deprecated_detector.go
│   ├── deprecated_detector_test.go
│   ├── similar_command_suggester.go
│   ├── similar_command_suggester_test.go
│   ├── error_message_generator.go
│   ├── error_message_generator_test.go
│   ├── comprehensive_error_formatter.go
│   ├── comprehensive_error_formatter_test.go
│   ├── user_friendly_help_system.go
│   ├── user_friendly_help_system_test.go
│   └── testdata/
│       ├── commands/
│       │   ├── valid_commands.json
│       │   ├── invalid_commands.json
│       │   └── command_variations.json
│       ├── errors/
│       │   ├── error_scenarios.json
│       │   └── expected_messages.json
│       └── help/
│           ├── help_contexts.json
│           └── expected_responses.json
├── transform/
│   ├── integrated_engine.go
│   ├── integrated_engine_test.go
│   └── testdata/
└── config/
    ├── integrated_config.go
    ├── integrated_config_test.go
    └── testdata/
```

#### 2. テストフレームワークとユーティリティ
```go
// internal/validation/testing_utils.go
package validation

import (
    "encoding/json"
    "fmt"
    "io/ioutil"
    "path/filepath"
    "testing"
    "time"
)

// TestHelper はテストヘルパー関数群
type TestHelper struct {
    t            *testing.T
    dataDir      string
    tempDir      string
}

// NewTestHelper は新しいテストヘルパーを作成
func NewTestHelper(t *testing.T) *TestHelper {
    return &TestHelper{
        t:       t,
        dataDir: "testdata",
        tempDir: t.TempDir(),
    }
}

// LoadTestData はテストデータをJSONから読み込み
func (th *TestHelper) LoadTestData(filename string, v interface{}) {
    path := filepath.Join(th.dataDir, filename)
    data, err := ioutil.ReadFile(path)
    if err != nil {
        th.t.Fatalf("テストデータ読み込みエラー %s: %v", path, err)
    }
    
    if err := json.Unmarshal(data, v); err != nil {
        th.t.Fatalf("JSONパースエラー %s: %v", path, err)
    }
}

// AssertEqual は値の等価性をアサート
func (th *TestHelper) AssertEqual(got, want interface{}, msgAndArgs ...interface{}) {
    if got != want {
        th.t.Helper()
        msg := fmt.Sprintf("Expected: %v, Got: %v", want, got)
        if len(msgAndArgs) > 0 {
            msg = fmt.Sprintf(msgAndArgs[0].(string), msgAndArgs[1:]...) + " - " + msg
        }
        th.t.Error(msg)
    }
}

// AssertContains は文字列の包含をアサート
func (th *TestHelper) AssertContains(haystack, needle string, msgAndArgs ...interface{}) {
    if !strings.Contains(haystack, needle) {
        th.t.Helper()
        msg := fmt.Sprintf("Expected '%s' to contain '%s'", haystack, needle)
        if len(msgAndArgs) > 0 {
            msg = fmt.Sprintf(msgAndArgs[0].(string), msgAndArgs[1:]...) + " - " + msg
        }
        th.t.Error(msg)
    }
}

// BenchmarkHelper はベンチマーク用ヘルパー
type BenchmarkHelper struct {
    dataCache map[string]interface{}
}

// LoadBenchmarkData はベンチマーク用データを読み込み
func (bh *BenchmarkHelper) LoadBenchmarkData(filename string) interface{} {
    if data, exists := bh.dataCache[filename]; exists {
        return data
    }
    // データ読み込みとキャッシュ
    // ...
    return nil
}
```

### コンポーネント別テスト実装

#### 1. メインコマンド検証器テスト
```go
// internal/validation/main_command_validator_test.go
package validation

import (
    "testing"
)

// TestMainCommandValidator_ValidCommands は有効コマンドのテスト
func TestMainCommandValidator_ValidCommands(t *testing.T) {
    tests := []struct {
        name    string
        command string
        want    bool
    }{
        {"Server command", "server", true},
        {"Disk command", "disk", true},
        {"Config command", "config", true},
        {"Version command", "version", true},
        {"Invalid command", "invalidcmd", false},
        {"Empty command", "", false},
        {"Case sensitive", "Server", false}, // usacloudは小文字
    }
    
    validator := NewMainCommandValidator()
    helper := NewTestHelper(t)
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := validator.IsValidCommand(tt.command)
            helper.AssertEqual(got, tt.want, "Command: %s", tt.command)
        })
    }
}

// TestMainCommandValidator_AllIaaSCommands は全IaaSコマンドのテスト
func TestMainCommandValidator_AllIaaSCommands(t *testing.T) {
    validator := NewMainCommandValidator()
    helper := NewTestHelper(t)
    
    // testdata/commands/iaas_commands.json から読み込み
    var iaasCommands []string
    helper.LoadTestData("commands/iaas_commands.json", &iaasCommands)
    
    for _, command := range iaasCommands {
        t.Run("IaaS_"+command, func(t *testing.T) {
            if !validator.IsValidCommand(command) {
                t.Errorf("IaaSコマンド '%s' が有効として認識されませんでした", command)
            }
            
            commandType := validator.GetCommandType(command)
            helper.AssertEqual(commandType, "iaas", "Command type for %s", command)
        })
    }
}

// TestMainCommandValidator_SimilarCommands は類似コマンド検索のテスト
func TestMainCommandValidator_SimilarCommands(t *testing.T) {
    tests := []struct {
        name      string
        input     string
        maxDist   int
        wantCount int
        wantFirst string
    }{
        {
            name:      "Server typo",
            input:     "serv",
            maxDist:   2,
            wantCount: 1,
            wantFirst: "server",
        },
        {
            name:      "Disk typo", 
            input:     "dsk",
            maxDist:   2,
            wantCount: 1,
            wantFirst: "disk",
        },
        {
            name:      "No similar commands",
            input:     "xxxxxxxxx",
            maxDist:   2,
            wantCount: 0,
            wantFirst: "",
        },
    }
    
    validator := NewMainCommandValidator()
    helper := NewTestHelper(t)
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            similar := validator.GetSimilarCommands(tt.input, tt.maxDist)
            helper.AssertEqual(len(similar), tt.wantCount, "Similar commands count")
            
            if tt.wantFirst != "" && len(similar) > 0 {
                helper.AssertEqual(similar[0], tt.wantFirst, "First similar command")
            }
        })
    }
}
```

#### 2. 類似コマンド提案器テスト
```go
// internal/validation/similar_command_suggester_test.go
package validation

import (
    "testing"
)

// TestLevenshteinDistance はLevenshtein距離計算のテスト
func TestLevenshteinDistance(t *testing.T) {
    tests := []struct {
        name string
        s1   string
        s2   string
        want int
    }{
        {"Same strings", "server", "server", 0},
        {"One character diff", "server", "serve", 1},
        {"Two character diff", "server", "sever", 1},
        {"Complete different", "server", "disk", 5},
        {"Empty string", "", "server", 6},
        {"Case insensitive", "Server", "server", 0},
    }
    
    suggester := NewSimilarCommandSuggester(3, 5)
    helper := NewTestHelper(t)
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := suggester.LevenshteinDistance(tt.s1, tt.s2)
            helper.AssertEqual(got, tt.want, "Distance between '%s' and '%s'", tt.s1, tt.s2)
        })
    }
}

// TestSuggestMainCommands はメインコマンド提案のテスト
func TestSuggestMainCommands(t *testing.T) {
    suggester := NewSimilarCommandSuggester(3, 5)
    helper := NewTestHelper(t)
    
    // 典型的なtypoケース
    tests := []struct {
        name          string
        input         string
        wantMinCount  int
        wantContains  string
        wantMinScore  float64
    }{
        {
            name:          "Server typo", 
            input:         "serv",
            wantMinCount:  1,
            wantContains:  "server",
            wantMinScore:  0.5,
        },
        {
            name:          "Config typo",
            input:         "conf",
            wantMinCount:  1,
            wantContains:  "config",
            wantMinScore:  0.5,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            results := suggester.SuggestMainCommands(tt.input)
            
            if len(results) < tt.wantMinCount {
                t.Errorf("期待する最小候補数 %d, 実際 %d", tt.wantMinCount, len(results))
            }
            
            found := false
            for _, result := range results {
                if result.Command == tt.wantContains {
                    found = true
                    if result.Score < tt.wantMinScore {
                        t.Errorf("スコアが低すぎます: %.2f < %.2f", result.Score, tt.wantMinScore)
                    }
                    break
                }
            }
            
            if !found {
                t.Errorf("期待するコマンド '%s' が候補に含まれていません", tt.wantContains)
            }
        })
    }
}

// BenchmarkLevenshteinDistance はLevenshtein距離計算のベンチマーク
func BenchmarkLevenshteinDistance(b *testing.B) {
    suggester := NewSimilarCommandSuggester(3, 5)
    
    benchmarks := []struct {
        name string
        s1   string
        s2   string
    }{
        {"Short strings", "serv", "server"},
        {"Medium strings", "database", "databse"},
        {"Long strings", "webaccelerator", "webacelerator"},
    }
    
    for _, bm := range benchmarks {
        b.Run(bm.name, func(b *testing.B) {
            for i := 0; i < b.N; i++ {
                suggester.LevenshteinDistance(bm.s1, bm.s2)
            }
        })
    }
}
```

#### 3. エラーメッセージ生成器テスト
```go
// internal/validation/error_message_generator_test.go
package validation

import (
    "strings" 
    "testing"
)

// TestErrorMessageGenerator_GenerateMessage はメッセージ生成のテスト
func TestErrorMessageGenerator_GenerateMessage(t *testing.T) {
    tests := []struct {
        name     string
        msgType  MessageType
        params   map[string]interface{}
        wantContains []string
        colorEnabled bool
    }{
        {
            name:    "Invalid command message",
            msgType: TypeInvalidCommand,
            params: map[string]interface{}{
                "command": "invalidcmd",
            },
            wantContains: []string{"invalidcmd", "有効なusacloudコマンドではありません", "--help"},
            colorEnabled: false,
        },
        {
            name:    "Invalid subcommand message", 
            msgType: TypeInvalidSubcommand,
            params: map[string]interface{}{
                "command":            "server",
                "subcommand":         "invalid",
                "availableSubcommands": []string{"list", "read", "create"},
            },
            wantContains: []string{"invalid", "server", "list", "read", "create"},
            colorEnabled: false,
        },
        {
            name:    "Deprecated command with color",
            msgType: TypeDeprecatedCommand,
            params: map[string]interface{}{
                "command":            "iso-image",
                "replacementCommand": "cdrom",
            },
            wantContains: []string{"iso-image", "廃止", "cdrom"},
            colorEnabled: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            generator := NewErrorMessageGenerator(tt.colorEnabled)
            helper := NewTestHelper(t)
            
            message := generator.GenerateMessage(tt.msgType, tt.params)
            
            for _, want := range tt.wantContains {
                helper.AssertContains(message, want, "Message should contain '%s'", want)
            }
            
            if tt.colorEnabled {
                // カラーコードが含まれているかチェック
                hasColorCode := strings.Contains(message, "\033[")
                if !hasColorCode {
                    t.Error("カラー有効時にカラーコードが含まれていません")
                }
            }
        })
    }
}

// TestErrorMessageGenerator_MessageConsistency はメッセージ一貫性のテスト
func TestErrorMessageGenerator_MessageConsistency(t *testing.T) {
    generator := NewErrorMessageGenerator(false)
    
    // 同じパラメータで複数回生成して一貫性をチェック
    params := map[string]interface{}{"command": "testcmd"}
    
    var messages []string
    for i := 0; i < 5; i++ {
        message := generator.GenerateMessage(TypeInvalidCommand, params)
        messages = append(messages, message)
    }
    
    // 全て同じメッセージであることを確認
    for i := 1; i < len(messages); i++ {
        if messages[0] != messages[i] {
            t.Errorf("メッセージに一貫性がありません:\n1: %s\n%d: %s", messages[0], i+1, messages[i])
        }
    }
}
```

### テストデータ管理

#### 1. テストデータファイル構造
```json
// testdata/commands/iaas_commands.json
[
    "server", "disk", "database", "loadbalancer", "dns", "gslb", "proxylb",
    "autobackup", "archive", "cdrom", "bridge", "packetfilter", "internet",
    "ipaddress", "ipv6addr", "ipv6net", "subnet", "swytch", "localrouter",
    "vpcrouter", "mobilegateway", "sim", "nfs", "license", "licenseinfo",
    "sshkey", "note", "icon", "privatehost", "privatehostplan", "zone",
    "region", "bill", "coupon", "authstatus", "self", "serviceclass",
    "enhanceddb", "containerregistry", "certificateauthority", "esme",
    "simplemonitor", "autoscale", "category"
]

// testdata/commands/subcommands.json
{
    "server": [
        "list", "read", "create", "update", "delete", 
        "boot", "shutdown", "reset", "send-nmi",
        "monitor-cpu", "ssh", "vnc", "rdp",
        "wait-until-ready", "wait-until-shutdown"
    ],
    "disk": [
        "list", "read", "create", "update", "delete",
        "connect", "disconnect"
    ],
    "config": [
        "list", "show", "use", "create", "edit", "delete"
    ]
}

// testdata/errors/typo_patterns.json
{
    "server": ["serv", "sever", "serve", "svr"],
    "disk": ["dsk", "disc", "disks"],
    "database": ["db", "databse", "datbase", "database"],
    "config": ["conf", "cfg", "configure"]
}

// testdata/errors/expected_messages.json
{
    "invalid_command": {
        "template": "エラー: '%s' は有効なusacloudコマンドではありません。\n利用可能なコマンドを確認するには 'usacloud --help' を実行してください。",
        "color_markers": ["\\033[31m", "\\033[0m"],
        "required_elements": ["エラー:", "有効なusacloudコマンドではありません", "usacloud --help"]
    }
}
```

### テスト実行戦略

#### 1. テスト分類とタグ
```go
// テストタグの例
// +build unit

// +build integration

// +build performance

// テスト実行コマンド
// go test -tags=unit ./...              # ユニットテストのみ
// go test -tags=integration ./...       # 統合テストのみ 
// go test -tags=performance ./...       # パフォーマンステストのみ
// go test ./...                         # 全てのテスト
```

#### 2. カバレッジ測定
```bash
# カバレッジ測定とレポート生成
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
go tool cover -func=coverage.out

# カバレッジ閾値チェック
go test -coverprofile=coverage.out ./... && \
go tool cover -func=coverage.out | grep "total:" | \
awk '{if($3+0 < 90) {print "カバレッジが90%未満:", $3; exit 1}}'
```

## テスト戦略
- **高速性**: 全ユニットテストが10秒以内で完了
- **独立性**: テスト間の順序や状態に依存しない設計
- **網羅性**: 正常系・異常系・境界値を包括的にカバー
- **可読性**: テストがドキュメントとして機能するよう命名と構造を最適化
- **保守性**: テストデータの変更が容易で、新機能追加時の拡張が簡単

## 依存関係
- 前提PBI: PBI-001～017 (全実装コンポーネント)
- テストフレームワーク: Go標準testingパッケージ
- 外部ツール: go tool cover (カバレッジ測定)

## 見積もり
- 開発工数: 12時間
  - テストフレームワーク設計: 2時間
  - コンポーネント別テスト実装: 6時間
  - テストデータ作成: 2時間
  - ベンチマークテスト実装: 1時間
  - テスト自動化設定: 1時間

## 完了の定義
- [ ] 全新規コンポーネントのユニットテストが実装されている
- [ ] テストカバレッジが90%以上達成されている
- [ ] 全テストが10秒以内で完了する
- [ ] テストデータが適切に構造化され管理されている
- [ ] ベンチマークテストが主要な処理に実装されている
- [ ] CI/CDパイプラインでの自動実行が設定されている
- [ ] テストドキュメントが作成されている
- [ ] カバレッジレポートが自動生成される
- [ ] 全テストが継続的に通過している
- [ ] コードレビューが完了している

## 実装状況
❌ **PBI-018は未実装** (2025-09-11)

**現在の状況**:
- 詳細なテスト戦略とアーキテクチャが設計済み
- テストフレームワークとユーティリティの仕様が完成
- 実装準備は整っているが、コード実装は未着手

**実装が必要な要素**:
- `internal/validation/*_test.go` - 各バリデーターのユニットテスト
- `internal/validation/testing_utils.go` - テストヘルパー関数群
- `testdata/` - 包括的なテストデータセット
- ベンチマークテストとカバレッジ測定の自動化
- CI/CDパイプラインとの統合

**次のステップ**:
1. テストヘルパーユーティリティの実装
2. メインコンポーネントのユニットテスト作成
3. テストデータの構造化と作成
4. カバレッジ測定とCI統合の設定

## 備考
- テストの実行速度と網羅性のバランスが重要
- 複雑なロジックほど詳細なテストケースが必要
- テストデータの保守性を考慮した設計が重要
- 将来的な機能拡張時のテスト追加を考慮した拡張性が必要

---

## 実装方針変更 (2025-09-11)

🔴 **当PBIは機能拡張のため実装を延期します**

### 延期理由
- 現在のシステム安定化を優先
- 既存機能の修復・改善が急務
- リソースをコア機能の品質向上に集中
- 新規テスト戦略よりも既存テストの修復が緊急

### 再検討時期
- v2.0.0安定版リリース後
- テスト安定化完了後（PBI-024〜030）
- 基幹機能の品質確保完了後
- 既存テストカバレッジ70%達成後

### 現在の優先度
**低優先度** - 将来のロードマップで再評価予定

### 注記
- 既存のユニットテストは引き続き保守・改善
- 新規テスト戦略の実装は延期
- 現在のテスト基盤の安定化を最優先