package e2e

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// E2ETestSuite はエンドツーエンドテストスイート
type E2ETestSuite struct {
	t          *testing.T
	testDir    string
	binaryPath string

	// テスト環境
	tempHome     string
	tempConfig   string
	cleanupFuncs []func()

	// 実行設定
	timeout time.Duration
	verbose bool
}

// E2ETestOptions はE2Eテストオプション
type E2ETestOptions struct {
	// 実行設定
	Arguments   []string          `yaml:"arguments"`
	Environment map[string]string `yaml:"environment"`
	WorkingDir  string            `yaml:"working_dir"`
	Timeout     string            `yaml:"timeout"`

	// 入力設定
	StdinInput        string   `yaml:"stdin_input"`
	InteractiveInputs []string `yaml:"interactive_inputs"`
	InputFiles        []string `yaml:"input_files"`

	// 期待結果
	ExpectedExitCode int               `yaml:"expected_exit_code"`
	ExpectedStdout   []string          `yaml:"expected_stdout"`
	ExpectedStderr   []string          `yaml:"expected_stderr"`
	ExpectedFiles    []FileExpectation `yaml:"expected_files"`
	ExpectedNoFiles  []string          `yaml:"expected_no_files"`

	// 検証設定
	ValidateOutput      bool   `yaml:"validate_output"`
	ValidateFiles       bool   `yaml:"validate_files"`
	ValidatePerformance bool   `yaml:"validate_performance"`
	MaxExecutionTime    string `yaml:"max_execution_time"`
	MaxMemoryUsage      string `yaml:"max_memory_usage"`
}

// FileExpectation はファイル期待値
type FileExpectation struct {
	Path            string   `yaml:"path"`
	ShouldExist     bool     `yaml:"should_exist"`
	ContentContains []string `yaml:"content_contains"`
	ContentExact    string   `yaml:"content_exact"`
	MinSize         int64    `yaml:"min_size"`
	MaxSize         int64    `yaml:"max_size"`
}

// E2ETestResult はE2Eテスト結果
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

// EnvironmentSnapshot は環境のスナップショット
type EnvironmentSnapshot struct {
	Files     map[string]FileInfo `json:"files"`
	Timestamp time.Time           `json:"timestamp"`
}

// FileInfo はファイル情報
type FileInfo struct {
	Size    int64     `json:"size"`
	ModTime time.Time `json:"mod_time"`
	IsDir   bool      `json:"is_dir"`
}

// NewE2ETestSuite は新しいE2Eテストスイートを作成
func NewE2ETestSuite(t *testing.T) *E2ETestSuite {
	suite := &E2ETestSuite{
		t:       t,
		testDir: t.TempDir(),
		timeout: 5 * time.Minute,
		verbose: testing.Verbose(),
	}

	suite.setupTestEnvironment()
	return suite
}

// setupTestEnvironment はテスト環境をセットアップ
func (e2e *E2ETestSuite) setupTestEnvironment() {
	e2e.t.Helper()

	// テストバイナリをビルド
	e2e.buildTestBinary()

	// 分離された環境をセットアップ
	e2e.setupIsolatedEnvironment()

	// クリーンアップ関数の登録
	e2e.t.Cleanup(e2e.cleanup)
}

// buildTestBinary はテストバイナリをビルド
func (e2e *E2ETestSuite) buildTestBinary() {
	e2e.t.Helper()

	// プロジェクトルートディレクトリを特定
	projectRoot, err := findProjectRoot()
	if err != nil {
		e2e.t.Fatalf("プロジェクトルートの特定に失敗: %v", err)
	}

	// バイナリパスを設定
	e2e.binaryPath = filepath.Join(e2e.testDir, "usacloud-update")
	if testing.Short() {
		// 短時間モードでは既存のバイナリを使用
		existingBinary := filepath.Join(projectRoot, "bin", "usacloud-update")
		if _, err := os.Stat(existingBinary); err == nil {
			e2e.binaryPath = existingBinary
			return
		}
	}

	// バイナリをビルド
	buildCmd := exec.Command("go", "build", "-o", e2e.binaryPath, "./cmd/usacloud-update")
	buildCmd.Dir = projectRoot
	buildCmd.Env = append(os.Environ(), "CGO_ENABLED=0")

	if output, err := buildCmd.CombinedOutput(); err != nil {
		e2e.t.Fatalf("バイナリビルドに失敗: %v\n出力: %s", err, output)
	}

	// 実行可能にする
	if err := os.Chmod(e2e.binaryPath, 0755); err != nil {
		e2e.t.Fatalf("バイナリの実行権限設定に失敗: %v", err)
	}
}

// setupIsolatedEnvironment は分離された環境をセットアップ
func (e2e *E2ETestSuite) setupIsolatedEnvironment() {
	e2e.t.Helper()

	// 分離されたホームディレクトリ
	e2e.tempHome = filepath.Join(e2e.testDir, "home")
	if err := os.MkdirAll(e2e.tempHome, 0755); err != nil {
		e2e.t.Fatalf("テストホームディレクトリ作成に失敗: %v", err)
	}

	// 分離された設定ディレクトリ
	configDir := filepath.Join(e2e.tempHome, ".config", "usacloud-update")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		e2e.t.Fatalf("設定ディレクトリ作成に失敗: %v", err)
	}

	e2e.tempConfig = filepath.Join(configDir, "usacloud-update.conf")
}

// RunE2ETest はE2Eテストを実行
func (e2e *E2ETestSuite) RunE2ETest(testName string, options *E2ETestOptions) *E2ETestResult {
	e2e.t.Helper()

	// テスト前の状態スナップショット
	beforeSnapshot := e2e.takeEnvironmentSnapshot()

	// テスト実行
	result := e2e.executeTest(options)

	// テスト後の状態スナップショット
	afterSnapshot := e2e.takeEnvironmentSnapshot()

	// 環境変化を記録
	result.FilesCreated = e2e.findCreatedFiles(beforeSnapshot, afterSnapshot)
	result.FilesModified = e2e.findModifiedFiles(beforeSnapshot, afterSnapshot)

	// 期待結果との比較
	e2e.validateResult(testName, result, options)

	return result
}

// executeTest はテストを実行
func (e2e *E2ETestSuite) executeTest(options *E2ETestOptions) *E2ETestResult {
	// コマンド構築
	args := append([]string{}, options.Arguments...)
	cmd := exec.Command(e2e.binaryPath, args...)

	// 環境変数設定
	cmd.Env = e2e.buildEnvironment(options.Environment)

	// 作業ディレクトリ設定
	if options.WorkingDir != "" {
		cmd.Dir = filepath.Join(e2e.testDir, options.WorkingDir)
	} else {
		cmd.Dir = e2e.testDir
	}

	// 標準入力設定
	if options.StdinInput != "" {
		cmd.Stdin = strings.NewReader(options.StdinInput)
	}

	// 出力キャプチャ設定
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// インタラクティブ入力処理
	if len(options.InteractiveInputs) > 0 {
		return e2e.executeInteractiveTest(cmd, options, &stdout, &stderr)
	}

	// タイムアウト設定
	timeout := e2e.timeout
	if options.Timeout != "" {
		if parsedTimeout, err := time.ParseDuration(options.Timeout); err == nil {
			timeout = parsedTimeout
		}
	}

	// コンテキストでタイムアウト制御
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// 実行時間測定
	startTime := time.Now()
	err := cmd.Start()
	if err != nil {
		return &E2ETestResult{
			Error: fmt.Errorf("コマンド開始に失敗: %v", err),
		}
	}

	// タイムアウト監視
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	var cmdError error
	select {
	case cmdError = <-done:
		// 正常終了またはエラー終了
	case <-ctx.Done():
		// タイムアウト
		cmd.Process.Kill()
		cmdError = fmt.Errorf("タイムアウト: %v", timeout)
	}

	executionTime := time.Since(startTime)

	// 結果構築
	result := &E2ETestResult{
		ExitCode:      0,
		Stdout:        stdout.String(),
		Stderr:        stderr.String(),
		ExecutionTime: executionTime,
		Error:         cmdError,
	}

	if cmd.ProcessState != nil {
		result.ExitCode = cmd.ProcessState.ExitCode()
	} else if cmdError != nil {
		result.ExitCode = 1
	}

	return result
}

// executeInteractiveTest はインタラクティブテストを実行
func (e2e *E2ETestSuite) executeInteractiveTest(cmd *exec.Cmd, options *E2ETestOptions, stdout, stderr *strings.Builder) *E2ETestResult {
	// パイプ作成
	stdinPipe, err := cmd.StdinPipe()
	if err != nil {
		return &E2ETestResult{
			Error: fmt.Errorf("標準入力パイプ作成に失敗: %v", err),
		}
	}

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return &E2ETestResult{
			Error: fmt.Errorf("標準出力パイプ作成に失敗: %v", err),
		}
	}

	cmd.Stderr = stderr

	// コマンド開始
	startTime := time.Now()
	if err := cmd.Start(); err != nil {
		return &E2ETestResult{
			Error: fmt.Errorf("インタラクティブコマンド開始に失敗: %v", err),
		}
	}

	// 出力読み取り用ゴルーチン
	go func() {
		io.Copy(stdout, stdoutPipe)
	}()

	// インタラクティブ入力送信
	go func() {
		defer stdinPipe.Close()

		for _, input := range options.InteractiveInputs {
			time.Sleep(100 * time.Millisecond) // 出力待機
			fmt.Fprintln(stdinPipe, input)
			if e2e.verbose {
				e2e.t.Logf("インタラクティブ入力: %s", input)
			}
		}
	}()

	// プロセス終了待機
	err = cmd.Wait()
	executionTime := time.Since(startTime)

	result := &E2ETestResult{
		ExitCode:      cmd.ProcessState.ExitCode(),
		Stdout:        stdout.String(),
		Stderr:        stderr.String(),
		ExecutionTime: executionTime,
		Error:         err,
	}

	return result
}

// buildEnvironment は環境変数を構築
func (e2e *E2ETestSuite) buildEnvironment(customEnv map[string]string) []string {
	env := os.Environ()

	// 分離された環境変数を設定
	isolatedEnv := map[string]string{
		"HOME":                            e2e.tempHome,
		"USACLOUD_UPDATE_CONFIG_DIR":      filepath.Dir(e2e.tempConfig),
		"USACLOUD_UPDATE_NO_UPDATE_CHECK": "1",
		"NO_COLOR":                        "1", // テストでは色出力を無効化
	}

	// 分離環境変数を追加
	for key, value := range isolatedEnv {
		env = append(env, fmt.Sprintf("%s=%s", key, value))
	}

	// カスタム環境変数を追加
	for key, value := range customEnv {
		env = append(env, fmt.Sprintf("%s=%s", key, value))
	}

	return env
}

// takeEnvironmentSnapshot は環境のスナップショットを取得
func (e2e *E2ETestSuite) takeEnvironmentSnapshot() *EnvironmentSnapshot {
	snapshot := &EnvironmentSnapshot{
		Files:     make(map[string]FileInfo),
		Timestamp: time.Now(),
	}

	// テストディレクトリ配下のファイルをスキャン
	filepath.Walk(e2e.testDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // エラーは無視
		}

		relPath, _ := filepath.Rel(e2e.testDir, path)
		snapshot.Files[relPath] = FileInfo{
			Size:    info.Size(),
			ModTime: info.ModTime(),
			IsDir:   info.IsDir(),
		}

		return nil
	})

	return snapshot
}

// findCreatedFiles は作成されたファイルを検出
func (e2e *E2ETestSuite) findCreatedFiles(before, after *EnvironmentSnapshot) []string {
	var created []string

	for path := range after.Files {
		if _, exists := before.Files[path]; !exists {
			created = append(created, path)
		}
	}

	return created
}

// findModifiedFiles は変更されたファイルを検出
func (e2e *E2ETestSuite) findModifiedFiles(before, after *EnvironmentSnapshot) []string {
	var modified []string

	for path, afterInfo := range after.Files {
		if beforeInfo, exists := before.Files[path]; exists {
			if !afterInfo.ModTime.Equal(beforeInfo.ModTime) || afterInfo.Size != beforeInfo.Size {
				modified = append(modified, path)
			}
		}
	}

	return modified
}

// validateResult は結果を検証
func (e2e *E2ETestSuite) validateResult(testName string, result *E2ETestResult, options *E2ETestOptions) {
	e2e.t.Helper()

	// 終了コードの検証
	if result.ExitCode != options.ExpectedExitCode {
		e2e.t.Errorf("%s: 終了コードが期待値と異なります: got=%d, want=%d",
			testName, result.ExitCode, options.ExpectedExitCode)
	}

	// 標準出力の検証
	for _, expected := range options.ExpectedStdout {
		if !strings.Contains(result.Stdout, expected) {
			e2e.t.Errorf("%s: 標準出力に期待する文字列が含まれていません: %q\n出力:\n%s",
				testName, expected, result.Stdout)
		}
	}

	// 標準エラー出力の検証
	for _, expected := range options.ExpectedStderr {
		if !strings.Contains(result.Stderr, expected) {
			e2e.t.Errorf("%s: 標準エラー出力に期待する文字列が含まれていません: %q\n出力:\n%s",
				testName, expected, result.Stderr)
		}
	}

	// ファイル検証
	if options.ValidateFiles {
		e2e.validateFiles(testName, options.ExpectedFiles, options.ExpectedNoFiles)
	}

	// パフォーマンス検証
	if options.ValidatePerformance {
		e2e.validatePerformance(testName, result, options)
	}

	// 詳細ログ出力
	if e2e.verbose {
		e2e.t.Logf("%s 実行結果:", testName)
		e2e.t.Logf("  終了コード: %d", result.ExitCode)
		e2e.t.Logf("  実行時間: %v", result.ExecutionTime)
		e2e.t.Logf("  作成ファイル: %v", result.FilesCreated)
		e2e.t.Logf("  変更ファイル: %v", result.FilesModified)
	}
}

// validateFiles はファイルを検証
func (e2e *E2ETestSuite) validateFiles(testName string, expectedFiles []FileExpectation, expectedNoFiles []string) {
	for _, fileExp := range expectedFiles {
		fullPath := filepath.Join(e2e.testDir, fileExp.Path)

		if fileExp.ShouldExist {
			if _, err := os.Stat(fullPath); os.IsNotExist(err) {
				e2e.t.Errorf("%s: ファイルが存在しません: %s", testName, fileExp.Path)
				continue
			}

			// ファイル内容の検証
			if len(fileExp.ContentContains) > 0 {
				content, err := ioutil.ReadFile(fullPath)
				if err != nil {
					e2e.t.Errorf("%s: ファイル読み取りエラー %s: %v", testName, fileExp.Path, err)
					continue
				}

				contentStr := string(content)
				for _, expected := range fileExp.ContentContains {
					if !strings.Contains(contentStr, expected) {
						e2e.t.Errorf("%s: ファイル %s に期待する内容が含まれていません: %q",
							testName, fileExp.Path, expected)
					}
				}
			}
		} else {
			if _, err := os.Stat(fullPath); !os.IsNotExist(err) {
				e2e.t.Errorf("%s: ファイルが存在すべきではありません: %s", testName, fileExp.Path)
			}
		}
	}

	// 存在すべきでないファイルの確認
	for _, noFile := range expectedNoFiles {
		fullPath := filepath.Join(e2e.testDir, noFile)
		if _, err := os.Stat(fullPath); !os.IsNotExist(err) {
			e2e.t.Errorf("%s: ファイルが存在すべきではありません: %s", testName, noFile)
		}
	}
}

// validatePerformance はパフォーマンスを検証
func (e2e *E2ETestSuite) validatePerformance(testName string, result *E2ETestResult, options *E2ETestOptions) {
	if options.MaxExecutionTime != "" {
		maxTime, err := time.ParseDuration(options.MaxExecutionTime)
		if err == nil && result.ExecutionTime > maxTime {
			e2e.t.Errorf("%s: 実行時間が上限を超過: %v > %v", testName, result.ExecutionTime, maxTime)
		}
	}
}

// CreateTempFile は一時ファイルを作成
func (e2e *E2ETestSuite) CreateTempFile(name, content string) string {
	filePath := filepath.Join(e2e.testDir, name)

	// ディレクトリ作成
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		e2e.t.Fatalf("ディレクトリ作成に失敗 %s: %v", dir, err)
	}

	// ファイル作成
	if err := ioutil.WriteFile(filePath, []byte(content), 0644); err != nil {
		e2e.t.Fatalf("ファイル作成に失敗 %s: %v", filePath, err)
	}

	return filePath
}

// CreateTempFileBytes はバイナリ一時ファイルを作成
func (e2e *E2ETestSuite) CreateTempFileBytes(name string, content []byte) string {
	filePath := filepath.Join(e2e.testDir, name)

	// ディレクトリ作成
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		e2e.t.Fatalf("ディレクトリ作成に失敗 %s: %v", dir, err)
	}

	// ファイル作成
	if err := ioutil.WriteFile(filePath, content, 0644); err != nil {
		e2e.t.Fatalf("ファイル作成に失敗 %s: %v", filePath, err)
	}

	return filePath
}

// CreateTempDir は一時ディレクトリを作成
func (e2e *E2ETestSuite) CreateTempDir(name string) string {
	dirPath := filepath.Join(e2e.testDir, name)

	if err := os.MkdirAll(dirPath, 0755); err != nil {
		e2e.t.Fatalf("ディレクトリ作成に失敗 %s: %v", dirPath, err)
	}

	return dirPath
}

// GetTestDir はテストディレクトリのパスを返す
func (e2e *E2ETestSuite) GetTestDir() string {
	return e2e.testDir
}

// cleanup はテスト後のクリーンアップを実行
func (e2e *E2ETestSuite) cleanup() {
	for _, cleanupFunc := range e2e.cleanupFuncs {
		cleanupFunc()
	}
}

// findProjectRoot はプロジェクトルートを特定
func findProjectRoot() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// go.modファイルを探してプロジェクトルートを特定
	dir := wd
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return "", fmt.Errorf("プロジェクトルート（go.mod）が見つかりません")
}
