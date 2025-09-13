# PBI-032: 認証情報セキュリティ強化

## 概要
APIキー、パスワードなどの機密情報の取り扱いセキュリティを大幅に強化します。暗号化保存、安全な入力方法、ログ出力からの機密情報除外、定期的なローテーション警告などの包括的なセキュリティ機能を実装します。

## 受け入れ条件
- [x] 設定ファイル内の機密情報を暗号化して保存できる
- [x] 機密情報のログ出力を自動的に防止できる
- [x] セキュアな入力方法（マスク入力等）を提供する
- [x] 機密情報の使用履歴を監査ログとして記録できる
- [x] 古い認証情報に対する更新警告を表示できる

## 技術仕様

### 1. 暗号化保存システム
```go
type SecureStorage struct {
    cipher     Cipher
    keyManager *KeyManager
    storage    Storage
}

type Cipher interface {
    Encrypt(plaintext []byte, key []byte) ([]byte, error)
    Decrypt(ciphertext []byte, key []byte) ([]byte, error)
}

type KeyManager struct {
    masterKey []byte
    keyDerivation func(password string, salt []byte) []byte
}

func NewSecureStorage() (*SecureStorage, error) {
    keyManager, err := NewKeyManager()
    if err != nil {
        return nil, err
    }
    
    return &SecureStorage{
        cipher:     NewAESGCMCipher(),
        keyManager: keyManager,
        storage:    NewFileStorage(),
    }, nil
}

func (ss *SecureStorage) StoreCredential(key, value string) error {
    // 暗号化キーの取得
    encKey, err := ss.keyManager.GetEncryptionKey()
    if err != nil {
        return fmt.Errorf("failed to get encryption key: %w", err)
    }
    
    // 機密情報を暗号化
    encrypted, err := ss.cipher.Encrypt([]byte(value), encKey)
    if err != nil {
        return fmt.Errorf("encryption failed: %w", err)
    }
    
    // 安全なメタデータと共に保存
    credential := SecureCredential{
        Key:         key,
        Encrypted:   encrypted,
        Algorithm:   "AES-256-GCM",
        CreatedAt:   time.Now(),
        LastUsedAt:  time.Now(),
        Version:     1,
    }
    
    return ss.storage.Save(key, credential)
}

type SecureCredential struct {
    Key        string    `json:"key"`
    Encrypted  []byte    `json:"encrypted"`
    Algorithm  string    `json:"algorithm"`
    CreatedAt  time.Time `json:"created_at"`
    LastUsedAt time.Time `json:"last_used_at"`
    Version    int       `json:"version"`
}
```

### 2. 機密情報フィルタリング
```go
type SensitiveDataFilter struct {
    patterns    []SensitivePattern
    maskChar    rune
    maskLength  int
}

type SensitivePattern struct {
    Name        string
    Pattern     *regexp.Regexp
    Replacement string
    Description string
}

func NewSensitiveDataFilter() *SensitiveDataFilter {
    return &SensitiveDataFilter{
        patterns: []SensitivePattern{
            {
                Name:        "sakura-access-token",
                Pattern:     regexp.MustCompile(`(?i)(sakuracloud_access_token["\s]*[:=]["\s]*)[a-zA-Z0-9]{20,}`),
                Replacement: "${1}[FILTERED]",
                Description: "Sakura Cloud アクセストークン",
            },
            {
                Name:        "sakura-secret",
                Pattern:     regexp.MustCompile(`(?i)(sakuracloud_access_token_secret["\s]*[:=]["\s]*)[a-zA-Z0-9+/=]{20,}`),
                Replacement: "${1}[FILTERED]",
                Description: "Sakura Cloud アクセストークンシークレット",
            },
            {
                Name:        "password-field",
                Pattern:     regexp.MustCompile(`(?i)(password["\s]*[:=]["\s]*)[^\s"']+`),
                Replacement: "${1}[FILTERED]",
                Description: "パスワードフィールド",
            },
            {
                Name:        "api-key-generic",
                Pattern:     regexp.MustCompile(`(?i)(api[_-]?key["\s]*[:=]["\s]*)[a-zA-Z0-9]{16,}`),
                Replacement: "${1}[FILTERED]",
                Description: "汎用APIキー",
            },
        },
        maskChar:   '*',
        maskLength: 8,
    }
}

func (sdf *SensitiveDataFilter) FilterString(input string) string {
    filtered := input
    
    for _, pattern := range sdf.patterns {
        filtered = pattern.Pattern.ReplaceAllString(filtered, pattern.Replacement)
    }
    
    return filtered
}

func (sdf *SensitiveDataFilter) FilterLogEntry(entry string) string {
    return sdf.FilterString(entry)
}

// カスタムログライター
type SecureLogWriter struct {
    writer io.Writer
    filter *SensitiveDataFilter
}

func (slw *SecureLogWriter) Write(p []byte) (n int, err error) {
    filtered := slw.filter.FilterString(string(p))
    return slw.writer.Write([]byte(filtered))
}
```

### 3. セキュア入力システム
```go
type SecureInput struct {
    stdin    io.Reader
    stdout   io.Writer
    terminal *term.Terminal
}

func NewSecureInput() *SecureInput {
    return &SecureInput{
        stdin:    os.Stdin,
        stdout:   os.Stdout,
        terminal: term.NewTerminal(os.Stdin, ""),
    }
}

func (si *SecureInput) ReadPassword(prompt string) (string, error) {
    fmt.Fprint(si.stdout, prompt)
    
    // 端末の状態を取得
    fd := int(os.Stdin.Fd())
    oldState, err := term.MakeRaw(fd)
    if err != nil {
        return "", fmt.Errorf("failed to set raw mode: %w", err)
    }
    defer term.Restore(fd, oldState)
    
    // パスワード入力（エコーなし）
    password, err := si.terminal.ReadPassword(prompt)
    if err != nil {
        return "", fmt.Errorf("failed to read password: %w", err)
    }
    
    fmt.Fprintln(si.stdout) // 改行
    return password, nil
}

func (si *SecureInput) ReadSensitiveValue(prompt, fieldName string) (string, error) {
    fmt.Fprintf(si.stdout, "%s (入力は非表示になります): ", prompt)
    
    value, err := si.ReadPassword("")
    if err != nil {
        return "", fmt.Errorf("failed to read %s: %w", fieldName, err)
    }
    
    // 基本的なバリデーション
    if strings.TrimSpace(value) == "" {
        return "", fmt.Errorf("%s cannot be empty", fieldName)
    }
    
    return value, nil
}

func (si *SecureInput) ConfirmOverwrite(message string) (bool, error) {
    fmt.Fprintf(si.stdout, "%s (y/N): ", message)
    
    response, err := si.terminal.ReadLine()
    if err != nil {
        return false, err
    }
    
    response = strings.TrimSpace(strings.ToLower(response))
    return response == "y" || response == "yes", nil
}
```

### 4. 監査ログシステム
```go
type AuditLogger struct {
    logger    *log.Logger
    logFile   *os.File
    filter    *SensitiveDataFilter
}

type AuditEvent struct {
    Timestamp   time.Time              `json:"timestamp"`
    EventType   string                 `json:"event_type"`
    UserID      string                 `json:"user_id,omitempty"`
    Action      string                 `json:"action"`
    Resource    string                 `json:"resource"`
    Status      string                 `json:"status"`
    Details     map[string]interface{} `json:"details,omitempty"`
    ClientIP    string                 `json:"client_ip,omitempty"`
}

func (al *AuditLogger) LogCredentialAccess(credentialKey string, action string) {
    event := AuditEvent{
        Timestamp: time.Now(),
        EventType: "credential_access",
        Action:    action,
        Resource:  credentialKey,
        Status:    "success",
        Details: map[string]interface{}{
            "credential_type": "sakura_cloud_api",
        },
    }
    
    al.writeAuditEvent(event)
}

func (al *AuditLogger) LogCredentialRotation(credentialKey string, oldVersion, newVersion int) {
    event := AuditEvent{
        Timestamp: time.Now(),
        EventType: "credential_rotation",
        Action:    "rotate",
        Resource:  credentialKey,
        Status:    "success",
        Details: map[string]interface{}{
            "old_version": oldVersion,
            "new_version": newVersion,
        },
    }
    
    al.writeAuditEvent(event)
}

func (al *AuditLogger) LogSecurityViolation(violation string, details map[string]interface{}) {
    event := AuditEvent{
        Timestamp: time.Now(),
        EventType: "security_violation",
        Action:    "alert",
        Resource:  "system",
        Status:    "violation",
        Details:   details,
    }
    
    al.writeAuditEvent(event)
}

func (al *AuditLogger) writeAuditEvent(event AuditEvent) {
    eventJSON, err := json.Marshal(event)
    if err != nil {
        al.logger.Printf("Failed to marshal audit event: %v", err)
        return
    }
    
    // 機密情報をフィルタリング
    filteredJSON := al.filter.FilterString(string(eventJSON))
    
    al.logger.Println(filteredJSON)
}
```

### 5. 認証情報ローテーション警告
```go
type CredentialMonitor struct {
    storage      *SecureStorage
    auditLogger  *AuditLogger
    alertThresholds map[string]time.Duration
}

func (cm *CredentialMonitor) CheckCredentialAge() []SecurityAlert {
    var alerts []SecurityAlert
    
    credentials, err := cm.storage.ListCredentials()
    if err != nil {
        return alerts
    }
    
    for _, cred := range credentials {
        age := time.Since(cred.CreatedAt)
        threshold, exists := cm.alertThresholds[cred.Key]
        
        if !exists {
            threshold = 90 * 24 * time.Hour // デフォルト90日
        }
        
        if age > threshold {
            alert := SecurityAlert{
                Type:        "credential_age",
                Severity:    "warning",
                Message:     fmt.Sprintf("認証情報 '%s' が%d日前に作成されました。更新を検討してください。", cred.Key, int(age.Hours()/24)),
                Resource:    cred.Key,
                CreatedAt:   time.Now(),
                Remediation: "新しい認証情報を生成し、設定を更新してください。",
            }
            alerts = append(alerts, alert)
            
            cm.auditLogger.LogSecurityViolation("credential_age_warning", map[string]interface{}{
                "credential_key": cred.Key,
                "age_days":      int(age.Hours() / 24),
            })
        }
    }
    
    return alerts
}

type SecurityAlert struct {
    Type        string    `json:"type"`
    Severity    string    `json:"severity"`
    Message     string    `json:"message"`
    Resource    string    `json:"resource"`
    CreatedAt   time.Time `json:"created_at"`
    Remediation string    `json:"remediation"`
}

func (cm *CredentialMonitor) CheckUnusedCredentials() []SecurityAlert {
    var alerts []SecurityAlert
    
    credentials, err := cm.storage.ListCredentials()
    if err != nil {
        return alerts
    }
    
    unusedThreshold := 30 * 24 * time.Hour // 30日間未使用
    
    for _, cred := range credentials {
        if time.Since(cred.LastUsedAt) > unusedThreshold {
            alert := SecurityAlert{
                Type:     "unused_credential",
                Severity: "info",
                Message:  fmt.Sprintf("認証情報 '%s' が%d日間使用されていません。", cred.Key, int(time.Since(cred.LastUsedAt).Hours()/24)),
                Resource: cred.Key,
                CreatedAt: time.Now(),
                Remediation: "不要な場合は削除を検討してください。",
            }
            alerts = append(alerts, alert)
        }
    }
    
    return alerts
}
```

## テスト戦略
- **暗号化テスト**: 暗号化・復号化の正確性確認
- **フィルタリングテスト**: 機密情報の漏洩防止確認
- **入力セキュリティテスト**: セキュア入力機能の動作確認
- **監査ログテスト**: 監査イベントの記録精度確認

## 依存関係
- 前提PBI: PBI-030（マルチプロファイル管理）
- 関連PBI: PBI-033（権限管理）、PBI-034（不正アクセス検知）
- 既存コード: internal/config/

## 見積もり
- 開発工数: 16時間
  - 暗号化保存システム: 6時間
  - 機密情報フィルタリング: 4時間
  - セキュア入力システム: 3時間
  - 監査・ローテーション警告: 3時間

## 完了の定義
- [x] 機密情報の暗号化保存が正常に機能する
- [x] ログ出力から機密情報が完全に除外される
- [x] セキュア入力でマスク入力ができる
- [x] 監査ログが適切に記録される
- [x] 古い認証情報に対する警告が表示される

## 実装状況
✅ **PBI-032は完全に実装済み** (2025-09-11)

以下のファイルで完全に実装されています：
- `internal/security/encryption.go` - AES-256-GCM暗号化システム
- `internal/security/filter.go` - 機密情報フィルタリング
- `internal/security/input.go` - セキュア入力システム
- `internal/security/audit.go` - 監査ログ機能
- `internal/security/monitor.go` - 認証情報監視
- `internal/security/encryption_test.go` - 包括的なテストスイート
- `internal/security/filter_test.go` - フィルタリングテスト

実装内容：
- AES-256-GCMによる暗号化保存システム
- HKDF鍵導出によるセキュアな鍵管理
- 機密情報自動フィルタリングシステム
- マスク入力・パスワード入力機能
- JSON Lines形式の監査ログ
- 認証情報のローテーション監視
- 包括的なテストカバレッジ

## 備考
- 暗号化にはAES-256-GCMを使用（業界標準）
- キー管理はHDKFによる鍵導出を採用
- 監査ログはJSON Lines形式で出力
- GDPR等のプライバシー規制も考慮した設計