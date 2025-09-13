# PBI-028: Integration Test設定修復

## 概要
統合テストフレームワークで設定関連のテスト失敗が発生しており、CI/CDパイプラインでの品質保証に影響を与えている。特にgo.modパス問題やビルド設定の不整合により、統合テストが不安定な状態となっており、継続的品質監視が困難になっている。

## 受け入れ条件
- [ ] 統合テストが安定して実行できること
- [ ] go.modパス問題が完全に解消されること
- [ ] ビルド・実行環境の設定が適切に構成されること
- [ ] CI/CD環境での統合テスト実行が安定すること

## 技術仕様

### 現在の問題
```bash
# 確認されるテスト失敗パターン
=== RUN   TestIntegrationFramework_Build
    integration_test_framework.go:XX: go build failed: go.mod file not found
--- FAIL: TestIntegrationFramework_Build

=== RUN   TestIntegrationFramework_Execute
    integration_test_framework.go:XX: execution failed: binary not found
--- FAIL: TestIntegrationFramework_Execute
```

### 1. go.modパス問題の完全解決
#### 現在の修正済み箇所の確認・強化
```go
// tests/integration/integration_test_framework.go
func (its *IntegrationTestSuite) BuildBinary() error {
    // 修正済み: プロジェクトルートでビルド実行
    rootDir := its.findProjectRoot()
    
    cmd := exec.Command("go", "build", "-o", its.binaryPath, "./cmd/usacloud-update")
    cmd.Dir = rootDir  // 既に修正済み
    cmd.Env = append(os.Environ(), "CGO_ENABLED=0")
    
    output, err := cmd.CombinedOutput()
    if err != nil {
        return fmt.Errorf("build failed: %w\nOutput: %s", err, string(output))
    }
    
    return nil
}

// プロジェクトルート検索の堅牢化
func (its *IntegrationTestSuite) findProjectRoot() string {
    dir, err := os.Getwd()
    if err != nil {
        return "."
    }
    
    // go.modファイルを探してプロジェクトルートを特定
    for {
        if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
            return dir
        }
        
        parent := filepath.Dir(dir)
        if parent == dir {
            // ルートディレクトリに到達した場合は現在のディレクトリを返す
            return "."
        }
        dir = parent
    }
}
```

### 2. ビルド・実行環境の堅牢化
#### テスト環境の分離と清掃
```go
type IntegrationTestSuite struct {
    tempDir    string
    binaryPath string
    configDir  string
    projectRoot string
}

func NewIntegrationTestSuite(t *testing.T) *IntegrationTestSuite {
    tempDir := t.TempDir()
    projectRoot := findProjectRootFromCwd()
    
    return &IntegrationTestSuite{
        tempDir:     tempDir,
        binaryPath:  filepath.Join(tempDir, "usacloud-update"),
        configDir:   filepath.Join(tempDir, "config"),
        projectRoot: projectRoot,
    }
}

func (its *IntegrationTestSuite) SetupTestEnvironment() error {
    // 設定ディレクトリの作成
    if err := os.MkdirAll(its.configDir, 0700); err != nil {
        return fmt.Errorf("failed to create config dir: %w", err)
    }
    
    // テスト用設定ファイルの作成
    configContent := `[sakuracloud]
access_token = test_token_integration
access_token_secret = test_secret_integration
zone = tk1v

[usacloud_update]
sandbox_enabled = true
timeout = 30
`
    
    configFile := filepath.Join(its.configDir, "usacloud-update.conf")
    if err := ioutil.WriteFile(configFile, []byte(configContent), 0600); err != nil {
        return fmt.Errorf("failed to create config file: %w", err)
    }
    
    return nil
}

func (its *IntegrationTestSuite) BuildBinary() error {
    cmd := exec.Command("go", "build", "-o", its.binaryPath, "./cmd/usacloud-update")
    cmd.Dir = its.projectRoot
    cmd.Env = append(os.Environ(), 
        "CGO_ENABLED=0",
        "GOOS="+runtime.GOOS,
        "GOARCH="+runtime.GOARCH,
    )
    
    output, err := cmd.CombinedOutput()
    if err != nil {
        return fmt.Errorf("build failed in %s: %w\nOutput: %s", its.projectRoot, err, string(output))
    }
    
    // バイナリの存在確認
    if _, err := os.Stat(its.binaryPath); os.IsNotExist(err) {
        return fmt.Errorf("binary was not created at %s", its.binaryPath)
    }
    
    return nil
}
```

### 3. テストフラグ・オプションの整理
#### 不適切なフラグの除去・修正
```go
func (its *IntegrationTestSuite) ExecuteCommand(args []string) (*IntegrationTestResult, error) {
    // 修正済み: 不適切なフラグを除去
    validArgs := its.filterValidArgs(args)
    
    cmd := exec.Command(its.binaryPath, validArgs...)
    cmd.Dir = its.tempDir
    cmd.Env = append(os.Environ(), 
        "USACLOUD_UPDATE_CONFIG_DIR="+its.configDir,
        "HOME="+its.tempDir,  // ホームディレクトリの隔離
    )
    
    var stdout, stderr bytes.Buffer
    cmd.Stdout = &stdout
    cmd.Stderr = &stderr
    
    startTime := time.Now()
    err := cmd.Run()
    duration := time.Since(startTime)
    
    result := &IntegrationTestResult{
        ExitCode: 0,
        Stdout:   stdout.String(),
        Stderr:   stderr.String(),
        Duration: duration,
    }
    
    if err != nil {
        if exitError, ok := err.(*exec.ExitError); ok {
            result.ExitCode = exitError.ExitCode()
        } else {
            return nil, fmt.Errorf("command execution failed: %w", err)
        }
    }
    
    return result, nil
}

// 有効な引数のみをフィルタリング
func (its *IntegrationTestSuite) filterValidArgs(args []string) []string {
    invalidFlags := map[string]bool{
        "--config":  true,  // 修正済み: 除去対象
        "--profile": true,  // 修正済み: 除去対象
    }
    
    var validArgs []string
    skip := false
    
    for i, arg := range args {
        if skip {
            skip = false
            continue
        }
        
        // フラグの形式確認
        if strings.HasPrefix(arg, "--") {
            flagName := strings.Split(arg, "=")[0]
            if invalidFlags[flagName] {
                // 次の引数も値として除外（= がない場合）
                if !strings.Contains(arg, "=") && i+1 < len(args) {
                    skip = true
                }
                continue
            }
        }
        
        validArgs = append(validArgs, arg)
    }
    
    return validArgs
}
```

### 4. 包括的テスト実装
#### 統合テストの安定化
```go
func TestIntegrationFramework_Complete(t *testing.T) {
    suite := NewIntegrationTestSuite(t)
    
    // 環境セットアップ
    if err := suite.SetupTestEnvironment(); err != nil {
        t.Fatalf("Failed to setup test environment: %v", err)
    }
    
    // ビルドテスト
    if err := suite.BuildBinary(); err != nil {
        t.Fatalf("Failed to build binary: %v", err)
    }
    
    // 基本実行テスト
    result, err := suite.ExecuteCommand([]string{"--help"})
    if err != nil {
        t.Fatalf("Failed to execute help command: %v", err)
    }
    
    if result.ExitCode != 0 {
        t.Errorf("Help command failed with exit code %d", result.ExitCode)
    }
    
    if !strings.Contains(result.Stdout, "usacloud-update") {
        t.Errorf("Help output does not contain expected content: %s", result.Stdout)
    }
    
    // バージョン確認テスト
    result, err = suite.ExecuteCommand([]string{"--version"})
    if err != nil {
        t.Fatalf("Failed to execute version command: %v", err)
    }
    
    if result.ExitCode != 0 {
        t.Errorf("Version command failed with exit code %d", result.ExitCode)
    }
}

func TestIntegrationFramework_ErrorHandling(t *testing.T) {
    suite := NewIntegrationTestSuite(t)
    
    if err := suite.SetupTestEnvironment(); err != nil {
        t.Fatalf("Failed to setup test environment: %v", err)
    }
    
    if err := suite.BuildBinary(); err != nil {
        t.Fatalf("Failed to build binary: %v", err)
    }
    
    // 無効な引数でのエラーテスト
    result, err := suite.ExecuteCommand([]string{"--invalid-flag"})
    if err != nil {
        t.Fatalf("Failed to execute invalid command: %v", err)
    }
    
    if result.ExitCode == 0 {
        t.Errorf("Expected non-zero exit code for invalid flag")
    }
}
```

## テスト戦略
- **ビルドテスト**: go.mod問題の回帰防止テスト
- **実行環境テスト**: 分離された環境での動作確認
- **エラーハンドリングテスト**: 不正な引数・環境での動作確認
- **CI/CD統合テスト**: 継続的統合環境での安定性確認

## 依存関係
- 前提PBI: なし（独立した修復タスク）
- 関連PBI: PBI-029（Test Coverage向上）、PBI-030（Regression Test安定化）
- 既存コード: tests/integration/ パッケージ

## 見積もり
- 開発工数: 4時間
  - go.modパス問題の完全解決: 1時間
  - ビルド・実行環境の堅牢化: 1.5時間
  - テストフラグ整理: 1時間
  - テスト強化: 0.5時間

## 完了の定義
- [ ] 統合テストが100%安定して実行できる
- [ ] go.mod関連エラーが完全解消
- [ ] CI/CD環境での統合テスト実行が成功
- [ ] テスト環境の分離と清掃が適切に動作
- [ ] 不適切なフラグ使用の問題が解消

## 備考
- 統合テストの安定性はCI/CDパイプラインの信頼性に直結
- テスト環境の適切な分離により、並列実行時の競合状態を防止
- プロジェクトルート検索の堅牢化により、様々な実行環境での安定性を確保

---

## 実装状況 (2025-09-11)

🟠 **PBI-028は未実装** (2025-09-11)

### 現在の状況
- 統合テストフレームワークで設定関連のテスト失敗が発生
- go.modパス問題やビルド設定の不整合が存在
- CI/CDパイプラインでの品質保証に影響
- 統合テストが不安定で継続的品質監視が困難

### 未実装要素
1. **go.modパス問題の完全解決**
   - プロジェクトルート検索の堅牢化
   - findProjectRoot()機能の完全実装
   - go.modファイルアクセスの確実性確保
   - 様々な実行環境での安定性確保

2. **ビルド・実行環境の堅牢化**
   - BuildBinary()機能の安定化
   - バイナリビルドの確実な実行
   - 実行環境の適切な設定と管理
   - テスト環境の分離と清掃

3. **テストフラグ整理と強化**
   - 不適切なフラグ使用の問題解決
   - CI/CD環境での統合テスト実行の安定化
   - ビルドテストの回帰防止テスト
   - エラーハンドリングテストの実装

### 次のステップ
1. go.modパス問題の完全解決とプロジェクトルート検索堅牢化
2. BuildBinary()機能の安定化とビルド環境堅牢化
3. テストフラグの整理と不適切なフラグ使用の解決
4. CI/CD環境での統合テスト安定化と検証
5. テスト環境分離と清掃機能の実装

### 技術要件
- Go 1.24.1対応
- 統合テストフレームワークの安定化
- CI/CDパイプラインでの確実な動作
- 4時間の作業見積もり

### 受け入れ条件の進捗
- [ ] 統合テストが安定して実行できること
- [ ] go.modパス問題が完全に解消されること
- [ ] ビルド・実行環境の設定が適切に構成されること
- [ ] CI/CD環境での統合テスト実行が安定すること