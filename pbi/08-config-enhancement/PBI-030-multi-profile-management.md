# PBI-030: マルチプロファイル管理機能

## 概要
複数のSakura Cloud環境（本番、ステージング、開発等）や異なるプロジェクト用の設定プロファイルを管理できる機能を実装します。プロファイルの切り替え、継承、バックアップ機能を含む包括的なプロファイル管理システムを提供します。

## 受け入れ条件
- [ ] 複数の設定プロファイルを作成・管理できる
- [ ] プロファイル間の素早い切り替えができる
- [ ] プロファイルの継承・テンプレート機能を提供する
- [ ] プロファイルの一覧表示・詳細確認ができる
- [ ] プロファイルのエクスポート・インポート機能を提供する

## 技術仕様

### 1. プロファイル データ構造
```go
type Profile struct {
    ID            string            `json:"id" yaml:"id"`
    Name          string            `json:"name" yaml:"name"`
    Description   string            `json:"description" yaml:"description"`
    Environment   string            `json:"environment" yaml:"environment"` // prod, staging, dev
    ParentID      string            `json:"parent_id,omitempty" yaml:"parent_id,omitempty"`
    Config        map[string]string `json:"config" yaml:"config"`
    CreatedAt     time.Time         `json:"created_at" yaml:"created_at"`
    UpdatedAt     time.Time         `json:"updated_at" yaml:"updated_at"`
    LastUsedAt    time.Time         `json:"last_used_at" yaml:"last_used_at"`
    Tags          []string          `json:"tags" yaml:"tags"`
    IsDefault     bool              `json:"is_default" yaml:"is_default"`
}

type ProfileManager struct {
    profiles      map[string]*Profile
    activeProfile *Profile
    configDir     string
    storage       ProfileStorage
}

type ProfileStorage interface {
    Save(profile *Profile) error
    Load(id string) (*Profile, error)
    LoadAll() (map[string]*Profile, error)
    Delete(id string) error
    SetDefault(id string) error
    GetDefault() (*Profile, error)
}
```

### 2. プロファイル管理機能
```go
func (pm *ProfileManager) CreateProfile(name, description, environment string, config map[string]string) (*Profile, error) {
    profile := &Profile{
        ID:          generateProfileID(),
        Name:        name,
        Description: description,
        Environment: environment,
        Config:      config,
        CreatedAt:   time.Now(),
        UpdatedAt:   time.Now(),
        Tags:        []string{},
    }
    
    // 既存プロファイル名の重複チェック
    if pm.profileExists(name) {
        return nil, fmt.Errorf("profile with name '%s' already exists", name)
    }
    
    // バリデーション
    if err := pm.validateProfile(profile); err != nil {
        return nil, fmt.Errorf("profile validation failed: %w", err)
    }
    
    pm.profiles[profile.ID] = profile
    return profile, pm.storage.Save(profile)
}

func (pm *ProfileManager) SwitchProfile(id string) error {
    profile, exists := pm.profiles[id]
    if !exists {
        return fmt.Errorf("profile not found: %s", id)
    }
    
    // 現在のプロファイルをバックアップ
    if pm.activeProfile != nil {
        pm.activeProfile.LastUsedAt = time.Now()
        pm.storage.Save(pm.activeProfile)
    }
    
    // 新しいプロファイルをアクティベート
    pm.activeProfile = profile
    profile.LastUsedAt = time.Now()
    
    // 環境変数に設定を適用
    return pm.applyProfileToEnvironment(profile)
}

func (pm *ProfileManager) applyProfileToEnvironment(profile *Profile) error {
    for key, value := range profile.Config {
        if err := os.Setenv(key, value); err != nil {
            return fmt.Errorf("failed to set environment variable %s: %w", key, err)
        }
    }
    
    // 設定ファイルにも反映
    configFile := filepath.Join(pm.configDir, "current.conf")
    return pm.writeConfigFile(configFile, profile.Config)
}
```

### 3. プロファイル継承機能
```go
func (pm *ProfileManager) CreateProfileFromParent(name, description string, parentID string, overrides map[string]string) (*Profile, error) {
    parent, exists := pm.profiles[parentID]
    if !exists {
        return nil, fmt.Errorf("parent profile not found: %s", parentID)
    }
    
    // 親の設定をベースにする
    config := make(map[string]string)
    for k, v := range parent.Config {
        config[k] = v
    }
    
    // オーバーライド設定を適用
    for k, v := range overrides {
        config[k] = v
    }
    
    profile := &Profile{
        ID:          generateProfileID(),
        Name:        name,
        Description: description,
        Environment: parent.Environment, // 親と同じ環境をデフォルト
        ParentID:    parentID,
        Config:      config,
        CreatedAt:   time.Now(),
        UpdatedAt:   time.Now(),
        Tags:        append([]string{}, parent.Tags...), // 親のタグを継承
    }
    
    pm.profiles[profile.ID] = profile
    return profile, pm.storage.Save(profile)
}

func (pm *ProfileManager) UpdateProfile(id string, updates map[string]string) error {
    profile, exists := pm.profiles[id]
    if !exists {
        return fmt.Errorf("profile not found: %s", id)
    }
    
    // 設定を更新
    for key, value := range updates {
        if value == "" {
            delete(profile.Config, key)
        } else {
            profile.Config[key] = value
        }
    }
    
    profile.UpdatedAt = time.Now()
    
    // 子プロファイルがある場合は継承関係をチェック
    if err := pm.updateChildProfiles(profile); err != nil {
        return fmt.Errorf("failed to update child profiles: %w", err)
    }
    
    return pm.storage.Save(profile)
}
```

### 4. プロファイル テンプレート機能
```go
type ProfileTemplate struct {
    Name        string            `json:"name" yaml:"name"`
    Description string            `json:"description" yaml:"description"`
    Environment string            `json:"environment" yaml:"environment"`
    ConfigKeys  []ConfigKeyDef    `json:"config_keys" yaml:"config_keys"`
    Tags        []string          `json:"tags" yaml:"tags"`
}

type ConfigKeyDef struct {
    Key         string `json:"key" yaml:"key"`
    Description string `json:"description" yaml:"description"`
    Required    bool   `json:"required" yaml:"required"`
    Default     string `json:"default" yaml:"default"`
    Type        string `json:"type" yaml:"type"` // string, int, bool, url
    Validation  string `json:"validation,omitempty" yaml:"validation,omitempty"` // regex pattern
}

func (pm *ProfileManager) GetBuiltinTemplates() []ProfileTemplate {
    return []ProfileTemplate{
        {
            Name:        "Sakura Cloud 本番環境",
            Description: "本番環境用の標準設定テンプレート",
            Environment: "production",
            ConfigKeys: []ConfigKeyDef{
                {
                    Key:         "SAKURACLOUD_ACCESS_TOKEN",
                    Description: "SakuraCloud APIアクセストークン",
                    Required:    true,
                    Type:        "string",
                },
                {
                    Key:         "SAKURACLOUD_ACCESS_TOKEN_SECRET",
                    Description: "SakuraCloud APIアクセストークンシークレット",
                    Required:    true,
                    Type:        "string",
                },
                {
                    Key:         "SAKURACLOUD_ZONE",
                    Description: "操作対象ゾーン",
                    Required:    true,
                    Default:     "tk1v",
                    Type:        "string",
                },
            },
            Tags: []string{"production", "sakura-cloud"},
        },
        {
            Name:        "開発・検証環境",
            Description: "開発・検証用の安全な設定テンプレート",
            Environment: "development",
            ConfigKeys: []ConfigKeyDef{
                {
                    Key:         "SAKURACLOUD_ACCESS_TOKEN",
                    Description: "開発用APIアクセストークン",
                    Required:    true,
                    Type:        "string",
                },
                {
                    Key:         "SAKURACLOUD_ACCESS_TOKEN_SECRET",
                    Description: "開発用APIアクセストークンシークレット",
                    Required:    true,
                    Type:        "string",
                },
                {
                    Key:         "SAKURACLOUD_ZONE",
                    Description: "検証用ゾーン（tk1v推奨）",
                    Required:    true,
                    Default:     "tk1v",
                    Type:        "string",
                },
                {
                    Key:         "USACLOUD_UPDATE_DRY_RUN",
                    Description: "デフォルトでドライラン実行",
                    Required:    false,
                    Default:     "true",
                    Type:        "bool",
                },
            },
            Tags: []string{"development", "safe"},
        },
    }
}
```

### 5. CLIコマンド実装
```go
type ProfileCommand struct {
    manager *ProfileManager
}

func (pc *ProfileCommand) ListProfiles(cmd *cobra.Command, args []string) error {
    profiles := pc.manager.ListProfiles()
    
    table := tablewriter.NewWriter(os.Stdout)
    table.SetHeader([]string{"ID", "名前", "環境", "最終使用", "デフォルト"})
    
    for _, profile := range profiles {
        isDefault := ""
        if profile.IsDefault {
            isDefault = "✓"
        }
        
        lastUsed := "未使用"
        if !profile.LastUsedAt.IsZero() {
            lastUsed = profile.LastUsedAt.Format("2006-01-02 15:04")
        }
        
        table.Append([]string{
            profile.ID[:8],
            profile.Name,
            profile.Environment,
            lastUsed,
            isDefault,
        })
    }
    
    table.Render()
    return nil
}

func (pc *ProfileCommand) ShowProfile(cmd *cobra.Command, args []string) error {
    if len(args) == 0 {
        return fmt.Errorf("profile ID or name required")
    }
    
    profile, err := pc.manager.GetProfile(args[0])
    if err != nil {
        return err
    }
    
    fmt.Printf("プロファイル詳細\n")
    fmt.Printf("================\n")
    fmt.Printf("ID: %s\n", profile.ID)
    fmt.Printf("名前: %s\n", profile.Name)
    fmt.Printf("説明: %s\n", profile.Description)
    fmt.Printf("環境: %s\n", profile.Environment)
    
    if profile.ParentID != "" {
        parent, _ := pc.manager.GetProfile(profile.ParentID)
        if parent != nil {
            fmt.Printf("継承元: %s\n", parent.Name)
        }
    }
    
    fmt.Printf("作成日時: %s\n", profile.CreatedAt.Format("2006-01-02 15:04:05"))
    fmt.Printf("更新日時: %s\n", profile.UpdatedAt.Format("2006-01-02 15:04:05"))
    
    if !profile.LastUsedAt.IsZero() {
        fmt.Printf("最終使用: %s\n", profile.LastUsedAt.Format("2006-01-02 15:04:05"))
    }
    
    if len(profile.Tags) > 0 {
        fmt.Printf("タグ: %s\n", strings.Join(profile.Tags, ", "))
    }
    
    fmt.Printf("\n設定項目:\n")
    for key, value := range profile.Config {
        // 機密情報はマスクして表示
        if isSensitiveKey(key) {
            fmt.Printf("  %s: %s\n", key, maskValue(value))
        } else {
            fmt.Printf("  %s: %s\n", key, value)
        }
    }
    
    return nil
}
```

## テスト戦略
- **プロファイル管理テスト**: CRUD操作の正確性確認
- **継承機能テスト**: 親子関係とオーバーライドの動作確認
- **切り替え機能テスト**: 環境変数適用の正確性確認
- **テンプレート機能テスト**: テンプレートからの生成確認

## 依存関係
- 前提PBI: なし（既存設定管理を拡張）
- 関連PBI: PBI-031（設定バリデーション）、PBI-032（設定暗号化）
- 既存コード: internal/config/

## 見積もり
- 開発工数: 14時間
  - プロファイル データ構造設計: 2時間
  - プロファイル管理機能実装: 6時間
  - 継承・テンプレート機能実装: 4時間
  - CLIコマンド実装: 2時間

## 完了の定義
- [ ] 複数プロファイルの作成・管理ができる
- [ ] プロファイル切り替えが正確に動作する
- [ ] 継承機能が期待通りに動作する
- [ ] テンプレートからの生成が正しく機能する
- [ ] エクスポート・インポート機能が動作する

## 備考
- 設定ファイルはYAML形式で保存（可読性重視）
- 機密情報（APIキー等）は適切にマスク表示
- デフォルトプロファイルの自動切り替え機能も提供

## 実装状況
❌ **PBI-030は未実装** (2025-09-11)

### 現在の状況
- マルチプロファイル管理機能は未実装
- 複数の設定プロファイル作成・管理機能なし
- プロファイル間の切り替え機能なし
- プロファイル継承・テンプレート機能なし
- エクスポート・インポート機能なし

### 実装すべき要素
1. **プロファイルデータ構造**
   - Profile 構造体とProfileManager クラスの実装
   - ProfileStorage インターフェースとYAMLストレージ実装
   - プロファイルメタデータ管理（作成日時、最終使用日時等）
   - タグ・環境別分類機能

2. **プロファイル管理機能**
   - CRUD操作（作成・読み込み・更新・削除）
   - プロファイル名重複チェック機能
   - プロファイルバリデーションシステム
   - アクティブプロファイル管理と環境変数適用

3. **プロファイル継承機能**
   - 親子関係管理システム
   - 設定オーバーライド機能
   - 子プロファイルへの変更伝播システム
   - 継承チェーンの整合性チェック

4. **テンプレートシステム**
   - ProfileTemplate とConfigKeyDef 構造体の実装
   - 組み込みテンプレート（本番・開発環境用）
   - カスタムテンプレート作成機能
   - テンプレートベースのプロファイル生成

5. **CLIコマンドインターフェース**
   - profile list/show/create/delete/switch サブコマンド
   - プロファイル情報の表形式表示
   - 機密情報のマスク表示機能
   - インタラクティブプロファイル作成ウィザード

6. **エクスポート・インポート機能**
   - YAML形式プロファイルエクスポート
   - プロファイルインポートとバリデーション
   - バックアップ・リストア機能
   - プロファイル移行サポート

### 次のステップ
1. `internal/profile/` パッケージの作成
2. Profile、ProfileManager データ構造の定義と実装
3. YAMLベースのProfileStorage実装
4. プロファイル継承システムの構築
5. 組み込みテンプレートシステムの作成
6. CLIコマンドインターフェースの実装
7. エクスポート・インポート機能の作成
8. 既存設定システムとの統合
9. 包括的なテストケース作成

### 実装状況

**📊 実装状況: 未実装**

#### 実装延期の判断理由
本機能は複雑なマルチプロファイル管理システムの新規実装を含む大規模な機能拡張です。現在のプロジェクトの優先順位として、既存システムの安定性確保とバグ修正を最優先としており、新機能開発は一時的に延期します。

#### 延期期間
- **延期期間**: 次期メジャーリリース（v2.0以降）まで延期
- **再評価時期**: 現在の安定化作業完了後（推定：2025年Q2以降）

#### 現在の状況
- ✅ 仕様策定完了
- ❌ 実装未開始
- ❌ テスト未作成
- ❌ ドキュメント未作成

#### 実装時の考慮点
1. 既存設定システムとの統合複雑性
2. YAMLフォーマットでの設定保存の新規実装が必要
3. プロファイル継承システムの複雑なロジック
4. CLIコマンドインターフェースの大幅な拡張
5. テンプレートシステムの新規開発が必要
6. エクスポート・インポート機能の複雑性

### 関連ファイル
- 実装予定: `internal/profile/manager.go`
- 実装予定: `internal/profile/storage.go`
- 実装予定: `internal/profile/template.go`
- 実装予定: `internal/profile/inheritance.go`
- 実装予定: `cmd/usacloud-update/profile.go`
- 統合対象: `internal/config/manager.go`
- 設定連携: `cmd/usacloud-update/main.go`