# PBI-026: Profile FileStorage機能修復

## 概要
ProfileManagerのファイルストレージ機能で複数のテスト失敗が発生しており、プロファイルデータの永続化・読み込み機能に重大な問題がある。特にYAMLシリアライゼーションエラーやファイル権限問題により、プロファイル管理の基幹機能が不安定な状態となっている。

## 受け入れ条件
- [ ] プロファイルの保存・読み込みが正常に動作すること
- [ ] YAMLシリアライゼーション/デシリアライゼーションが正常に機能すること
- [ ] ファイル権限の適切な設定・検証が行われること
- [ ] 全てのProfileManager関連テストが通過すること

## 技術仕様

### 現在の問題
```bash
# 確認されるテスト失敗パターン
=== RUN   TestProfileManager_FileOperations
    manager_test.go:XXX: Failed to save profile: yaml: ...
--- FAIL: TestProfileManager_FileOperations

=== RUN   TestProfileManager_LoadProfiles
    manager_test.go:XXX: Failed to load profiles: permission denied
--- FAIL: TestProfileManager_LoadProfiles
```

### 1. YAMLシリアライゼーション修正
#### 現在の問題調査が必要な領域
```go
// internal/config/profile/manager.go のProfile構造体確認
type Profile struct {
    ID          string            `yaml:"id"`
    Name        string            `yaml:"name"`
    Description string            `yaml:"description"`
    Environment string            `yaml:"environment"`
    Config      map[string]string `yaml:"config"`
    Tags        []string          `yaml:"tags"`
    ParentID    string            `yaml:"parent_id,omitempty"`
    IsDefault   bool              `yaml:"is_default"`
    CreatedAt   time.Time         `yaml:"created_at"`
    UpdatedAt   time.Time         `yaml:"updated_at"`
    LastUsedAt  time.Time         `yaml:"last_used_at"`
}
```

#### YAML処理の堅牢化
```go
// 安全なYAMLマーシャリング
func (pm *ProfileManager) saveProfile(profile *Profile) error {
    data, err := yaml.Marshal(profile)
    if err != nil {
        return fmt.Errorf("YAML marshal error: %w", err)
    }
    
    filePath := filepath.Join(pm.configDir, profile.ID+".yaml")
    
    // 原子的書き込みによるデータ整合性保証
    tempFile := filePath + ".tmp"
    if err := ioutil.WriteFile(tempFile, data, 0600); err != nil {
        return fmt.Errorf("failed to write temp file: %w", err)
    }
    
    if err := os.Rename(tempFile, filePath); err != nil {
        os.Remove(tempFile) // クリーンアップ
        return fmt.Errorf("failed to move temp file: %w", err)
    }
    
    return nil
}

// エラーハンドリング強化したYAMLアンマーシャリング
func (pm *ProfileManager) loadProfile(filePath string) (*Profile, error) {
    data, err := ioutil.ReadFile(filePath)
    if err != nil {
        return nil, fmt.Errorf("failed to read profile file %s: %w", filePath, err)
    }
    
    var profile Profile
    if err := yaml.Unmarshal(data, &profile); err != nil {
        return nil, fmt.Errorf("YAML unmarshal error for %s: %w", filePath, err)
    }
    
    // バリデーション
    if err := pm.validateProfile(&profile); err != nil {
        return nil, fmt.Errorf("profile validation failed for %s: %w", filePath, err)
    }
    
    return &profile, nil
}
```

### 2. ファイル権限・ディレクトリ管理修正
#### ディレクトリ初期化の堅牢化
```go
func NewProfileManager(configDir string) (*ProfileManager, error) {
    // ディレクトリ存在確認・作成
    if err := os.MkdirAll(configDir, 0700); err != nil {
        return nil, fmt.Errorf("failed to create config directory: %w", err)
    }
    
    // ディレクトリ権限確認
    info, err := os.Stat(configDir)
    if err != nil {
        return nil, fmt.Errorf("failed to stat config directory: %w", err)
    }
    
    if !info.IsDir() {
        return nil, fmt.Errorf("config path is not a directory: %s", configDir)
    }
    
    // 権限チェック（Unix系のみ）
    if runtime.GOOS != "windows" {
        if info.Mode().Perm() != 0700 {
            if err := os.Chmod(configDir, 0700); err != nil {
                return nil, fmt.Errorf("failed to set directory permissions: %w", err)
            }
        }
    }
    
    pm := &ProfileManager{
        configDir: configDir,
        profiles:  make(map[string]*Profile),
    }
    
    // 既存プロファイルの読み込み
    if err := pm.loadExistingProfiles(); err != nil {
        return nil, fmt.Errorf("failed to load existing profiles: %w", err)
    }
    
    return pm, nil
}
```

### 3. テスト修正・強化
#### ファイルシステムテストの改善
```go
func TestProfileManager_FileOperations(t *testing.T) {
    tempDir := t.TempDir()
    
    // 権限確認可能なテスト環境の構築
    if runtime.GOOS != "windows" {
        err := os.Chmod(tempDir, 0700)
        if err != nil {
            t.Fatalf("Failed to set temp dir permissions: %v", err)
        }
    }
    
    manager, err := NewProfileManager(tempDir)
    if err != nil {
        t.Fatalf("NewProfileManager() failed: %v", err)
    }
    
    // プロファイル作成・保存テスト
    profile, err := manager.CreateProfile(ProfileCreateOptions{
        Name:        "Test Profile",
        Description: "Test description",
        Environment: "test",
        Config: map[string]string{
            "key1": "value1",
            "SAKURACLOUD_ACCESS_TOKEN": "test-token",
            "SAKURACLOUD_ACCESS_TOKEN_SECRET": "test-secret",
        },
    })
    if err != nil {
        t.Fatalf("CreateProfile() failed: %v", err)
    }
    
    // ファイル存在確認
    profileFile := filepath.Join(tempDir, profile.ID+".yaml")
    if _, err := os.Stat(profileFile); os.IsNotExist(err) {
        t.Errorf("Profile file was not created: %s", profileFile)
    }
    
    // ファイル権限確認
    if runtime.GOOS != "windows" {
        info, err := os.Stat(profileFile)
        if err != nil {
            t.Fatalf("Failed to stat profile file: %v", err)
        }
        
        if info.Mode().Perm() != 0600 {
            t.Errorf("Expected file permissions 0600, got %o", info.Mode().Perm())
        }
    }
    
    // 再読み込みテスト
    manager2, err := NewProfileManager(tempDir)
    if err != nil {
        t.Fatalf("Failed to reload ProfileManager: %v", err)
    }
    
    loadedProfile, err := manager2.GetProfile(profile.ID)
    if err != nil {
        t.Fatalf("Failed to get reloaded profile: %v", err)
    }
    
    if loadedProfile.Name != profile.Name {
        t.Errorf("Profile name mismatch after reload: expected %s, got %s", 
            profile.Name, loadedProfile.Name)
    }
}
```

## テスト戦略
- **ユニットテスト**: Profile CRUD操作の個別テスト
- **ファイルシステムテスト**: 権限・読み書き操作の確認
- **エラーハンドリングテスト**: 破損ファイル・権限不足時の動作確認
- **パフォーマンステスト**: 大量プロファイル処理時の性能確認

## 依存関係
- 前提PBI: なし（独立した修復タスク）
- 関連PBI: PBI-025（Export/Import機能修復）、PBI-027（Template管理機能修復）
- 既存コード: internal/config/profile/ パッケージ全体

## 見積もり
- 開発工数: 6時間
  - YAML処理修正: 2時間
  - ファイル権限・ディレクトリ管理修正: 2時間
  - テスト修正・強化: 2時間

## 完了の定義
- [ ] 全ProfileManager関連テストが通過
- [ ] プロファイルの保存・読み込み機能が安定動作
- [ ] ファイル権限が適切に設定・管理される
- [ ] YAML処理エラーの完全解消
- [ ] エラーハンドリングの堅牢性確認

## 備考
- プロファイル管理はusacloud-updateの基幹機能のため、最高レベルの安定性が要求される
- セキュリティ面でファイル権限の適切な管理は必須
- 原子的書き込みによるデータ整合性保証を実装し、システム障害時の安全性を確保

---

## 実装状況 (2025-09-11)

🟠 **PBI-026は未実装** (2025-09-11)

### 現在の状況
- ProfileManagerのファイルストレージ機能で複数のテスト失敗が発生
- YAMLシリアライゼーションエラーが発生している
- ファイル権限問題によりプロファイル管理機能が不安定
- プロファイルデータの永続化・読み込み機能に重大な問題

### 未実装要素
1. **YAMLシリアライゼーション修正**
   - Profile構造体のYAMLタグ設定の確認と修正
   - time.Timeフィールドのシリアライゼーション問題解決
   - シリアライゼーションエラーの完全解消

2. **ファイル権限・ディレクトリ管理修正**
   - ファイル権限の適切な設定・検証
   - ディレクトリ作成権限の確認と管理
   - permission deniedエラーの解決
   - 原子的書き込みの実装

3. **テスト修正・強化**
   - TestProfileManager_FileOperationsの修正
   - TestProfileManager_LoadProfilesの修正
   - ファイルシステムテストの実装
   - エラーハンドリングテストの実装
   - パフォーマンステストの実装

### 次のステップ
1. Profile構造体のYAMLタグ設定確認と修正
2. YAMLシリアライゼーションエラーの解決
3. ファイル権限・ディレクトリ管理機能の修正
4. ProfileManager関連テストの修正と強化
5. パフォーマンステストとエラーハンドリングテストの実装

### 技術要件
- Go 1.24.1対応
- YAMLライブラリ (gopkg.in/yaml.v3)の適切な利用
- ファイルシステムの安全な操作と権限管理
- 6時間の作業見積もり

### 受け入れ条件の進捗
- [ ] プロファイルの保存・読み込みが正常に動作すること
- [ ] YAMLシリアライゼーション/デシリアライゼーションが正常に機能すること
- [ ] ファイル権限の適切な設定・検証が行われること
- [ ] 全てのProfileManager関連テストが通過すること