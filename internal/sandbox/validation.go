package sandbox

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
)

// Severity はバリデーション結果の重要度を表す
type Severity int

const (
	SeverityInfo Severity = iota
	SeverityWarning
	SeverityError
	SeverityCritical
)

// String はSeverityの文字列表現を返す
func (s Severity) String() string {
	switch s {
	case SeverityInfo:
		return "info"
	case SeverityWarning:
		return "warning"
	case SeverityError:
		return "error"
	case SeverityCritical:
		return "critical"
	default:
		return "unknown"
	}
}

// ValidationResult はバリデーション結果を表す
type ValidationResult struct {
	CheckName string        `json:"check_name"`
	Passed    bool          `json:"passed"`
	Message   string        `json:"message"`
	Severity  Severity      `json:"severity"`
	FixAction string        `json:"fix_action,omitempty"`
	HelpURL   string        `json:"help_url,omitempty"`
	Details   string        `json:"details,omitempty"`
	Duration  time.Duration `json:"duration"`
}

// ValidationCheck はバリデーションチェックのインターフェース
type ValidationCheck interface {
	Name() string
	Description() string
	Validate() *ValidationResult
	Fix() error
}

// EnvironmentValidator は環境設定のバリデーションを管理する
type EnvironmentValidator struct {
	checks []ValidationCheck
}

// NewEnvironmentValidator は新しいEnvironmentValidatorを作成する
func NewEnvironmentValidator() *EnvironmentValidator {
	ev := &EnvironmentValidator{}

	// 標準チェックを追加
	ev.AddCheck(&USACloudCLICheck{requiredVersion: "1.43.0"})
	ev.AddCheck(&APIKeyCheck{})
	ev.AddCheck(&NetworkCheck{
		endpoints: []string{
			"https://secure.sakura.ad.jp",
			"https://cloud-api.sakura.ad.jp",
		},
		timeout: 10 * time.Second,
	})
	ev.AddCheck(&ZoneAccessCheck{zone: "tk1v"})
	ev.AddCheck(&ConfigFileCheck{})

	return ev
}

// AddCheck は新しいバリデーションチェックを追加する
func (ev *EnvironmentValidator) AddCheck(check ValidationCheck) {
	ev.checks = append(ev.checks, check)
}

// RunAllChecks はすべてのバリデーションチェックを実行する
func (ev *EnvironmentValidator) RunAllChecks() []*ValidationResult {
	var results []*ValidationResult

	for _, check := range ev.checks {
		start := time.Now()
		result := check.Validate()
		result.CheckName = check.Name()
		result.Duration = time.Since(start)
		results = append(results, result)
	}

	return results
}

// HasCriticalErrors は重大なエラーがあるかチェックする
func (ev *EnvironmentValidator) HasCriticalErrors(results []*ValidationResult) bool {
	for _, result := range results {
		if !result.Passed && result.Severity >= SeverityError {
			return true
		}
	}
	return false
}

// GenerateReport は検証結果のレポートを生成する
func (ev *EnvironmentValidator) GenerateReport(results []*ValidationResult) string {
	var report strings.Builder

	report.WriteString(color.HiWhiteString("🔍 サンドボックス環境検証結果\n"))
	report.WriteString(color.HiWhiteString("================================\n\n"))

	passedCount := 0
	warningCount := 0
	errorCount := 0

	for _, result := range results {
		icon := color.GreenString("✅")
		if !result.Passed {
			switch result.Severity {
			case SeverityWarning:
				icon = color.YellowString("⚠️")
				warningCount++
			case SeverityError:
				icon = color.RedString("❌")
				errorCount++
			case SeverityCritical:
				icon = color.HiRedString("🚫")
				errorCount++
			}
		} else {
			passedCount++
		}

		report.WriteString(fmt.Sprintf("%s %s: %s", icon, result.CheckName, result.Message))
		if result.Duration > 0 {
			report.WriteString(color.HiBlackString(fmt.Sprintf(" (%v)", result.Duration.Truncate(time.Millisecond))))
		}
		report.WriteString("\n")

		if !result.Passed && result.FixAction != "" {
			report.WriteString(color.CyanString(fmt.Sprintf("   💡 対処方法: %s\n", result.FixAction)))
			if result.HelpURL != "" {
				report.WriteString(color.BlueString(fmt.Sprintf("   📖 詳細: %s\n", result.HelpURL)))
			}
		}

		if result.Details != "" {
			report.WriteString(color.HiBlackString(fmt.Sprintf("   ℹ️  詳細: %s\n", result.Details)))
		}

		report.WriteString("\n")
	}

	// サマリー
	report.WriteString(color.HiWhiteString("📊 検証サマリー\n"))
	report.WriteString("================\n")
	report.WriteString(color.GreenString(fmt.Sprintf("✅ 成功: %d\n", passedCount)))
	if warningCount > 0 {
		report.WriteString(color.YellowString(fmt.Sprintf("⚠️  警告: %d\n", warningCount)))
	}
	if errorCount > 0 {
		report.WriteString(color.RedString(fmt.Sprintf("❌ エラー: %d\n", errorCount)))
	}

	if errorCount > 0 {
		report.WriteString("\n" + color.HiRedString("⚠️  重大なエラーがあります。上記の対処方法に従って修正してください。\n"))
	} else if warningCount > 0 {
		report.WriteString("\n" + color.YellowString("ℹ️  警告がありますが、サンドボックス実行は可能です。\n"))
	} else {
		report.WriteString("\n" + color.GreenString("✅ すべての検証が正常に完了しました。サンドボックス実行の準備ができています。\n"))
	}

	return report.String()
}

// USACloudCLICheck はusacloud CLIの存在とバージョンをチェックする
type USACloudCLICheck struct {
	requiredVersion string
}

func (c *USACloudCLICheck) Name() string {
	return "usacloud CLI"
}

func (c *USACloudCLICheck) Description() string {
	return "usacloud CLIのインストール状況とバージョンを確認します"
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
			Details:   fmt.Sprintf("実行エラー: %v", err),
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

func (c *USACloudCLICheck) Fix() error {
	return fmt.Errorf("自動修復は対応していません。手動でusacloud CLIをインストールしてください")
}

func (c *USACloudCLICheck) isVersionCompatible(version string) bool {
	// バージョン文字列から数値部分を抽出（例: "usacloud version 1.43.0" → "1.43.0"）
	re := regexp.MustCompile(`(\d+)\.(\d+)\.(\d+)`)
	matches := re.FindStringSubmatch(version)
	if len(matches) < 4 {
		return false
	}

	currentMajor, _ := strconv.Atoi(matches[1])
	currentMinor, _ := strconv.Atoi(matches[2])
	currentPatch, _ := strconv.Atoi(matches[3])

	requiredMatches := re.FindStringSubmatch(c.requiredVersion)
	if len(requiredMatches) < 4 {
		return true // 必要バージョンがパースできない場合は通す
	}

	requiredMajor, _ := strconv.Atoi(requiredMatches[1])
	requiredMinor, _ := strconv.Atoi(requiredMatches[2])
	requiredPatch, _ := strconv.Atoi(requiredMatches[3])

	// バージョン比較
	if currentMajor > requiredMajor {
		return true
	}
	if currentMajor == requiredMajor && currentMinor > requiredMinor {
		return true
	}
	if currentMajor == requiredMajor && currentMinor == requiredMinor && currentPatch >= requiredPatch {
		return true
	}

	return false
}

// APIKeyCheck はAPIキーの有効性をチェックする
type APIKeyCheck struct{}

func (c *APIKeyCheck) Name() string {
	return "APIキー"
}

func (c *APIKeyCheck) Description() string {
	return "Sakura Cloud APIキーの設定と有効性を確認します"
}

func (c *APIKeyCheck) Validate() *ValidationResult {
	accessToken := os.Getenv("SAKURACLOUD_ACCESS_TOKEN")
	accessTokenSecret := os.Getenv("SAKURACLOUD_ACCESS_TOKEN_SECRET")

	if accessToken == "" || accessTokenSecret == "" {
		return &ValidationResult{
			Passed:    false,
			Message:   "APIキーが設定されていません",
			Severity:  SeverityCritical,
			FixAction: "環境変数またはusacloud-update.confファイルにAPIキーを設定してください",
			HelpURL:   "https://docs.usacloud.jp/configuration/",
		}
	}

	// APIキーの形式チェック（基本的な長さチェック）
	if len(accessToken) < 20 || len(accessTokenSecret) < 30 {
		return &ValidationResult{
			Passed:    false,
			Message:   "APIキーの形式が正しくありません",
			Severity:  SeverityError,
			FixAction: "正しい形式のAPIキーを設定してください",
			HelpURL:   "https://docs.usacloud.jp/configuration/",
		}
	}

	// 簡単なAPI呼び出しでキーの有効性をテスト
	cmd := exec.Command("usacloud", "auth-status", "--zone", "tk1v")
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("SAKURACLOUD_ACCESS_TOKEN=%s", accessToken),
		fmt.Sprintf("SAKURACLOUD_ACCESS_TOKEN_SECRET=%s", accessTokenSecret),
	)

	output, err := cmd.Output()
	if err != nil {
		return &ValidationResult{
			Passed:    false,
			Message:   "APIキーが無効です",
			Severity:  SeverityError,
			FixAction: "正しいAPIキーを設定してください",
			HelpURL:   "https://docs.usacloud.jp/configuration/",
			Details:   fmt.Sprintf("認証テストに失敗: %v", err),
		}
	}

	return &ValidationResult{
		Passed:   true,
		Message:  "APIキーは有効です",
		Severity: SeverityInfo,
		Details:  strings.TrimSpace(string(output)),
	}
}

func (c *APIKeyCheck) Fix() error {
	return fmt.Errorf("自動修復は対応していません。手動でAPIキーを設定してください")
}

// NetworkCheck はネットワーク接続をチェックする
type NetworkCheck struct {
	endpoints []string
	timeout   time.Duration
}

func (c *NetworkCheck) Name() string {
	return "ネットワーク接続"
}

func (c *NetworkCheck) Description() string {
	return "Sakura Cloud APIエンドポイントへの接続を確認します"
}

func (c *NetworkCheck) Validate() *ValidationResult {
	var failedEndpoints []string

	for _, endpoint := range c.endpoints {
		if !c.testConnection(endpoint) {
			failedEndpoints = append(failedEndpoints, endpoint)
		}
	}

	if len(failedEndpoints) > 0 {
		return &ValidationResult{
			Passed:    false,
			Message:   fmt.Sprintf("一部のエンドポイントへの接続に失敗しました: %s", strings.Join(failedEndpoints, ", ")),
			Severity:  SeverityError,
			FixAction: "ネットワーク接続またはプロキシ設定を確認してください",
			HelpURL:   "https://docs.usacloud.jp/troubleshooting/network/",
		}
	}

	return &ValidationResult{
		Passed:   true,
		Message:  "すべてのエンドポイントに正常に接続できます",
		Severity: SeverityInfo,
		Details:  fmt.Sprintf("テスト対象: %s", strings.Join(c.endpoints, ", ")),
	}
}

func (c *NetworkCheck) Fix() error {
	return fmt.Errorf("自動修復は対応していません。ネットワーク設定を手動で確認してください")
}

func (c *NetworkCheck) testConnection(endpoint string) bool {
	client := &http.Client{
		Timeout: c.timeout,
	}

	resp, err := client.Get(endpoint)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	// 接続できればOK（認証エラーなどは気にしない）
	return true
}

// ZoneAccessCheck は指定ゾーンへのアクセス権限をチェックする
type ZoneAccessCheck struct {
	zone string
}

func (c *ZoneAccessCheck) Name() string {
	return "ゾーンアクセス権限"
}

func (c *ZoneAccessCheck) Description() string {
	return "指定されたゾーンへのアクセス権限を確認します"
}

func (c *ZoneAccessCheck) Validate() *ValidationResult {
	// 簡単なリソース一覧取得でゾーンアクセスをテスト
	cmd := exec.Command("usacloud", "zone", "list", "--zone", c.zone)
	output, err := cmd.Output()

	if err != nil {
		return &ValidationResult{
			Passed:    false,
			Message:   fmt.Sprintf("ゾーン '%s' へのアクセスに失敗しました", c.zone),
			Severity:  SeverityError,
			FixAction: "ゾーンアクセス権限または設定を確認してください",
			HelpURL:   "https://docs.usacloud.jp/configuration/",
			Details:   fmt.Sprintf("実行エラー: %v", err),
		}
	}

	return &ValidationResult{
		Passed:   true,
		Message:  fmt.Sprintf("ゾーン '%s' にアクセス可能です", c.zone),
		Severity: SeverityInfo,
		Details:  fmt.Sprintf("レスポンス: %s", strings.TrimSpace(string(output))),
	}
}

func (c *ZoneAccessCheck) Fix() error {
	return fmt.Errorf("自動修復は対応していません。ゾーンアクセス権限を手動で確認してください")
}

// ConfigFileCheck は設定ファイルの存在と内容をチェックする
type ConfigFileCheck struct{}

func (c *ConfigFileCheck) Name() string {
	return "設定ファイル"
}

func (c *ConfigFileCheck) Description() string {
	return "usacloud-update設定ファイルの存在と設定を確認します"
}

func (c *ConfigFileCheck) Validate() *ValidationResult {
	// 設定ファイルのパスを確認
	configPaths := []string{
		os.ExpandEnv("$HOME/.config/usacloud-update/usacloud-update.conf"),
		"./usacloud-update.conf",
		"./.env",
	}

	var foundPath string
	for _, path := range configPaths {
		if _, err := os.Stat(path); err == nil {
			foundPath = path
			break
		}
	}

	if foundPath == "" {
		return &ValidationResult{
			Passed:    false,
			Message:   "設定ファイルが見つかりません",
			Severity:  SeverityWarning,
			FixAction: "usacloud-update.conf.sample を参考に設定ファイルを作成してください",
			HelpURL:   "https://github.com/armaniacs/usacloud-update#configuration",
			Details:   fmt.Sprintf("検索パス: %s", strings.Join(configPaths, ", ")),
		}
	}

	// 設定ファイルの内容をチェック（基本的な存在確認のみ）
	content, err := os.ReadFile(foundPath)
	if err != nil {
		return &ValidationResult{
			Passed:    false,
			Message:   "設定ファイルの読み込みに失敗しました",
			Severity:  SeverityError,
			FixAction: "設定ファイルの権限を確認してください",
			Details:   fmt.Sprintf("ファイル: %s, エラー: %v", foundPath, err),
		}
	}

	// 基本的な設定項目があるかチェック
	contentStr := string(content)
	hasConfig := strings.Contains(contentStr, "SAKURACLOUD_ACCESS_TOKEN") ||
		strings.Contains(contentStr, "access_token")

	if !hasConfig {
		return &ValidationResult{
			Passed:    false,
			Message:   "設定ファイルにAPIキー設定がありません",
			Severity:  SeverityWarning,
			FixAction: "設定ファイルにAPIキーを追加してください",
			Details:   fmt.Sprintf("ファイル: %s", foundPath),
		}
	}

	return &ValidationResult{
		Passed:   true,
		Message:  "設定ファイルが正常に設定されています",
		Severity: SeverityInfo,
		Details:  fmt.Sprintf("ファイル: %s", foundPath),
	}
}

func (c *ConfigFileCheck) Fix() error {
	return fmt.Errorf("自動修復は対応していません。設定ファイルを手動で作成・編集してください")
}
