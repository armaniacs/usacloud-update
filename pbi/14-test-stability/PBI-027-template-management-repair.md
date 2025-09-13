# PBI-027: Template管理機能修復

## 概要
TemplateManagerの機能でテンプレート取得・管理に関するテスト失敗が発生しており、プロファイル作成時のテンプレート機能が不安定な状態となっている。テンプレート機能はユーザビリティ向上の重要な機能であり、修復により設定の標準化とユーザ体験の向上を図る。

## 受け入れ条件
- [ ] TemplateManager の全機能が正常に動作すること
- [ ] テンプレートベースのプロファイル作成が正常に機能すること
- [ ] 組み込みテンプレートの取得・一覧表示が正常動作すること
- [ ] Template関連の全テストが通過すること

## 技術仕様

### 現在の問題
```bash
# 予想されるテスト失敗パターン
=== RUN   TestTemplateManager_GetTemplate
    template_test.go:XX: GetTemplate() failed: template not found
--- FAIL: TestTemplateManager_GetTemplate

=== RUN   TestTemplateManager_CreateFromTemplate
    template_test.go:XX: CreateProfileFromTemplate() failed: invalid template
--- FAIL: TestTemplateManager_CreateFromTemplate
```

### 1. TemplateManager構造とインターフェース設計
#### 現在の構造確認・修正が必要な領域
```go
// internal/config/profile/template.go
type TemplateManager struct {
    builtinTemplates []ProfileTemplate
    customTemplates  []ProfileTemplate
    mutex           sync.RWMutex
}

type ProfileTemplate struct {
    Name        string                 `yaml:"name"`
    Description string                 `yaml:"description"`
    Environment string                 `yaml:"environment"`
    ConfigKeys  []ConfigKeyDefinition  `yaml:"config_keys"`
    Tags        []string               `yaml:"tags"`
    Version     string                 `yaml:"version"`
}

type ConfigKeyDefinition struct {
    Key         string `yaml:"key"`
    Type        string `yaml:"type"`          // "string", "int", "bool", "secret"
    Description string `yaml:"description"`
    Required    bool   `yaml:"required"`
    Default     string `yaml:"default"`
    Validation  string `yaml:"validation"`   // regex pattern
}
```

#### TemplateManager機能の完全実装
```go
func NewTemplateManager() *TemplateManager {
    tm := &TemplateManager{
        builtinTemplates: getBuiltinTemplates(),
        customTemplates:  []ProfileTemplate{},
    }
    return tm
}

func (tm *TemplateManager) GetTemplate(name string) (*ProfileTemplate, error) {
    tm.mutex.RLock()
    defer tm.mutex.RUnlock()
    
    // 組み込みテンプレートから検索
    for _, template := range tm.builtinTemplates {
        if template.Name == name {
            return &template, nil
        }
    }
    
    // カスタムテンプレートから検索
    for _, template := range tm.customTemplates {
        if template.Name == name {
            return &template, nil
        }
    }
    
    return nil, fmt.Errorf("template '%s' not found", name)
}

func (tm *TemplateManager) GetAllTemplates() []ProfileTemplate {
    tm.mutex.RLock()
    defer tm.mutex.RUnlock()
    
    var allTemplates []ProfileTemplate
    allTemplates = append(allTemplates, tm.builtinTemplates...)
    allTemplates = append(allTemplates, tm.customTemplates...)
    
    return allTemplates
}

func (tm *TemplateManager) GetTemplateByEnvironment(environment string) []ProfileTemplate {
    tm.mutex.RLock()
    defer tm.mutex.RUnlock()
    
    var filtered []ProfileTemplate
    for _, template := range tm.GetAllTemplates() {
        if template.Environment == environment || template.Environment == "any" {
            filtered = append(filtered, template)
        }
    }
    
    return filtered
}
```

### 2. 組み込みテンプレートの定義
#### 実用的なテンプレート例の実装
```go
func getBuiltinTemplates() []ProfileTemplate {
    return []ProfileTemplate{
        {
            Name:        "development",
            Description: "開発環境用の基本設定",
            Environment: "development",
            ConfigKeys: []ConfigKeyDefinition{
                {
                    Key:         "SAKURACLOUD_ACCESS_TOKEN",
                    Type:        "secret",
                    Description: "Sakura Cloud アクセストークン",
                    Required:    true,
                    Validation:  "^[a-f0-9-]+$",
                },
                {
                    Key:         "SAKURACLOUD_ACCESS_TOKEN_SECRET",
                    Type:        "secret",
                    Description: "Sakura Cloud アクセストークンシークレット",
                    Required:    true,
                    Validation:  "^[A-Za-z0-9+/=]+$",
                },
                {
                    Key:         "SAKURACLOUD_ZONE",
                    Type:        "string",
                    Description: "デフォルトゾーン",
                    Required:    false,
                    Default:     "tk1v",
                    Validation:  "^(tk1v|is1a|is1b|tk1a)$",
                },
            },
            Tags: []string{"development", "default"},
            Version: "1.0",
        },
        {
            Name:        "production",
            Description: "本番環境用の設定",
            Environment: "production",
            ConfigKeys: []ConfigKeyDefinition{
                {
                    Key:         "SAKURACLOUD_ACCESS_TOKEN",
                    Type:        "secret",
                    Description: "Sakura Cloud アクセストークン（本番用）",
                    Required:    true,
                    Validation:  "^[a-f0-9-]+$",
                },
                {
                    Key:         "SAKURACLOUD_ACCESS_TOKEN_SECRET",
                    Type:        "secret",
                    Description: "Sakura Cloud アクセストークンシークレット（本番用）",
                    Required:    true,
                    Validation:  "^[A-Za-z0-9+/=]+$",
                },
                {
                    Key:         "SAKURACLOUD_ZONE",
                    Type:        "string",
                    Description: "本番環境のゾーン",
                    Required:    true,
                    Validation:  "^(tk1v|is1a|is1b|tk1a)$",
                },
                {
                    Key:         "TIMEOUT",
                    Type:        "int",
                    Description: "タイムアウト値（秒）",
                    Required:    false,
                    Default:     "300",
                    Validation:  "^[0-9]+$",
                },
            },
            Tags: []string{"production", "secure"},
            Version: "1.0",
        },
    }
}
```

### 3. テンプレートからのプロファイル作成機能
#### CreateProfileFromTemplate実装
```go
func (tm *TemplateManager) CreateProfileFromTemplate(templateName, profileName string, config map[string]string) (*Profile, error) {
    template, err := tm.GetTemplate(templateName)
    if err != nil {
        return nil, fmt.Errorf("template not found: %w", err)
    }
    
    // テンプレートの設定キー検証
    if err := tm.validateConfigAgainstTemplate(template, config); err != nil {
        return nil, fmt.Errorf("config validation failed: %w", err)
    }
    
    // デフォルト値の適用
    finalConfig := tm.applyTemplateDefaults(template, config)
    
    profile := &Profile{
        ID:          generateID(),
        Name:        profileName,
        Description: fmt.Sprintf("Created from template: %s", template.Description),
        Environment: template.Environment,
        Config:      finalConfig,
        Tags:        append([]string{}, template.Tags...),
        CreatedAt:   time.Now(),
        UpdatedAt:   time.Now(),
    }
    
    return profile, nil
}

func (tm *TemplateManager) validateConfigAgainstTemplate(template *ProfileTemplate, config map[string]string) error {
    for _, keyDef := range template.ConfigKeys {
        value, exists := config[keyDef.Key]
        
        // 必須キーのチェック
        if keyDef.Required && !exists {
            return fmt.Errorf("required key '%s' is missing", keyDef.Key)
        }
        
        if exists && keyDef.Validation != "" {
            // バリデーションパターンのチェック
            matched, err := regexp.MatchString(keyDef.Validation, value)
            if err != nil {
                return fmt.Errorf("invalid validation pattern for key '%s': %w", keyDef.Key, err)
            }
            if !matched {
                return fmt.Errorf("value for key '%s' does not match required pattern", keyDef.Key)
            }
        }
    }
    
    return nil
}

func (tm *TemplateManager) applyTemplateDefaults(template *ProfileTemplate, config map[string]string) map[string]string {
    result := make(map[string]string)
    
    // 既存の設定をコピー
    for k, v := range config {
        result[k] = v
    }
    
    // デフォルト値を適用（既存の値がない場合のみ）
    for _, keyDef := range template.ConfigKeys {
        if _, exists := result[keyDef.Key]; !exists && keyDef.Default != "" {
            result[keyDef.Key] = keyDef.Default
        }
    }
    
    return result
}
```

### 4. 包括的テスト実装
#### テンプレート機能のテスト
```go
func TestTemplateManager_GetTemplate(t *testing.T) {
    tm := NewTemplateManager()
    
    // 存在するテンプレートの取得テスト
    template, err := tm.GetTemplate("development")
    if err != nil {
        t.Fatalf("GetTemplate() failed: %v", err)
    }
    
    if template.Name != "development" {
        t.Errorf("Expected template name 'development', got %s", template.Name)
    }
    
    // 存在しないテンプレートのテスト
    _, err = tm.GetTemplate("nonexistent")
    if err == nil {
        t.Errorf("Expected error for nonexistent template")
    }
}

func TestTemplateManager_CreateFromTemplate(t *testing.T) {
    tm := NewTemplateManager()
    
    config := map[string]string{
        "SAKURACLOUD_ACCESS_TOKEN":        "test-token-123",
        "SAKURACLOUD_ACCESS_TOKEN_SECRET": "test-secret-456",
    }
    
    profile, err := tm.CreateProfileFromTemplate("development", "Test Profile", config)
    if err != nil {
        t.Fatalf("CreateProfileFromTemplate() failed: %v", err)
    }
    
    if profile.Name != "Test Profile" {
        t.Errorf("Expected profile name 'Test Profile', got %s", profile.Name)
    }
    
    if profile.Environment != "development" {
        t.Errorf("Expected environment 'development', got %s", profile.Environment)
    }
    
    // デフォルト値の適用確認
    if profile.Config["SAKURACLOUD_ZONE"] != "tk1v" {
        t.Errorf("Expected default zone 'tk1v', got %s", profile.Config["SAKURACLOUD_ZONE"])
    }
}
```

## テスト戦略
- **ユニットテスト**: TemplateManager各機能の個別テスト
- **統合テスト**: プロファイル作成フローでのテンプレート使用テスト
- **バリデーションテスト**: 設定値検証・エラーハンドリングテスト

## 依存関係
- 前提PBI: なし（独立した修復タスク）
- 関連PBI: PBI-026（Profile FileStorage機能修復）
- 既存コード: internal/config/profile/ パッケージ

## 見積もり
- 開発工数: 5時間
  - TemplateManager完全実装: 2時間
  - 組み込みテンプレート定義: 1時間
  - テンプレート作成機能実装: 1.5時間
  - テスト実装: 0.5時間

## 完了の定義
- [ ] 全TemplateManager関連テストが通過
- [ ] テンプレートベースのプロファイル作成が正常動作
- [ ] 組み込みテンプレートが適切に提供される
- [ ] 設定値バリデーションが正常機能
- [ ] ユーザビリティテスト完了

## 備考
- テンプレート機能はユーザビリティ向上の重要な機能
- 将来的なカスタムテンプレート機能拡張の基盤となる
- 設定の標準化により、ユーザの設定ミスを防止できる

---

## 実装状況 (2025-09-11)

🟠 **PBI-027は未実装** (2025-09-11)

### 現在の状況
- TemplateManagerの機能でテンプレート取得・管理に関するテスト失敗が発生
- プロファイル作成時のテンプレート機能が不安定
- テンプレートベースのプロファイル作成が動作不能
- 組み込みテンプレートの取得・一覧表示に問題

### 未実装要素
1. **TemplateManager完全実装**
   - TemplateManager構造とインターフェース設計の確立
   - ProfileTemplate構造体の定義と実装
   - GetTemplate、CreateFromTemplate機能の実装
   - テンプレート取得・管理機能の確立

2. **組み込みテンプレート定義**
   - 標準テンプレートの定義と実装
   - 環境別テンプレートの提供
   - ConfigKeyDefinitionの定義と実装
   - テンプレート一覧表示機能

3. **テンプレート作成機能実装**
   - テンプレートベースのプロファイル作成機能
   - 設定値バリデーション機能
   - ユーザビリティテストの実装
   - カスタムテンプレート機能拡張の基盤作り

### 次のステップ
1. TemplateManager構造とインターフェースの完全実装
2. 組み込みテンプレートの定義と実装
3. テンプレートベースのプロファイル作成機能の実装
4. TemplateManager関連テストの実装と検証
5. ユーザビリティテストと設定値バリデーションの実装

### 技術要件
- Go 1.24.1対応
- YAMLライブラリ (gopkg.in/yaml.v3)の活用
- テンプレート管理とプロファイル作成機能の完全実装
- 5時間の作業見積もり

### 受け入れ条件の進捗
- [ ] TemplateManager の全機能が正常に動作すること
- [ ] テンプレートベースのプロファイル作成が正常に機能すること
- [ ] 組み込みテンプレートの取得・一覧表示が正常動作すること
- [ ] Template関連の全テストが通過すること