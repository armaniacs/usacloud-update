# PBI-017: 設定・プロファイル統合

## 概要
新しいコマンド検証・エラーフィードバックシステムの設定を既存のusacloud-update設定システムと統合し、ユーザープロファイル機能、設定の階層管理、環境別設定の一元管理を実装する。複雑な設定を直感的に管理できるシステムを提供する。

## 受け入れ条件
- [ ] 既存の設定ファイル形式と完全互換性が保たれている
- [ ] ユーザープロファイル機能が実装されている
- [ ] 環境別設定（開発・本番・テスト）が適切に管理できる
- [ ] 設定の優先度制御が正しく動作している
- [ ] 設定変更の即座反映と永続化が実装されている

## 技術仕様

### 統合設定アーキテクチャ

#### 1. 統合設定ファイル構造
```ini
# usacloud-update.conf - 統合設定ファイル

[general]
# 基本設定
version = "1.9.0"
color_output = true
language = "ja"
verbose = false
interactive_by_default = false
profile = "default"

[transform]
# 変換設定（既存機能）
preserve_comments = true
add_explanatory_comments = true
show_line_numbers = true
backup_original = false

[validation]
# 検証設定（新機能）
enable_validation = true
strict_mode = false
validate_before_transform = true
validate_after_transform = true
max_suggestions = 5
max_edit_distance = 3
skip_deprecated_warnings = false
typo_detection_enabled = true

[error_feedback]
# エラーフィードバック設定
error_format = "comprehensive"
show_suggestions = true
show_alternatives = true
show_migration_guide = true
suggestion_confidence_threshold = 0.5

[help_system]
# ヘルプシステム設定
enable_interactive_help = true
skill_level = "intermediate"  # beginner, intermediate, advanced, expert
preferred_help_format = "detailed"  # basic, detailed, interactive, example
show_common_mistakes = true
enable_learning_tracking = true

[performance]
# パフォーマンス設定
parallel_processing = true
cache_enabled = true
cache_size_mb = 100
batch_size = 1000
worker_count = 0  # 0 = auto-detect

[output]
# 出力設定
format = "auto"  # auto, plain, colored, json
show_progress = true
progress_style = "bar"  # bar, percentage, dots
report_level = "summary"  # minimal, summary, detailed

[profiles.default]
# デフォルトプロファイル
description = "標準設定"
based_on = ""

[profiles.beginner]
# 初心者向けプロファイル
description = "初心者向け設定"
based_on = "default"
interactive_by_default = true
verbose = true
skill_level = "beginner"
max_suggestions = 8
enable_interactive_help = true
show_common_mistakes = true

[profiles.expert]
# エキスパート向けプロファイル  
description = "エキスパート向け設定"
based_on = "default"
strict_mode = true
max_suggestions = 3
parallel_processing = true
show_progress = false
report_level = "minimal"

[profiles.ci]
# CI/CD環境向けプロファイル
description = "CI/CD環境向け設定"
based_on = "expert"
color_output = false
verbose = false
interactive_by_default = false
show_progress = false
report_level = "detailed"

[environments.development]
# 開発環境設定
sakura_api_endpoint = "https://secure.sakura.ad.jp/cloud/zone/is1a/api/cloud/1.1/"
timeout_seconds = 30
retry_count = 3

[environments.production]  
# 本番環境設定
sakura_api_endpoint = "https://secure.sakura.ad.jp/cloud/zone/is1a/api/cloud/1.1/"
timeout_seconds = 60
retry_count = 5
strict_mode = true
```

#### 2. 統合設定管理システム
```go
// internal/config/integrated_config.go
package config

import (
    "fmt"
    "os"
    "path/filepath"
    "time"
    
    "gopkg.in/ini.v1"
)

// IntegratedConfig は統合設定システム
type IntegratedConfig struct {
    // 基本設定
    configPath     string
    profileName    string
    environmentName string
    
    // 設定データ
    General        *GeneralConfig
    Transform      *TransformConfig
    Validation     *ValidationConfig
    ErrorFeedback  *ErrorFeedbackConfig
    HelpSystem     *HelpSystemConfig
    Performance    *PerformanceConfig
    Output         *OutputConfig
    
    // プロファイル・環境
    Profiles       map[string]*ProfileConfig
    Environments   map[string]*EnvironmentConfig
    
    // メタデータ
    LastModified   time.Time
    ConfigVersion  string
    autoSave       bool
}

// ProfileConfig はプロファイル設定
type ProfileConfig struct {
    Name        string
    Description string
    BasedOn     string
    Overrides   map[string]interface{}
    CreatedAt   time.Time
    LastUsed    time.Time
    UsageCount  int
}

// EnvironmentConfig は環境設定
type EnvironmentConfig struct {
    Name              string
    SakuraAPIEndpoint string
    TimeoutSeconds    int
    RetryCount        int
    StrictMode        bool
    Overrides         map[string]interface{}
}

// ValidationConfig は検証設定
type ValidationConfig struct {
    EnableValidation       bool    `ini:"enable_validation"`
    StrictMode            bool    `ini:"strict_mode"`
    ValidateBeforeTransform bool   `ini:"validate_before_transform"`
    ValidateAfterTransform bool    `ini:"validate_after_transform"`
    MaxSuggestions        int     `ini:"max_suggestions"`
    MaxEditDistance       int     `ini:"max_edit_distance"`
    SkipDeprecatedWarnings bool   `ini:"skip_deprecated_warnings"`
    TypoDetectionEnabled  bool    `ini:"typo_detection_enabled"`
}

// LoadIntegratedConfig は統合設定を読み込み
func LoadIntegratedConfig(configPath string) (*IntegratedConfig, error) {
    config := &IntegratedConfig{
        configPath:   configPath,
        Profiles:     make(map[string]*ProfileConfig),
        Environments: make(map[string]*EnvironmentConfig),
        autoSave:     true,
    }
    
    // 設定ファイル読み込み
    if err := config.loadFromFile(); err != nil {
        if os.IsNotExist(err) {
            // 設定ファイルが存在しない場合はデフォルト設定作成
            if err := config.createDefaultConfig(); err != nil {
                return nil, fmt.Errorf("デフォルト設定作成に失敗: %w", err)
            }
        } else {
            return nil, fmt.Errorf("設定ファイル読み込みに失敗: %w", err)
        }
    }
    
    // 環境変数オーバーライド
    config.applyEnvironmentOverrides()
    
    // プロファイル適用
    if err := config.applyProfile(config.General.Profile); err != nil {
        return nil, fmt.Errorf("プロファイル適用に失敗: %w", err)
    }
    
    return config, nil
}
```

### プロファイル管理システム

#### 1. プロファイル操作
```go
// ProfileManager はプロファイル管理器
type ProfileManager struct {
    config      *IntegratedConfig
    profilesDir string
}

// CreateProfile は新しいプロファイルを作成
func (pm *ProfileManager) CreateProfile(name, basedOn, description string) error {
    if _, exists := pm.config.Profiles[name]; exists {
        return fmt.Errorf("プロファイル '%s' は既に存在します", name)
    }
    
    // ベースプロファイルの検証
    var baseProfile *ProfileConfig
    if basedOn != "" {
        var exists bool
        baseProfile, exists = pm.config.Profiles[basedOn]
        if !exists {
            return fmt.Errorf("ベースプロファイル '%s' が見つかりません", basedOn)
        }
    }
    
    // 新プロファイル作成
    newProfile := &ProfileConfig{
        Name:        name,
        Description: description,
        BasedOn:     basedOn,
        Overrides:   make(map[string]interface{}),
        CreatedAt:   time.Now(),
        UsageCount:  0,
    }
    
    // ベースプロファイルの設定を継承
    if baseProfile != nil {
        for key, value := range baseProfile.Overrides {
            newProfile.Overrides[key] = value
        }
    }
    
    pm.config.Profiles[name] = newProfile
    
    // 自動保存
    if pm.config.autoSave {
        return pm.config.Save()
    }
    
    return nil
}

// SwitchProfile はアクティブプロファイルを切り替え
func (pm *ProfileManager) SwitchProfile(profileName string) error {
    profile, exists := pm.config.Profiles[profileName]
    if !exists {
        return fmt.Errorf("プロファイル '%s' が見つかりません", profileName)
    }
    
    // 現在のプロファイルの使用統計を更新
    if currentProfile, exists := pm.config.Profiles[pm.config.profileName]; exists {
        currentProfile.LastUsed = time.Now()
        currentProfile.UsageCount++
    }
    
    // プロファイル切り替え
    pm.config.profileName = profileName
    pm.config.General.Profile = profileName
    
    // プロファイル設定を適用
    if err := pm.config.applyProfile(profileName); err != nil {
        return fmt.Errorf("プロファイル適用に失敗: %w", err)
    }
    
    // 統計更新
    profile.LastUsed = time.Now()
    profile.UsageCount++
    
    if pm.config.autoSave {
        return pm.config.Save()
    }
    
    return nil
}

// ListProfiles は利用可能なプロファイル一覧を取得
func (pm *ProfileManager) ListProfiles() []*ProfileConfig {
    profiles := make([]*ProfileConfig, 0, len(pm.config.Profiles))
    for _, profile := range pm.config.Profiles {
        profiles = append(profiles, profile)
    }
    return profiles
}
```

#### 2. 動的設定更新
```go
// ConfigWatcher は設定変更監視器
type ConfigWatcher struct {
    config     *IntegratedConfig
    watchers   []chan ConfigChangeEvent
    stopChan   chan bool
}

// ConfigChangeEvent は設定変更イベント
type ConfigChangeEvent struct {
    Section   string
    Key       string
    OldValue  interface{}
    NewValue  interface{}
    Timestamp time.Time
}

// UpdateSetting は設定を動的に更新
func (ic *IntegratedConfig) UpdateSetting(section, key string, value interface{}) error {
    oldValue := ic.getSetting(section, key)
    
    // 設定値の更新
    if err := ic.setSetting(section, key, value); err != nil {
        return err
    }
    
    // 変更イベントの通知
    event := ConfigChangeEvent{
        Section:   section,
        Key:       key,
        OldValue:  oldValue,
        NewValue:  value,
        Timestamp: time.Now(),
    }
    
    ic.notifyConfigChange(event)
    
    // 自動保存
    if ic.autoSave {
        return ic.Save()
    }
    
    return nil
}

// applyEnvironmentOverrides は環境変数からのオーバーライドを適用
func (ic *IntegratedConfig) applyEnvironmentOverrides() {
    // 既存の環境変数サポート（後方互換性）
    if token := os.Getenv("SAKURACLOUD_ACCESS_TOKEN"); token != "" {
        // 認証情報は環境変数優先
    }
    
    // 新しい環境変数
    envMappings := map[string]func(string){
        "USACLOUD_UPDATE_PROFILE": func(v string) {
            ic.General.Profile = v
        },
        "USACLOUD_UPDATE_STRICT_MODE": func(v string) {
            ic.Validation.StrictMode = (v == "true" || v == "1")
        },
        "USACLOUD_UPDATE_PARALLEL": func(v string) {
            ic.Performance.ParallelProcessing = (v == "true" || v == "1")
        },
        "USACLOUD_UPDATE_COLOR": func(v string) {
            ic.General.ColorOutput = (v == "true" || v == "1")
        },
    }
    
    for envVar, setter := range envMappings {
        if value := os.Getenv(envVar); value != "" {
            setter(value)
        }
    }
}
```

### 設定CLI拡張

#### 1. 設定管理コマンド
```bash
# プロファイル管理
usacloud-update config profile list                    # プロファイル一覧
usacloud-update config profile create myprofile        # プロファイル作成
usacloud-update config profile use myprofile           # プロファイル切替
usacloud-update config profile delete myprofile        # プロファイル削除
usacloud-update config profile show myprofile          # プロファイル詳細

# 設定表示・編集
usacloud-update config show                            # 現在の設定表示
usacloud-update config show --profile expert           # 特定プロファイルの設定
usacloud-update config edit                            # 設定ファイルをエディタで編集
usacloud-update config set validation.strict_mode true # 設定値を個別変更
usacloud-update config get validation.max_suggestions  # 設定値を個別取得

# 環境管理
usacloud-update config env list                        # 環境一覧
usacloud-update config env use production              # 環境切替
usacloud-update config env show                        # 現在の環境設定

# 設定検証・リセット
usacloud-update config validate                        # 設定ファイル検証
usacloud-update config reset                           # デフォルト設定にリセット
usacloud-update config backup                          # 設定バックアップ
usacloud-update config restore                         # 設定復元
```

#### 2. インタラクティブ設定
```go
// InteractiveConfigManager はインタラクティブ設定管理器
type InteractiveConfigManager struct {
    config  *IntegratedConfig
    ui      *ConfigUI
}

// RunInteractiveSetup はインタラクティブ設定を実行
func (icm *InteractiveConfigManager) RunInteractiveSetup() error {
    fmt.Println("🚀 usacloud-update 設定セットアップ")
    fmt.Println("==================================")
    
    // Step 1: 基本設定
    if err := icm.setupBasicConfig(); err != nil {
        return err
    }
    
    // Step 2: プロファイル選択
    if err := icm.selectProfile(); err != nil {
        return err
    }
    
    // Step 3: 検証設定
    if err := icm.setupValidationConfig(); err != nil {
        return err
    }
    
    // Step 4: 出力設定
    if err := icm.setupOutputConfig(); err != nil {
        return err
    }
    
    // Step 5: 設定保存
    return icm.config.Save()
}

func (icm *InteractiveConfigManager) selectProfile() error {
    fmt.Println("\n📋 プロファイル選択:")
    fmt.Println("   1. default    - 標準設定")
    fmt.Println("   2. beginner   - 初心者向け（丁寧なヘルプ）")
    fmt.Println("   3. expert     - エキスパート向け（最小出力）")
    fmt.Println("   4. ci         - CI/CD環境向け")
    fmt.Println("   5. custom     - カスタムプロファイル作成")
    
    choice := icm.ui.promptChoice("選択してください [1-5]", []string{"1", "2", "3", "4", "5"})
    
    profileMap := map[string]string{
        "1": "default",
        "2": "beginner", 
        "3": "expert",
        "4": "ci",
    }
    
    if profileName, exists := profileMap[choice]; exists {
        icm.config.General.Profile = profileName
        fmt.Printf("✅ プロファイル '%s' を選択しました\n", profileName)
    } else if choice == "5" {
        return icm.createCustomProfile()
    }
    
    return nil
}
```

### マイグレーション支援

#### 1. 設定マイグレーション
```go
// ConfigMigrator は設定マイグレーション処理器
type ConfigMigrator struct {
    fromVersion string
    toVersion   string
}

// MigrateConfig は設定ファイルをマイグレート
func (cm *ConfigMigrator) MigrateConfig(configPath string) error {
    // 既存設定ファイルの読み込み
    oldConfig, err := cm.loadOldConfig(configPath)
    if err != nil {
        return err
    }
    
    // バックアップ作成
    backupPath := configPath + ".backup." + time.Now().Format("20060102-150405")
    if err := cm.backupConfig(configPath, backupPath); err != nil {
        return fmt.Errorf("バックアップ作成に失敗: %w", err)
    }
    
    // 新形式に変換
    newConfig, err := cm.convertConfig(oldConfig)
    if err != nil {
        return fmt.Errorf("設定変換に失敗: %w", err)
    }
    
    // 新設定ファイル保存
    if err := newConfig.SaveAs(configPath); err != nil {
        return fmt.Errorf("新設定保存に失敗: %w", err)
    }
    
    fmt.Printf("✅ 設定ファイルを v%s から v%s に更新しました\n", cm.fromVersion, cm.toVersion)
    fmt.Printf("   バックアップ: %s\n", backupPath)
    
    return nil
}

// 既存の.envファイルからの移行
func (cm *ConfigMigrator) MigrateFromEnvFile(envPath, configPath string) error {
    fmt.Println("🔄 .envファイルから新設定形式への移行を開始します")
    
    envVars, err := cm.loadEnvFile(envPath)
    if err != nil {
        return err
    }
    
    config := NewDefaultIntegratedConfig()
    
    // 環境変数を設定に変換
    envMappings := map[string]func(string){
        "SAKURACLOUD_ACCESS_TOKEN":        func(v string) { /* 認証設定 */ },
        "SAKURACLOUD_ACCESS_TOKEN_SECRET": func(v string) { /* 認証設定 */ },
        "USACLOUD_COLOR_OUTPUT":          func(v string) { 
            config.General.ColorOutput = (v == "true")
        },
    }
    
    for envKey, value := range envVars {
        if mapper, exists := envMappings[envKey]; exists {
            mapper(value)
        }
    }
    
    return config.SaveAs(configPath)
}
```

## テスト戦略
- 設定ファイルテスト：各種設定ファイル形式の読み書きが正しく動作することを確認
- プロファイルテスト：プロファイル作成・切り替え・削除が正しく動作することを確認
- 環境変数テスト：環境変数オーバーライドが正しく適用されることを確認
- マイグレーションテスト：既存設定からの移行が正しく行われることを確認
- 動的更新テスト：設定の動的変更と反映が正しく動作することを確認
- インタラクティブテスト：対話的設定が期待通りに動作することを確認

## 依存関係
- 前提PBI: PBI-015 (統合CLI), PBI-016 (変換エンジン統合)
- 既存コード: 既存の設定システムとの統合
- 外部ライブラリ: gopkg.in/ini.v1 (INIファイル処理)

## 見積もり
- 開発工数: 8時間
  - 統合設定システム実装: 3時間
  - プロファイル管理実装: 2時間
  - 設定CLIコマンド実装: 1.5時間
  - マイグレーション機能実装: 1時間
  - ユニットテスト作成: 0.5時間

## 完了の定義
- [ ] `internal/config/integrated_config.go`ファイルが作成されている
- [ ] 統合設定ファイル形式が実装され、既存形式との互換性が保たれている
- [ ] プロファイル管理システムが完全に実装されている
- [ ] 環境別設定管理が実装されている
- [ ] 設定管理CLIコマンドが実装されている
- [ ] インタラクティブ設定機能が実装されている
- [ ] 設定マイグレーション機能が実装されている
- [ ] 動的設定変更と即座反映が実装されている
- [ ] 包括的なユニットテストが作成され、すべて通過している
- [ ] 既存設定からの移行テストが通過している
- [ ] コードレビューが完了している

## 備考
- 既存ユーザーの設定に影響を与えない移行パスが最重要
- 複雑な設定をシンプルに管理できるUXが重要
- 設定ファイルの可読性と保守性を重視
- 将来的な設定項目追加に対する拡張性を考慮した設計

## 実装状況
🟠 **PBI-017は部分実装** (2025-09-11)

### 現在の状況
- 基本的な設定管理システムは実装済み（`internal/config/`パッケージ）
- INI形式の設定ファイルサポートは存在
- 基本的な環境変数読み込み機能は実装済み
- サンドボックス機能用の設定システムは一部実装済み

### 未実装の要素
1. **IntegratedConfig 統合システム**
   - IntegratedConfig 構造体と統合設定アーキテクチャ
   - ValidationConfig, ErrorFeedbackConfig, HelpSystemConfig 新設定セクション
   - PerformanceConfig とOutputConfig 拡張設定
   - 設定の自動保存とリアルタイム更新機能

2. **プロファイル管理システム**
   - ProfileManager クラスとProfileConfig 構造体
   - CreateProfile(), SwitchProfile(), ListProfiles() 機能
   - プロファイル継承（based_on）システム
   - プロファイル使用統計とメタデータ管理

3. **環境別設定管理**
   - EnvironmentConfig 構造体と環境切り替え機能
   - 開発・本番・テスト環境の設定管理
   - 環境別APIエンドポイントとタイムアウト設定
   - 環境固有のオーバーライド設定

4. **動的設定更新**
   - ConfigWatcher とConfigChangeEvent システム
   - UpdateSetting() 動的設定変更機能
   - 設定変更イベント通知システム
   - 環境変数オーバーライド機能の拡張

5. **設定CLIコマンド**
   - config profile/env/show/edit/set/get サブコマンド
   - インタラクティブ設定セットアップ
   - 設定ファイルの検証・リセット・バックアップ機能
   - InteractiveConfigManager とConfigUI システム

6. **マイグレーション支援**
   - ConfigMigrator クラスと設定ファイルマイグレーション
   - .envファイルから新形式への移行機能
   - 設定バージョン管理と互換性保持
   - 自動バックアップとロールバック機能

### 部分実装済みの要素
✅ **基本設定システム**: internal/config/ パッケージ
✅ **INIファイルサポート**: 設定ファイル読み込み機能
✅ **環境変数サポート**: 基本的な環境変数読み込み
✅ **サンドボックス設定**: サンドボックス機能用の一部設定

### 次のステップ
1. `internal/config/integrated_config.go` ファイルの作成
2. IntegratedConfig 構造体と新設定セクションの実装
3. ProfileManager とプロファイル管理機能の構築
4. 環境別設定管理システムの実装
5. 動的設定更新とイベント通知システムの実装
6. 設定CLIコマンドとインタラクティブセットアップの作成
7. 設定マイグレーション機能の実装
8. 既存設定システムとの連携と互換性テスト
9. 包括的な設定テストケース作成

### 関連ファイル
- 拡張対象: `internal/config/` パッケージ ✅
- 実装予定: `internal/config/integrated_config.go`
- 実装予定: `internal/config/profile_manager.go`
- 実装予定: `internal/config/environment_manager.go`
- 実装予定: `internal/config/config_watcher.go`
- 実装予定: `internal/config/migrator.go`
- 実装予定: `cmd/usacloud-update/config.go`
- 実装予定: `internal/config/interactive.go`
- テスト連携: `internal/config/config_test.go`