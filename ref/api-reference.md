# usacloud-update API リファレンス

このドキュメントは、usacloud-update プロジェクトの公開API（Application Programming Interface）の包括的なリファレンスです。開発者がシステムを統合・拡張する際に使用する各パッケージのインターフェース、型、メソッドについて詳細に説明します。

## 目次

- [1. Transform パッケージ](#1-transform-パッケージ)
- [2. Validation パッケージ](#2-validation-パッケージ)
- [3. Config パッケージ](#3-config-パッケージ)
- [4. Sandbox パッケージ](#4-sandbox-パッケージ)
- [5. TUI パッケージ](#5-tui-パッケージ)
- [6. Testing Framework](#6-testing-framework)
- [7. エラーハンドリング](#7-エラーハンドリング)
- [8. 使用例](#8-使用例)

---

## 1. Transform パッケージ

### 概要
Transform パッケージは、usacloudコマンドの変換処理を担当するコアエンジンです。ルールベースの変換システムにより、v0/v1.0のコマンドをv1.1互換に変換します。

### 核心インターフェース

#### Rule インターフェース
```go
type Rule interface {
    Name() string
    Apply(line string) (newLine string, changed bool, beforeFrag string, afterFrag string)
}
```

**目的**: コマンド変換ルールの標準インターフェース

**メソッド**:
- `Name()`: ルール名を返す
- `Apply(line)`: 行に変換ルールを適用し、結果を返す

**戻り値**:
- `newLine`: 変換後の行
- `changed`: 変換が行われたかのフラグ
- `beforeFrag`: 変換前の該当部分
- `afterFrag`: 変換後の該当部分

**スレッドセーフティ**: 読み取り専用操作なので安全

#### Engine 型
```go
type Engine struct {
    rules []Rule
}

func NewDefaultEngine() *Engine
func (e *Engine) Apply(line string) Result
```

**目的**: 変換ルールを順次適用するメインエンジン

**使用例**:
```go
engine := transform.NewDefaultEngine()
result := engine.Apply("usacloud server list --output-type=csv")

if result.Changed {
    fmt.Printf("変換前: %s\n", "usacloud server list --output-type=csv")
    fmt.Printf("変換後: %s\n", result.Line)
    for _, change := range result.Changes {
        fmt.Printf("適用ルール: %s\n", change.RuleName)
    }
}
```

### データ型

#### Result 型
```go
type Result struct {
    Line    string   // 変換後の行
    Changed bool     // 変換が行われたか
    Changes []Change // 適用された変更の詳細
}
```

#### Change 型
```go
type Change struct {
    RuleName string // 適用されたルール名
    Before   string // 変換前の該当部分
    After    string // 変更後の該当部分
}
```

### エラーハンドリング
Transformパッケージはパニックを発生させず、無効な入力に対しては元の行をそのまま返します。

**例**:
```go
// 空行やコメント行はそのまま返される
result := engine.Apply("# これはコメントです")
fmt.Println(result.Changed) // false
fmt.Println(result.Line)    // "# これはコメントです"
```

---

## 2. Validation パッケージ

### 概要
Validation パッケージは、usacloudコマンドの構文解析と妥当性検証を担当します。コマンドライン解析、コマンド辞書による検証、エラーメッセージ生成を提供します。

### 核心型とインターフェース

#### Parser 型
```go
type Parser struct{}

func NewParser() *Parser
func (p *Parser) Parse(commandLine string) (*CommandLine, error)
func (p *Parser) IsUsacloudCommand(commandLine string) bool
```

**目的**: コマンドライン文字列の構文解析

**使用例**:
```go
parser := validation.NewParser()
cmd, err := parser.Parse("usacloud server list --zone=is1a --format=json")
if err != nil {
    log.Fatalf("解析エラー: %v", err)
}

fmt.Printf("メインコマンド: %s\n", cmd.MainCommand)     // "server"
fmt.Printf("サブコマンド: %s\n", cmd.SubCommand)       // "list"
fmt.Printf("オプション: %v\n", cmd.Options)           // map[zone:is1a format:json]
```

#### CommandLine 型
```go
type CommandLine struct {
    Raw         string            // 元のコマンドライン
    MainCommand string            // メインコマンド (server, disk, etc.)
    SubCommand  string            // サブコマンド (list, create, etc.)
    Arguments   []string          // 位置引数
    Options     map[string]string // オプション (--key=value)
    Flags       []string          // フラグ (--force, --dry-run, etc.)
}

// ヘルパーメソッド
func (c *CommandLine) HasOption(key string) bool
func (c *CommandLine) GetOption(key string) string
func (c *CommandLine) HasFlag(flag string) bool
func (c *CommandLine) GetArgument(index int) string
func (c *CommandLine) IsValid() bool
func (c *CommandLine) GetCommandType() string
```

**メソッド解説**:
- `HasOption(key)`: 指定されたオプションが存在するか確認
- `GetOption(key)`: オプション値を取得（存在しない場合は空文字列）
- `HasFlag(flag)`: 指定されたフラグが存在するか確認
- `IsValid()`: コマンド辞書に基づく妥当性検証
- `GetCommandType()`: コマンドタイプ（iaas/misc/root/deprecated/unknown）を返す

### エラー型

#### ParseError 型
```go
type ParseError struct {
    Message  string
    Position int
    Input    string
}

func (e *ParseError) Error() string
```

**使用例**:
```go
_, err := parser.Parse("usacloud server list --invalid='unclosed quote")
if parseErr, ok := err.(*validation.ParseError); ok {
    fmt.Printf("エラー位置: %d\n", parseErr.Position)
    fmt.Printf("エラー内容: %s\n", parseErr.Message)
}
```

### コマンド妥当性検証

#### 検証関数
```go
// IaaS コマンド検証
func IsValidIaaSCommand(mainCommand string) bool
func IsValidIaaSSubcommand(mainCommand, subCommand string) bool

// その他コマンド検証
func IsValidMiscCommand(mainCommand string) bool
func IsValidMiscSubcommand(mainCommand, subCommand string) bool

// ルートコマンド検証
func IsValidRootCommand(mainCommand string) bool
func ValidateRootCommandUsage(mainCommand, subCommand string) string

// 非推奨コマンド検証
func IsDeprecatedCommand(mainCommand string) bool
```

**使用例**:
```go
if validation.IsValidIaaSCommand("server") {
    fmt.Println("server は有効な IaaS コマンドです")
}

if validation.IsDeprecatedCommand("summary") {
    fmt.Println("summary コマンドは非推奨です")
}
```

---

## 3. Config パッケージ

### 概要
Config パッケージは、アプリケーションの設定管理を担当します。INI形式の設定ファイル、プロファイル管理、環境変数による設定オーバーライドをサポートします。

### 核心型

#### IntegratedConfig 型
```go
type IntegratedConfig struct {
    General       *GeneralConfig
    Transform     *TransformConfig
    Validation    *ValidationConfig
    ErrorFeedback *ErrorFeedbackConfig
    HelpSystem    *HelpSystemConfig
    Performance   *PerformanceConfig
    Output        *OutputConfig
    
    Profiles     map[string]*ProfileConfig
    Environments map[string]*EnvironmentConfig
    
    LastModified  time.Time
    ConfigVersion string
}

func NewIntegratedConfig() *IntegratedConfig
func LoadIntegratedConfig(configPath string) (*IntegratedConfig, error)
func (ic *IntegratedConfig) Save() error
func (ic *IntegratedConfig) SaveAs(configPath string) error
func (ic *IntegratedConfig) UpdateSetting(section, key string, value interface{}) error
```

**使用例**:
```go
// 設定の読み込み
config, err := config.LoadIntegratedConfig("/path/to/config.conf")
if err != nil {
    log.Fatalf("設定読み込みエラー: %v", err)
}

// 設定値の取得
fmt.Printf("カラー出力: %t\n", config.General.ColorOutput)
fmt.Printf("厳密モード: %t\n", config.Validation.StrictMode)

// 設定値の更新
err = config.UpdateSetting("general", "color_output", false)
if err != nil {
    log.Printf("設定更新エラー: %v", err)
}
```

### 設定セクション型

#### GeneralConfig 型
```go
type GeneralConfig struct {
    Version              string `ini:"version"`
    ColorOutput          bool   `ini:"color_output"`
    Language             string `ini:"language"`
    Verbose              bool   `ini:"verbose"`
    InteractiveByDefault bool   `ini:"interactive_by_default"`
    Profile              string `ini:"profile"`
}
```

#### ValidationConfig 型
```go
type ValidationConfig struct {
    EnableValidation        bool `ini:"enable_validation"`
    StrictMode              bool `ini:"strict_mode"`
    ValidateBeforeTransform bool `ini:"validate_before_transform"`
    ValidateAfterTransform  bool `ini:"validate_after_transform"`
    MaxSuggestions          int  `ini:"max_suggestions"`
    MaxEditDistance         int  `ini:"max_edit_distance"`
    SkipDeprecatedWarnings  bool `ini:"skip_deprecated_warnings"`
    TypoDetectionEnabled    bool `ini:"typo_detection_enabled"`
}
```

### プロファイル管理

#### ProfileConfig 型
```go
type ProfileConfig struct {
    Name        string                 `ini:"name"`
    Description string                 `ini:"description"`
    BasedOn     string                 `ini:"based_on"`
    Overrides   map[string]interface{} `ini:"-"`
    CreatedAt   time.Time              `ini:"-"`
    LastUsed    time.Time              `ini:"-"`
    UsageCount  int                    `ini:"-"`
}
```

**プロファイル適用例**:
```go
// プロファイルの適用
err := config.applyProfile("beginner")
if err != nil {
    log.Printf("プロファイル適用エラー: %v", err)
}

// 利用可能プロファイルの確認
for name, profile := range config.Profiles {
    fmt.Printf("プロファイル: %s - %s\n", name, profile.Description)
}
```

### 環境変数サポート

設定は以下の環境変数でオーバーライド可能：

- `USACLOUD_UPDATE_PROFILE`: 使用プロファイル
- `USACLOUD_UPDATE_STRICT_MODE`: 厳密モード (true/false)
- `USACLOUD_UPDATE_PARALLEL`: 並列処理 (true/false)
- `USACLOUD_UPDATE_COLOR`: カラー出力 (true/false)
- `USACLOUD_UPDATE_VERBOSE`: 詳細出力 (true/false)

---

## 4. Sandbox パッケージ

### 概要
Sandbox パッケージは、usacloudコマンドの安全な実行環境を提供します。Sakura Cloud の tk1v ゾーン（無料ゾーン）での実行、タイムアウト制御、エラーハンドリングを担当します。

### 核心型

#### Executor 型
```go
type Executor struct {
    config        *config.SandboxConfig
    usacloudRegex *regexp.Regexp
}

func NewExecutor(cfg *config.SandboxConfig) *Executor
func (e *Executor) ExecuteScript(lines []string) ([]*ExecutionResult, error)
func (e *Executor) ExecuteCommand(command string) (*ExecutionResult, error)
func (e *Executor) PrintSummary(results []*ExecutionResult)
```

**使用例**:
```go
// サンドボックス設定
sandboxConfig := &config.SandboxConfig{
    Zone:        "tk1v",
    DryRun:      false,
    Timeout:     30 * time.Second,
    Debug:       true,
}

// Executor の作成
executor := sandbox.NewExecutor(sandboxConfig)

// コマンド実行
result, err := executor.ExecuteCommand("usacloud server list --zone=tk1v")
if err != nil {
    log.Printf("実行エラー: %v", err)
}

if result.Success {
    fmt.Printf("実行成功: %s\n", result.Output)
} else {
    fmt.Printf("実行失敗: %s\n", result.Error)
}
```

#### ExecutionResult 型
```go
type ExecutionResult struct {
    Command    string        `json:"command"`     // 実行されたコマンド
    Success    bool          `json:"success"`     // 実行成功フラグ
    Output     string        `json:"output"`      // 標準出力
    Error      string        `json:"error,omitempty"` // エラーメッセージ
    Duration   time.Duration `json:"duration"`    // 実行時間
    Skipped    bool          `json:"skipped"`     // スキップフラグ
    SkipReason string        `json:"skip_reason,omitempty"` // スキップ理由
}
```

### セキュリティ機能

#### コマンド検証
```go
func (e *Executor) validateCommand(command string) error
```

**安全性チェック**:
- usacloudコマンドで始まること
- --zone=tk1v が指定されること（自動追加）
- 危険な操作（delete, shutdown等）の禁止

**使用例**:
```go
// 危険なコマンドは自動的に拒否される
result, err := executor.ExecuteCommand("usacloud server delete 123456789")
// エラー: "operation 'delete' not allowed in sandbox mode for safety"
```

#### タイムアウト制御
```go
// 30秒でタイムアウト
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()
```

### バッチ実行

#### スクリプト実行
```go
script := []string{
    "usacloud server list",
    "usacloud disk list", 
    "# これはコメントなのでスキップされる",
    "usacloud zone list",
}

results, err := executor.ExecuteScript(script)
if err != nil {
    log.Printf("スクリプト実行エラー: %v", err)
}

// 結果サマリの表示
executor.PrintSummary(results)
```

---

## 5. TUI パッケージ

### 概要
TUI（Terminal User Interface）パッケージは、対話的なターミナルインターフェースを提供します。コマンド選択、実行結果確認、ヘルプ表示などの機能を含みます。

### 核心型

#### App 型
```go
type App struct {
    app      *tview.Application
    config   *config.SandboxConfig
    executor *sandbox.Executor
    commands []*CommandItem
    
    // UI コンポーネント
    commandList *tview.List
    detailView  *tview.TextView  
    resultView  *tview.TextView
    statusBar   *tview.TextView
    helpText    *tview.TextView
    
    // 状態
    currentIndex  int
    helpVisible   bool
}

func NewApp(cfg *config.SandboxConfig) *App
func (a *App) LoadScript(lines []string) error
func (a *App) Run() error
func (a *App) Stop()
```

**使用例**:
```go
// TUI アプリケーションの作成
config := &config.SandboxConfig{
    Zone:    "tk1v", 
    DryRun:  false,
    Timeout: 30 * time.Second,
}

app := tui.NewApp(config)

// スクリプトの読み込み
script := []string{
    "usacloud server list",
    "usacloud disk list",
}
err := app.LoadScript(script)
if err != nil {
    log.Fatalf("スクリプト読み込みエラー: %v", err)
}

// TUI 起動
if err := app.Run(); err != nil {
    log.Fatalf("TUI実行エラー: %v", err)
}
```

#### CommandItem 型
```go
type CommandItem struct {
    Original   string                   // 元のコマンド
    Converted  string                   // 変換後のコマンド
    LineNumber int                      // 行番号
    Changed    bool                     // 変換されたか
    RuleName   string                   // 適用されたルール名
    Selected   bool                     // 選択されているか
    Result     *sandbox.ExecutionResult // 実行結果
}
```

### キーバインド

TUI での操作方法：

| キー | 機能 |
|------|------|
| `Enter` | 選択したコマンドの実行状態を切り替え |
| `Space` | コマンドの選択/選択解除 |
| `a` | 全コマンドを選択 |
| `n` | 全選択を解除 |
| `e` | 選択されたコマンドを実行 |
| `q` | TUI を終了 |
| `↑↓` | コマンドリスト内の移動 |
| `Tab` | パネル間の移動 |
| `?` | ヘルプパネルの表示/非表示切り替え |

### UI レイアウト

#### ヘルプ表示制御
```go
func (a *App) toggleHelp()
```

ヘルプの表示/非表示により、動的にレイアウトが調整されます：

**ヘルプ表示時**:
- メインコンテンツエリア
- ステータスバー
- プログレスバー  
- ヘルプテキスト

**ヘルプ非表示時**:
- メインコンテンツエリア（拡張）
- ステータスバー
- プログレスバー

#### プログレス表示
```go
func (a *App) updateProgressBar(current, total int, message string)
```

コマンド実行時の進捗をリアルタイムで表示：
```
Progress: [████████████░░░░] 75.0% (3/4) - Executing: usacloud disk list
```

---

## 6. Testing Framework

### 概要
Testing Framework は、E2E（End-to-End）テスト、統合テスト、回帰テストを実行するためのフレームワークです。分離されたテスト環境、自動ビルド、結果検証を提供します。

### E2E テストフレームワーク

#### E2ETestSuite 型
```go
type E2ETestSuite struct {
    t          *testing.T
    testDir    string
    binaryPath string
    tempHome   string
    tempConfig string
    timeout    time.Duration
    verbose    bool
}

func NewE2ETestSuite(t *testing.T) *E2ETestSuite
func (e2e *E2ETestSuite) RunE2ETest(testName string, options *E2ETestOptions) *E2ETestResult
func (e2e *E2ETestSuite) CreateTempFile(name, content string) string
func (e2e *E2ETestSuite) CreateTempDir(name string) string
```

**使用例**:
```go
func TestBasicTransformation(t *testing.T) {
    suite := NewE2ETestSuite(t)
    
    // テスト用入力ファイル作成
    inputFile := suite.CreateTempFile("input.sh", `
usacloud server list --output-type=csv
usacloud disk list --selector="Name,Size"
`)
    
    // テストオプション設定
    options := &E2ETestOptions{
        Arguments: []string{"--in", inputFile, "--stats"},
        ExpectedExitCode: 0,
        ExpectedStdout: []string{
            "usacloud server list --output-type=json",
            "2 lines processed",
        },
    }
    
    // E2E テスト実行
    result := suite.RunE2ETest("BasicTransformation", options)
    
    // 追加検証
    assert.True(t, result.ExecutionTime < 5*time.Second)
}
```

#### E2ETestOptions 型
```go
type E2ETestOptions struct {
    // 実行設定
    Arguments   []string          `yaml:"arguments"`
    Environment map[string]string `yaml:"environment"`
    WorkingDir  string            `yaml:"working_dir"`
    Timeout     string            `yaml:"timeout"`
    
    // 入力設定
    StdinInput        string   `yaml:"stdin_input"`
    InteractiveInputs []string `yaml:"interactive_inputs"`
    
    // 期待結果
    ExpectedExitCode int      `yaml:"expected_exit_code"`
    ExpectedStdout   []string `yaml:"expected_stdout"`
    ExpectedStderr   []string `yaml:"expected_stderr"`
    
    // 検証設定
    ValidateOutput      bool   `yaml:"validate_output"`
    ValidateFiles       bool   `yaml:"validate_files"`
    ValidatePerformance bool   `yaml:"validate_performance"`
    MaxExecutionTime    string `yaml:"max_execution_time"`
}
```

#### E2ETestResult 型
```go
type E2ETestResult struct {
    ExitCode      int           `json:"exit_code"`
    Stdout        string        `json:"stdout"`
    Stderr        string        `json:"stderr"`
    ExecutionTime time.Duration `json:"execution_time"`
    MemoryUsage   int64         `json:"memory_usage"`
    FilesCreated  []string      `json:"files_created"`
    FilesModified []string      `json:"files_modified"`
    Error         error         `json:"error,omitempty"`
}
```

### インタラクティブテスト

#### インタラクティブ入力の処理
```go
options := &E2ETestOptions{
    Arguments: []string{"--sandbox", "--interactive"},
    InteractiveInputs: []string{
        "y",           // 設定作成に同意
        "test-token",  // API Token
        "test-secret", // API Secret  
        "q",           // TUI終了
    },
    ExpectedExitCode: 0,
}
```

### 環境分離

#### 分離されたテスト環境
各テストは完全に分離された環境で実行されます：

- 独立した HOME ディレクトリ
- 分離された設定ディレクトリ  
- テンポラリファイルシステム
- 環境変数の分離

#### 環境スナップショット
```go
type EnvironmentSnapshot struct {
    Files     map[string]FileInfo `json:"files"`
    Timestamp time.Time           `json:"timestamp"`
}

func (e2e *E2ETestSuite) takeEnvironmentSnapshot() *EnvironmentSnapshot
```

テスト前後の環境変化を自動で検出し、作成・変更されたファイルを追跡します。

---

## 7. エラーハンドリング

### エラー型の階層

#### 共通エラーインターフェース
すべてのパッケージで一貫したエラーハンドリングを提供：

```go
// Validation パッケージのエラー
var (
    ErrEmptyCommand       = errors.New("empty command")
    ErrNotUsacloudCommand = errors.New("not a usacloud command")
    ErrInvalidSyntax      = errors.New("invalid command syntax")
)

// Config パッケージのエラー
type ConfigError struct {
    Section string
    Key     string
    Message string
}

func (e *ConfigError) Error() string {
    return fmt.Sprintf("config error in [%s].%s: %s", e.Section, e.Key, e.Message)
}
```

### エラーラッピング

Go 1.13+ のエラーラッピングを活用：

```go
// エラーの包装
if err := config.Load(); err != nil {
    return fmt.Errorf("設定読み込みに失敗: %w", err)
}

// エラーのチェック
if errors.Is(err, validation.ErrNotUsacloudCommand) {
    // usacloud コマンドでない場合の処理
}

// 特定のエラー型の抽出
var parseErr *validation.ParseError
if errors.As(err, &parseErr) {
    fmt.Printf("解析エラー (位置 %d): %s", parseErr.Position, parseErr.Message)
}
```

### コンテキストベースエラーハンドリング

#### タイムアウトとキャンセレーション
```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

result, err := executor.ExecuteWithContext(ctx, command)
if err != nil {
    if errors.Is(err, context.DeadlineExceeded) {
        return fmt.Errorf("コマンド実行がタイムアウトしました: %w", err)
    }
    return fmt.Errorf("コマンド実行エラー: %w", err)
}
```

---

## 8. 使用例

### 基本的な変換パイプライン

#### 単純な変換実行
```go
package main

import (
    "fmt"
    "log"
    
    "github.com/armaniacs/usacloud-update/internal/transform"
)

func main() {
    // 変換エンジンの作成
    engine := transform.NewDefaultEngine()
    
    // コマンドの変換
    input := "usacloud server list --output-type=csv --zone=is1a"
    result := engine.Apply(input)
    
    if result.Changed {
        fmt.Printf("変換前: %s\n", input)
        fmt.Printf("変換後: %s\n", result.Line)
        
        for _, change := range result.Changes {
            fmt.Printf("適用ルール: %s (%s → %s)\n", 
                change.RuleName, change.Before, change.After)
        }
    }
}
```

### バリデーション付き変換

#### コマンド妥当性の事前チェック
```go
package main

import (
    "fmt"
    "log"
    
    "github.com/armaniacs/usacloud-update/internal/validation"
    "github.com/armaniacs/usacloud-update/internal/transform"
)

func processCommand(commandLine string) error {
    // 1. コマンド解析
    parser := validation.NewParser()
    cmd, err := parser.Parse(commandLine)
    if err != nil {
        return fmt.Errorf("コマンド解析エラー: %w", err)
    }
    
    // 2. 妥当性検証
    if !cmd.IsValid() {
        return fmt.Errorf("無効なコマンド: %s %s", cmd.MainCommand, cmd.SubCommand)
    }
    
    // 3. 非推奨コマンドの警告
    if cmd.GetCommandType() == "deprecated" {
        fmt.Printf("警告: %s は非推奨コマンドです\n", cmd.MainCommand)
    }
    
    // 4. 変換実行
    engine := transform.NewDefaultEngine()
    result := engine.Apply(commandLine)
    
    fmt.Printf("変換結果: %s\n", result.Line)
    return nil
}
```

### 設定管理の統合

#### プロファイルベース設定
```go
package main

import (
    "fmt"
    "log"
    
    "github.com/armaniacs/usacloud-update/internal/config"
)

func setupConfiguration() (*config.IntegratedConfig, error) {
    // 設定の読み込み
    configPath := "~/.config/usacloud-update/usacloud-update.conf"
    cfg, err := config.LoadIntegratedConfig(configPath)
    if err != nil {
        return nil, fmt.Errorf("設定読み込みエラー: %w", err)
    }
    
    // 環境に応じたプロファイル切り替え
    if cfg.General.Profile == "beginner" {
        fmt.Println("初心者モードで実行します")
        cfg.Validation.MaxSuggestions = 8
        cfg.HelpSystem.ShowCommonMistakes = true
    }
    
    return cfg, nil
}
```

### サンドボックス実行の統合

#### 安全なコマンド実行
```go
package main

import (
    "fmt"
    "log"
    "time"
    
    "github.com/armaniacs/usacloud-update/internal/sandbox"
    "github.com/armaniacs/usacloud-update/internal/config"
)

func executeInSandbox(commands []string) error {
    // サンドボックス設定
    sandboxConfig := &config.SandboxConfig{
        Zone:        "tk1v",
        DryRun:      false,
        Timeout:     30 * time.Second,
        Debug:       true,
    }
    
    // Executor 作成
    executor := sandbox.NewExecutor(sandboxConfig)
    
    // usacloud CLI の存在確認
    if !sandbox.IsUsacloudInstalled() {
        return fmt.Errorf("usacloud CLI がインストールされていません")
    }
    
    // バッチ実行
    results, err := executor.ExecuteScript(commands)
    if err != nil {
        return fmt.Errorf("スクリプト実行エラー: %w", err)
    }
    
    // 結果サマリ
    executor.PrintSummary(results)
    
    // 失敗したコマンドの詳細表示
    for i, result := range results {
        if !result.Success && !result.Skipped {
            fmt.Printf("コマンド %d 失敗: %s\n", i+1, result.Error)
        }
    }
    
    return nil
}
```

### TUI統合アプリケーション

#### インタラクティブ変換・実行
```go
package main

import (
    "bufio"
    "fmt"
    "log"
    "os"
    "time"
    
    "github.com/armaniacs/usacloud-update/internal/tui"
    "github.com/armaniacs/usacloud-update/internal/config"
)

func runInteractiveMode(scriptPath string) error {
    // スクリプト読み込み
    file, err := os.Open(scriptPath)
    if err != nil {
        return fmt.Errorf("ファイル読み込みエラー: %w", err)
    }
    defer file.Close()
    
    var lines []string
    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        lines = append(lines, scanner.Text())
    }
    
    // サンドボックス設定
    sandboxConfig := &config.SandboxConfig{
        Zone:    "tk1v",
        DryRun:  false,
        Timeout: 30 * time.Second,
    }
    
    // TUI アプリ作成
    app := tui.NewApp(sandboxConfig)
    
    // スクリプト読み込み
    if err := app.LoadScript(lines); err != nil {
        return fmt.Errorf("スクリプト読み込みエラー: %w", err)
    }
    
    // TUI 実行
    fmt.Println("対話型モードを開始します...")
    fmt.Println("操作方法: Enter=選択切替, Space=実行切替, e=実行, ?=ヘルプ, q=終了")
    
    return app.Run()
}
```

### E2E テストの実装

#### カスタムテストケース
```go
package main

import (
    "testing"
    "time"
    
    "github.com/armaniacs/usacloud-update/tests/e2e"
)

func TestComplexWorkflow(t *testing.T) {
    suite := e2e.NewE2ETestSuite(t)
    
    // 複雑な入力スクリプト
    scriptContent := `#!/bin/bash
# 複数のコマンドを含むスクリプト
usacloud server list --output-type=csv --zone=is1a
usacloud disk list --selector="Name,Size" 
usacloud cdrom list --zone=all
usacloud ipaddress list --output-type=tsv
`
    
    inputFile := suite.CreateTempFile("complex_script.sh", scriptContent)
    outputFile := suite.CreateTempFile("output.sh", "")
    
    // テストオプション
    options := &e2e.E2ETestOptions{
        Arguments: []string{
            "--in", inputFile,
            "--out", outputFile, 
            "--stats",
        },
        ExpectedExitCode: 0,
        ExpectedStdout: []string{
            "4 lines processed",
            "4 lines changed",
        },
        ExpectedFiles: []e2e.FileExpectation{
            {
                Path:        "output.sh",
                ShouldExist: true,
                ContentContains: []string{
                    "--output-type=json",
                    "usacloud-update:",
                },
            },
        },
        ValidateFiles:       true,
        ValidatePerformance: true,
        MaxExecutionTime:    "10s",
    }
    
    // テスト実行
    result := suite.RunE2ETest("ComplexWorkflow", options)
    
    // カスタム検証
    if result.ExecutionTime > 5*time.Second {
        t.Errorf("実行時間が想定より長い: %v", result.ExecutionTime)
    }
}
```

---

## まとめ

このAPIリファレンスは、usacloud-updateプロジェクトの主要なパッケージとその使用方法を包括的に説明しています。各パッケージは明確に分離された責務を持ち、一貫したAPIインターフェースを提供します。

### 開発者向けベストプラクティス

1. **エラーハンドリング**: Go 1.13+ のエラーラッピングを活用
2. **コンテキスト**: 長時間実行される操作にはcontextを使用
3. **テスト**: 各機能について包括的なテストを作成
4. **設定管理**: プロファイルベースの設定でユーザビリティを向上
5. **セキュリティ**: サンドボックス環境での安全なコマンド実行

### 拡張性

このAPIは拡張性を考慮して設計されており、以下の拡張が可能です：

- **新しい変換ルール**: `Rule`インターフェースの実装
- **カスタムバリデータ**: 独自のコマンド検証ロジック
- **プラグインシステム**: 外部パッケージの統合
- **カスタムテストフレームワーク**: 独自のテスト環境構築

詳細な実装例や最新の情報については、プロジェクトの他のドキュメントや実際のソースコードを参照してください。