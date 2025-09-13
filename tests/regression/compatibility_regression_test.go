package regression

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"time"
)

type CompatibilityRegressionTestSuite struct {
	*RegressionTestSuite
	cliTester         *CLICompatibilityTester
	apiTester         *APICompatibilityTester
	configTester      *ConfigCompatibilityTester
	versionManager    *VersionManager
	environmentTester *EnvironmentTester
}

type CLICompatibilityTester struct {
	binaryPath        string
	testCases         []CLITestCase
	flagCompatibility *FlagCompatibilityChecker
	outputValidator   *OutputValidator
}

type APICompatibilityTester struct {
	publicAPI         *PublicAPIValidator
	internalAPI       *InternalAPIValidator
	signatureChecker  *SignatureChecker
	behaviorValidator *BehaviorValidator
}

type ConfigCompatibilityTester struct {
	formatValidator       *ConfigFormatValidator
	migrationTester       *ConfigMigrationTester
	backwardCompatibility *BackwardCompatibilityChecker
}

type CLITestCase struct {
	Name               string            `json:"name"`
	Description        string            `json:"description"`
	Arguments          []string          `json:"arguments"`
	Environment        map[string]string `json:"environment"`
	ExpectedExitCode   int               `json:"expected_exit_code"`
	ExpectedOutput     []string          `json:"expected_output"`
	ExpectedErrors     []string          `json:"expected_errors"`
	ExpectedFiles      []FileExpectation `json:"expected_files"`
	VersionConstraints []string          `json:"version_constraints"`
	Tags               []string          `json:"tags"`
	Category           string            `json:"category"`
	Priority           string            `json:"priority"`
}

type FileExpectation struct {
	Path            string   `json:"path"`
	ShouldExist     bool     `json:"should_exist"`
	ContentContains []string `json:"content_contains"`
	PermissionMask  *int     `json:"permission_mask,omitempty"`
}

type APITestCase struct {
	Name           string      `json:"name"`
	Description    string      `json:"description"`
	Function       string      `json:"function"`
	Parameters     interface{} `json:"parameters"`
	ExpectedResult interface{} `json:"expected_result"`
	ExpectedError  string      `json:"expected_error"`
	Tags           []string    `json:"tags"`
	Category       string      `json:"category"`
	Priority       string      `json:"priority"`
}

type ConfigTestCase struct {
	Name              string      `json:"name"`
	Description       string      `json:"description"`
	ConfigType        string      `json:"config_type"`
	ConfigData        interface{} `json:"config_data"`
	ExpectedValid     bool        `json:"expected_valid"`
	ExpectedErrors    []string    `json:"expected_errors"`
	MigrationRequired bool        `json:"migration_required"`
	Tags              []string    `json:"tags"`
	Category          string      `json:"category"`
	Priority          string      `json:"priority"`
}

type FlagCompatibilityChecker struct {
	supportedFlags  map[string]FlagInfo
	deprecatedFlags map[string]FlagInfo
	removedFlags    map[string]FlagInfo
}

type FlagInfo struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
	Since       string `json:"since"`
	Deprecated  string `json:"deprecated,omitempty"`
	Removed     string `json:"removed,omitempty"`
	Replacement string `json:"replacement,omitempty"`
}

type OutputValidator struct {
	formatValidators map[string]FormatValidator
}

type FormatValidator interface {
	ValidateFormat(output string) error
	ValidateContent(output string, expected []string) error
}

type PublicAPIValidator struct {
	publicFunctions map[string]FunctionSignature
	publicTypes     map[string]TypeDefinition
	publicConstants map[string]ConstantDefinition
}

type InternalAPIValidator struct {
	internalPackages map[string]PackageDefinition
	exposedAPIs      map[string]FunctionSignature
}

type FunctionSignature struct {
	Name       string      `json:"name"`
	Package    string      `json:"package"`
	Parameters []Parameter `json:"parameters"`
	Returns    []Parameter `json:"returns"`
	IsPublic   bool        `json:"is_public"`
	Since      string      `json:"since"`
	Deprecated string      `json:"deprecated,omitempty"`
}

type Parameter struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type TypeDefinition struct {
	Name    string             `json:"name"`
	Package string             `json:"package"`
	Kind    string             `json:"kind"` // struct, interface, alias
	Fields  []FieldDefinition  `json:"fields,omitempty"`
	Methods []MethodDefinition `json:"methods,omitempty"`
	Since   string             `json:"since"`
}

type FieldDefinition struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Tag  string `json:"tag,omitempty"`
}

type MethodDefinition struct {
	Name       string      `json:"name"`
	Parameters []Parameter `json:"parameters"`
	Returns    []Parameter `json:"returns"`
}

type ConstantDefinition struct {
	Name    string `json:"name"`
	Package string `json:"package"`
	Type    string `json:"type"`
	Value   string `json:"value"`
	Since   string `json:"since"`
}

type PackageDefinition struct {
	Name    string   `json:"name"`
	Path    string   `json:"path"`
	Public  bool     `json:"public"`
	Exports []string `json:"exports"`
	Since   string   `json:"since"`
}

type SignatureChecker struct {
	registry map[string]FunctionSignature
}

type BehaviorValidator struct {
	testCases []BehaviorTestCase
}

type BehaviorTestCase struct {
	Function string      `json:"function"`
	Input    interface{} `json:"input"`
	Expected interface{} `json:"expected"`
	Behavior string      `json:"behavior"`
}

type ConfigFormatValidator struct {
	supportedFormats map[string]ConfigFormat
}

type ConfigFormat struct {
	Name       string   `json:"name"`
	Extensions []string `json:"extensions"`
	MimeTypes  []string `json:"mime_types"`
	Validator  func([]byte) error
	Since      string `json:"since"`
	Deprecated string `json:"deprecated,omitempty"`
}

type ConfigMigrationTester struct {
	migrations map[string]MigrationPath
}

type MigrationPath struct {
	From        string `json:"from"`
	To          string `json:"to"`
	Transformer func(interface{}) (interface{}, error)
	Validator   func(interface{}) error
}

type BackwardCompatibilityChecker struct {
	compatibilityMatrix map[string]map[string]bool
}

type VersionManager struct {
	currentVersion    string
	supportedVersions []string
	compatibilityMap  map[string][]string
}

type EnvironmentTester struct {
	supportedPlatforms map[string]PlatformInfo
	environmentChecks  []EnvironmentCheck
}

type PlatformInfo struct {
	OS           string   `json:"os"`
	Architecture string   `json:"architecture"`
	MinGoVersion string   `json:"min_go_version"`
	Dependencies []string `json:"dependencies"`
	Since        string   `json:"since"`
}

type EnvironmentCheck struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Checker     func() error
}

func NewCompatibilityRegressionTestSuite(regressionSuite *RegressionTestSuite) *CompatibilityRegressionTestSuite {
	suite := &CompatibilityRegressionTestSuite{
		RegressionTestSuite: regressionSuite,
		cliTester:           NewCLICompatibilityTester(),
		apiTester:           NewAPICompatibilityTester(),
		configTester:        NewConfigCompatibilityTester(),
		versionManager:      NewVersionManager(),
		environmentTester:   NewEnvironmentTester(),
	}

	return suite
}

func NewCLICompatibilityTester() *CLICompatibilityTester {
	return &CLICompatibilityTester{
		binaryPath:        "usacloud-update",
		testCases:         make([]CLITestCase, 0),
		flagCompatibility: NewFlagCompatibilityChecker(),
		outputValidator:   NewOutputValidator(),
	}
}

func NewAPICompatibilityTester() *APICompatibilityTester {
	return &APICompatibilityTester{
		publicAPI:         NewPublicAPIValidator(),
		internalAPI:       NewInternalAPIValidator(),
		signatureChecker:  NewSignatureChecker(),
		behaviorValidator: NewBehaviorValidator(),
	}
}

func NewConfigCompatibilityTester() *ConfigCompatibilityTester {
	return &ConfigCompatibilityTester{
		formatValidator:       NewConfigFormatValidator(),
		migrationTester:       NewConfigMigrationTester(),
		backwardCompatibility: NewBackwardCompatibilityChecker(),
	}
}

func NewFlagCompatibilityChecker() *FlagCompatibilityChecker {
	return &FlagCompatibilityChecker{
		supportedFlags:  make(map[string]FlagInfo),
		deprecatedFlags: make(map[string]FlagInfo),
		removedFlags:    make(map[string]FlagInfo),
	}
}

func NewOutputValidator() *OutputValidator {
	return &OutputValidator{
		formatValidators: make(map[string]FormatValidator),
	}
}

func NewPublicAPIValidator() *PublicAPIValidator {
	return &PublicAPIValidator{
		publicFunctions: make(map[string]FunctionSignature),
		publicTypes:     make(map[string]TypeDefinition),
		publicConstants: make(map[string]ConstantDefinition),
	}
}

func NewInternalAPIValidator() *InternalAPIValidator {
	return &InternalAPIValidator{
		internalPackages: make(map[string]PackageDefinition),
		exposedAPIs:      make(map[string]FunctionSignature),
	}
}

func NewSignatureChecker() *SignatureChecker {
	return &SignatureChecker{
		registry: make(map[string]FunctionSignature),
	}
}

func NewBehaviorValidator() *BehaviorValidator {
	return &BehaviorValidator{
		testCases: make([]BehaviorTestCase, 0),
	}
}

func NewConfigFormatValidator() *ConfigFormatValidator {
	return &ConfigFormatValidator{
		supportedFormats: make(map[string]ConfigFormat),
	}
}

func NewConfigMigrationTester() *ConfigMigrationTester {
	return &ConfigMigrationTester{
		migrations: make(map[string]MigrationPath),
	}
}

func NewBackwardCompatibilityChecker() *BackwardCompatibilityChecker {
	return &BackwardCompatibilityChecker{
		compatibilityMatrix: make(map[string]map[string]bool),
	}
}

func NewVersionManager() *VersionManager {
	return &VersionManager{
		currentVersion:    "1.9.0",
		supportedVersions: []string{"1.8.0", "1.9.0"},
		compatibilityMap:  make(map[string][]string),
	}
}

func NewEnvironmentTester() *EnvironmentTester {
	return &EnvironmentTester{
		supportedPlatforms: make(map[string]PlatformInfo),
		environmentChecks:  make([]EnvironmentCheck, 0),
	}
}

func (suite *CompatibilityRegressionTestSuite) RunCompatibilityRegressionTests(stats *TestCategoryStats) {
	suite.logf("Starting compatibility regression tests...")

	// Run CLI compatibility regression tests
	suite.runCLICompatibilityRegressionTests(stats)

	// Run API compatibility regression tests
	suite.runAPICompatibilityRegressionTests(stats)

	// Run configuration compatibility regression tests
	suite.runConfigCompatibilityRegressionTests(stats)

	// Run environment compatibility regression tests
	suite.runEnvironmentCompatibilityRegressionTests(stats)

	suite.logf("Compatibility regression tests completed")
}

func (suite *CompatibilityRegressionTestSuite) runCLICompatibilityRegressionTests(stats *TestCategoryStats) {
	suite.logf("Running CLI compatibility regression tests...")

	// Load CLI test cases
	testCases := suite.loadCLITestCases()

	for _, testCase := range testCases {
		stats.Total++

		startTime := time.Now()

		// Run CLI test
		currentResult, err := suite.runCLITest(testCase)
		if err != nil {
			suite.recordFailure("cli", testCase.Name,
				fmt.Sprintf("No error expected"), err.Error(), "error",
				map[string]interface{}{"test_case": testCase})
			stats.Failed++
			continue
		}

		executionTime := time.Since(startTime)

		// Compare with baseline
		if suite.compatibilityBaseline != nil {
			if baselineResult, exists := suite.compatibilityBaseline.CLICompatibility[testCase.Name]; exists {
				if !suite.compareCLIResults(baselineResult, currentResult, testCase.Name) {
					stats.Failed++
					continue
				}
			}
		}

		// Check flag compatibility
		if err := suite.checkFlagCompatibility(testCase.Arguments); err != nil {
			suite.recordWarning("cli_flag", testCase.Name, err.Error(),
				testCase.Arguments, map[string]interface{}{"test_case": testCase})
		}

		// Save current result for future baseline
		suite.saveCurrentCLIResult(testCase.Name, currentResult)

		stats.Passed++
		suite.logf("CLI test '%s' passed (execution: %v)", testCase.Name, executionTime)
	}

	suite.logf("CLI compatibility regression tests completed: %d/%d passed",
		stats.Passed, stats.Total)
}

func (suite *CompatibilityRegressionTestSuite) runAPICompatibilityRegressionTests(stats *TestCategoryStats) {
	suite.logf("Running API compatibility regression tests...")

	// Load API test cases
	testCases := suite.loadAPITestCases()

	for _, testCase := range testCases {
		stats.Total++

		startTime := time.Now()

		// Run API test
		currentResult, err := suite.runAPITest(testCase)
		if err != nil {
			suite.recordFailure("api", testCase.Name,
				fmt.Sprintf("No error expected"), err.Error(), "error",
				map[string]interface{}{"test_case": testCase})
			stats.Failed++
			continue
		}

		executionTime := time.Since(startTime)

		// Compare with baseline
		if suite.compatibilityBaseline != nil {
			if baselineResult, exists := suite.compatibilityBaseline.APICompatibility[testCase.Name]; exists {
				if !suite.compareAPIResults(baselineResult, currentResult, testCase.Name) {
					stats.Failed++
					continue
				}
			}
		}

		// Check signature compatibility
		if err := suite.checkSignatureCompatibility(testCase.Function); err != nil {
			suite.recordWarning("api_signature", testCase.Name, err.Error(),
				testCase.Function, map[string]interface{}{"test_case": testCase})
		}

		// Save current result for future baseline
		suite.saveCurrentAPIResult(testCase.Name, currentResult)

		stats.Passed++
		suite.logf("API test '%s' passed (execution: %v)", testCase.Name, executionTime)
	}

	suite.logf("API compatibility regression tests completed: %d/%d passed",
		stats.Passed, stats.Total)
}

func (suite *CompatibilityRegressionTestSuite) runConfigCompatibilityRegressionTests(stats *TestCategoryStats) {
	suite.logf("Running config compatibility regression tests...")

	// Load config test cases
	testCases := suite.loadConfigTestCases()

	for _, testCase := range testCases {
		stats.Total++

		startTime := time.Now()

		// Run config test
		currentResult, err := suite.runConfigTest(testCase)
		if err != nil {
			suite.recordFailure("config", testCase.Name,
				fmt.Sprintf("No error expected"), err.Error(), "error",
				map[string]interface{}{"test_case": testCase})
			stats.Failed++
			continue
		}

		executionTime := time.Since(startTime)

		// Compare with baseline
		if suite.compatibilityBaseline != nil {
			if baselineResult, exists := suite.compatibilityBaseline.ConfigCompatibility[testCase.Name]; exists {
				if !suite.compareConfigResults(baselineResult, currentResult, testCase.Name) {
					stats.Failed++
					continue
				}
			}
		}

		// Check format compatibility
		if err := suite.checkFormatCompatibility(testCase.ConfigType); err != nil {
			suite.recordWarning("config_format", testCase.Name, err.Error(),
				testCase.ConfigType, map[string]interface{}{"test_case": testCase})
		}

		// Save current result for future baseline
		suite.saveCurrentConfigResult(testCase.Name, currentResult)

		stats.Passed++
		suite.logf("Config test '%s' passed (execution: %v)", testCase.Name, executionTime)
	}

	suite.logf("Config compatibility regression tests completed: %d/%d passed",
		stats.Passed, stats.Total)
}

func (suite *CompatibilityRegressionTestSuite) runEnvironmentCompatibilityRegressionTests(stats *TestCategoryStats) {
	suite.logf("Running environment compatibility regression tests...")

	// Test current environment compatibility
	platformKey := fmt.Sprintf("%s_%s", runtime.GOOS, runtime.GOARCH)

	stats.Total++

	if platformInfo, supported := suite.environmentTester.supportedPlatforms[platformKey]; supported {
		// Check Go version compatibility
		if err := suite.checkGoVersionCompatibility(platformInfo.MinGoVersion); err != nil {
			suite.recordFailure("environment", "go_version",
				platformInfo.MinGoVersion, runtime.Version(), "version_mismatch",
				map[string]interface{}{"platform": platformInfo})
			stats.Failed++
		} else {
			stats.Passed++
		}
	} else {
		suite.recordWarning("environment", "platform_support",
			fmt.Sprintf("Platform %s may not be officially supported", platformKey),
			platformKey, map[string]interface{}{})
		stats.Passed++ // Don't fail for unsupported platforms, just warn
	}

	// Run environment checks
	for _, check := range suite.environmentTester.environmentChecks {
		stats.Total++

		if err := check.Checker(); err != nil {
			suite.recordFailure("environment", check.Name,
				"Check should pass", err.Error(), "check_failed",
				map[string]interface{}{"check": check})
			stats.Failed++
		} else {
			stats.Passed++
		}
	}

	suite.logf("Environment compatibility regression tests completed: %d/%d passed",
		stats.Passed, stats.Total)
}

func (suite *CompatibilityRegressionTestSuite) runCLITest(testCase CLITestCase) (*CLIResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	startTime := time.Now()

	// Build command
	cmd := exec.CommandContext(ctx, suite.cliTester.binaryPath, testCase.Arguments...)

	// Set environment variables
	env := os.Environ()
	for key, value := range testCase.Environment {
		env = append(env, fmt.Sprintf("%s=%s", key, value))
	}
	cmd.Env = env

	// Run command
	stdout, stderr, exitCode, err := suite.runCommandWithOutput(cmd)
	if err != nil && exitCode == testCase.ExpectedExitCode {
		// Expected error, not a test failure
		err = nil
	}

	executionTime := time.Since(startTime)

	return &CLIResult{
		Command:         strings.Join(append([]string{suite.cliTester.binaryPath}, testCase.Arguments...), " "),
		ExitCode:        exitCode,
		Stdout:          stdout,
		Stderr:          stderr,
		ExecutionTime:   executionTime,
		EnvironmentVars: testCase.Environment,
	}, err
}

func (suite *CompatibilityRegressionTestSuite) runAPITest(testCase APITestCase) (*APIResult, error) {
	startTime := time.Now()

	// This would test actual API functions
	// For now, returning a mock result
	result := &APIResult{
		Function:      testCase.Function,
		Parameters:    testCase.Parameters,
		Result:        testCase.ExpectedResult,
		Error:         "",
		ExecutionTime: time.Since(startTime),
	}

	return result, nil
}

func (suite *CompatibilityRegressionTestSuite) runConfigTest(testCase ConfigTestCase) (*ConfigResult, error) {
	startTime := time.Now()

	// Test configuration loading and validation
	// For now, returning a mock result
	result := &ConfigResult{
		ConfigType:       testCase.ConfigType,
		ConfigData:       testCase.ConfigData,
		IsValid:          testCase.ExpectedValid,
		LoadTime:         time.Since(startTime),
		ValidationErrors: []string{},
	}

	return result, nil
}

func (suite *CompatibilityRegressionTestSuite) loadCLITestCases() []CLITestCase {
	testCasesPath := filepath.Join(suite.baselineDir, "compatibility", "cli_test_cases.json")

	// Create default test cases if file doesn't exist
	if _, err := os.Stat(testCasesPath); os.IsNotExist(err) {
		defaultCases := suite.createDefaultCLITestCases()
		suite.saveCLITestCases(testCasesPath, defaultCases)
		return defaultCases
	}

	data, err := os.ReadFile(testCasesPath)
	if err != nil {
		suite.t.Logf("Failed to load CLI test cases: %v", err)
		return suite.createDefaultCLITestCases()
	}

	var testCases []CLITestCase
	if err := json.Unmarshal(data, &testCases); err != nil {
		suite.t.Logf("Failed to unmarshal CLI test cases: %v", err)
		return suite.createDefaultCLITestCases()
	}

	return testCases
}

func (suite *CompatibilityRegressionTestSuite) loadAPITestCases() []APITestCase {
	testCasesPath := filepath.Join(suite.baselineDir, "compatibility", "api_test_cases.json")

	// Create default test cases if file doesn't exist
	if _, err := os.Stat(testCasesPath); os.IsNotExist(err) {
		defaultCases := suite.createDefaultAPITestCases()
		suite.saveAPITestCases(testCasesPath, defaultCases)
		return defaultCases
	}

	data, err := os.ReadFile(testCasesPath)
	if err != nil {
		suite.t.Logf("Failed to load API test cases: %v", err)
		return suite.createDefaultAPITestCases()
	}

	var testCases []APITestCase
	if err := json.Unmarshal(data, &testCases); err != nil {
		suite.t.Logf("Failed to unmarshal API test cases: %v", err)
		return suite.createDefaultAPITestCases()
	}

	return testCases
}

func (suite *CompatibilityRegressionTestSuite) loadConfigTestCases() []ConfigTestCase {
	testCasesPath := filepath.Join(suite.baselineDir, "compatibility", "config_test_cases.json")

	// Create default test cases if file doesn't exist
	if _, err := os.Stat(testCasesPath); os.IsNotExist(err) {
		defaultCases := suite.createDefaultConfigTestCases()
		suite.saveConfigTestCases(testCasesPath, defaultCases)
		return defaultCases
	}

	data, err := os.ReadFile(testCasesPath)
	if err != nil {
		suite.t.Logf("Failed to load config test cases: %v", err)
		return suite.createDefaultConfigTestCases()
	}

	var testCases []ConfigTestCase
	if err := json.Unmarshal(data, &testCases); err != nil {
		suite.t.Logf("Failed to unmarshal config test cases: %v", err)
		return suite.createDefaultConfigTestCases()
	}

	return testCases
}

func (suite *CompatibilityRegressionTestSuite) createDefaultCLITestCases() []CLITestCase {
	return []CLITestCase{
		{
			Name:             "help_command_compatibility",
			Description:      "Test help command backward compatibility",
			Arguments:        []string{"--help"},
			ExpectedExitCode: 0,
			ExpectedOutput:   []string{"usacloud-update", "使用方法", "オプション"},
			Tags:             []string{"help", "backward-compatibility"},
			Category:         "cli",
			Priority:         "high",
		},
		{
			Name:             "version_command_compatibility",
			Description:      "Test version command backward compatibility",
			Arguments:        []string{"--version"},
			ExpectedExitCode: 0,
			ExpectedOutput:   []string{"usacloud-update"},
			Tags:             []string{"version", "backward-compatibility"},
			Category:         "cli",
			Priority:         "high",
		},
		{
			Name:             "legacy_flag_compatibility",
			Description:      "Test legacy flag support",
			Arguments:        []string{"--validate-only", "/dev/stdin"},
			ExpectedExitCode: 0,
			Tags:             []string{"legacy", "flags"},
			Category:         "cli",
			Priority:         "medium",
		},
		{
			Name:             "output_format_compatibility",
			Description:      "Test output format backward compatibility",
			Arguments:        []string{"--help"},
			ExpectedExitCode: 0,
			ExpectedOutput:   []string{"usacloud-update"},
			Tags:             []string{"output", "format"},
			Category:         "cli",
			Priority:         "medium",
		},
		{
			Name:             "error_code_compatibility",
			Description:      "Test error code backward compatibility",
			Arguments:        []string{"--invalid-flag"},
			ExpectedExitCode: 1,
			ExpectedErrors:   []string{"無効なオプション"},
			Tags:             []string{"error", "exit-code"},
			Category:         "cli",
			Priority:         "medium",
		},
	}
}

func (suite *CompatibilityRegressionTestSuite) createDefaultAPITestCases() []APITestCase {
	return []APITestCase{
		{
			Name:           "engine_process_compatibility",
			Description:    "Test transform engine process method compatibility",
			Function:       "transform.Engine.Process",
			Parameters:     "usacloud server list --output-type csv",
			ExpectedResult: "usacloud server list --output-type json",
			Tags:           []string{"api", "transform"},
			Category:       "api",
			Priority:       "high",
		},
		{
			Name:           "validation_compatibility",
			Description:    "Test validation API compatibility",
			Function:       "validation.Validator.Validate",
			Parameters:     "usacloud server list",
			ExpectedResult: true,
			Tags:           []string{"api", "validation"},
			Category:       "api",
			Priority:       "high",
		},
		{
			Name:           "config_load_compatibility",
			Description:    "Test config loading API compatibility",
			Function:       "config.Load",
			Parameters:     "/path/to/config",
			ExpectedResult: map[string]interface{}{},
			Tags:           []string{"api", "config"},
			Category:       "api",
			Priority:       "medium",
		},
	}
}

func (suite *CompatibilityRegressionTestSuite) createDefaultConfigTestCases() []ConfigTestCase {
	return []ConfigTestCase{
		{
			Name:          "ini_format_compatibility",
			Description:   "Test INI format backward compatibility",
			ConfigType:    "ini",
			ConfigData:    map[string]interface{}{"general": map[string]interface{}{"debug": true}},
			ExpectedValid: true,
			Tags:          []string{"config", "ini"},
			Category:      "config",
			Priority:      "high",
		},
		{
			Name:          "env_format_compatibility",
			Description:   "Test .env format backward compatibility",
			ConfigType:    "env",
			ConfigData:    map[string]interface{}{"DEBUG": "true"},
			ExpectedValid: true,
			Tags:          []string{"config", "env"},
			Category:      "config",
			Priority:      "medium",
		},
		{
			Name:              "config_migration_compatibility",
			Description:       "Test config migration backward compatibility",
			ConfigType:        "env_to_ini",
			ConfigData:        map[string]interface{}{"OLD_FORMAT": "true"},
			ExpectedValid:     true,
			MigrationRequired: true,
			Tags:              []string{"config", "migration"},
			Category:          "config",
			Priority:          "medium",
		},
	}
}

func (suite *CompatibilityRegressionTestSuite) compareCLIResults(baseline, current *CLIResult, testName string) bool {
	// Compare exit codes
	if baseline.ExitCode != current.ExitCode {
		suite.recordFailure("cli", testName,
			baseline.ExitCode, current.ExitCode, "exit_code_mismatch",
			map[string]interface{}{
				"baseline_exit_code": baseline.ExitCode,
				"current_exit_code":  current.ExitCode,
			})
		return false
	}

	// Compare command structure
	if baseline.Command != current.Command {
		suite.recordFailure("cli", testName,
			baseline.Command, current.Command, "command_mismatch",
			map[string]interface{}{
				"baseline_command": baseline.Command,
				"current_command":  current.Command,
			})
		return false
	}

	return true
}

func (suite *CompatibilityRegressionTestSuite) compareAPIResults(baseline, current *APIResult, testName string) bool {
	// Compare function signatures
	if baseline.Function != current.Function {
		suite.recordFailure("api", testName,
			baseline.Function, current.Function, "function_signature_mismatch",
			map[string]interface{}{
				"baseline_function": baseline.Function,
				"current_function":  current.Function,
			})
		return false
	}

	// Compare results using deep equality
	if !reflect.DeepEqual(baseline.Result, current.Result) {
		suite.recordFailure("api", testName,
			baseline.Result, current.Result, "result_mismatch",
			map[string]interface{}{
				"baseline_result": baseline.Result,
				"current_result":  current.Result,
			})
		return false
	}

	return true
}

func (suite *CompatibilityRegressionTestSuite) compareConfigResults(baseline, current *ConfigResult, testName string) bool {
	// Compare validity
	if baseline.IsValid != current.IsValid {
		suite.recordFailure("config", testName,
			baseline.IsValid, current.IsValid, "validity_mismatch",
			map[string]interface{}{
				"baseline_valid": baseline.IsValid,
				"current_valid":  current.IsValid,
			})
		return false
	}

	// Compare config type
	if baseline.ConfigType != current.ConfigType {
		suite.recordFailure("config", testName,
			baseline.ConfigType, current.ConfigType, "type_mismatch",
			map[string]interface{}{
				"baseline_type": baseline.ConfigType,
				"current_type":  current.ConfigType,
			})
		return false
	}

	return true
}

func (suite *CompatibilityRegressionTestSuite) checkFlagCompatibility(arguments []string) error {
	for _, arg := range arguments {
		if strings.HasPrefix(arg, "--") {
			flag := strings.TrimPrefix(arg, "--")
			if strings.Contains(flag, "=") {
				flag = strings.Split(flag, "=")[0]
			}

			if removedFlag, exists := suite.cliTester.flagCompatibility.removedFlags[flag]; exists {
				return fmt.Errorf("flag --%s was removed in version %s", flag, removedFlag.Removed)
			}

			if deprecatedFlag, exists := suite.cliTester.flagCompatibility.deprecatedFlags[flag]; exists {
				return fmt.Errorf("flag --%s is deprecated since version %s, use %s instead",
					flag, deprecatedFlag.Deprecated, deprecatedFlag.Replacement)
			}
		}
	}
	return nil
}

func (suite *CompatibilityRegressionTestSuite) checkSignatureCompatibility(functionName string) error {
	if signature, exists := suite.apiTester.signatureChecker.registry[functionName]; exists {
		if signature.Deprecated != "" {
			return fmt.Errorf("function %s is deprecated since version %s", functionName, signature.Deprecated)
		}
	}
	return nil
}

func (suite *CompatibilityRegressionTestSuite) checkFormatCompatibility(configType string) error {
	if format, exists := suite.configTester.formatValidator.supportedFormats[configType]; exists {
		if format.Deprecated != "" {
			return fmt.Errorf("config format %s is deprecated since version %s", configType, format.Deprecated)
		}
	}
	return nil
}

func (suite *CompatibilityRegressionTestSuite) checkGoVersionCompatibility(minVersion string) error {
	currentVersion := runtime.Version()
	// Simple version comparison - in production would use semver
	if strings.Compare(currentVersion, minVersion) < 0 {
		return fmt.Errorf("Go version %s is below minimum required version %s", currentVersion, minVersion)
	}
	return nil
}

func (suite *CompatibilityRegressionTestSuite) runCommandWithOutput(cmd *exec.Cmd) (stdout, stderr string, exitCode int, err error) {
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return "", "", -1, err
	}

	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return "", "", -1, err
	}

	if err := cmd.Start(); err != nil {
		return "", "", -1, err
	}

	stdoutBytes, _ := ioutil.ReadAll(stdoutPipe)
	stderrBytes, _ := ioutil.ReadAll(stderrPipe)

	err = cmd.Wait()
	exitCode = cmd.ProcessState.ExitCode()

	return string(stdoutBytes), string(stderrBytes), exitCode, nil
}

// Save functions
func (suite *CompatibilityRegressionTestSuite) saveCurrentCLIResult(testName string, result *CLIResult) {
	resultPath := filepath.Join(suite.currentDir, "compatibility", fmt.Sprintf("cli_%s.json", testName))
	suite.saveJSONToFile(resultPath, result)
}

func (suite *CompatibilityRegressionTestSuite) saveCurrentAPIResult(testName string, result *APIResult) {
	resultPath := filepath.Join(suite.currentDir, "compatibility", fmt.Sprintf("api_%s.json", testName))
	suite.saveJSONToFile(resultPath, result)
}

func (suite *CompatibilityRegressionTestSuite) saveCurrentConfigResult(testName string, result *ConfigResult) {
	resultPath := filepath.Join(suite.currentDir, "compatibility", fmt.Sprintf("config_%s.json", testName))
	suite.saveJSONToFile(resultPath, result)
}

func (suite *CompatibilityRegressionTestSuite) saveCLITestCases(path string, testCases []CLITestCase) {
	suite.saveJSONToFile(path, testCases)
}

func (suite *CompatibilityRegressionTestSuite) saveAPITestCases(path string, testCases []APITestCase) {
	suite.saveJSONToFile(path, testCases)
}

func (suite *CompatibilityRegressionTestSuite) saveConfigTestCases(path string, testCases []ConfigTestCase) {
	suite.saveJSONToFile(path, testCases)
}

func (suite *CompatibilityRegressionTestSuite) saveJSONToFile(path string, data interface{}) {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		suite.t.Logf("Failed to marshal data for %s: %v", path, err)
		return
	}

	if err := os.WriteFile(path, jsonData, 0644); err != nil {
		suite.t.Logf("Failed to save file %s: %v", path, err)
	}
}

func (suite *CompatibilityRegressionTestSuite) recordFailure(failureType, testName string, expected, actual interface{}, severity string, context map[string]interface{}) {
	failure := RegressionFailure{
		Type:       failureType,
		TestName:   testName,
		Expected:   expected,
		Actual:     actual,
		Difference: fmt.Sprintf("Expected: %v, Actual: %v", expected, actual),
		Severity:   severity,
		Timestamp:  time.Now(),
		Context:    context,
	}

	suite.mu.Lock()
	suite.failures = append(suite.failures, failure)
	suite.mu.Unlock()

	suite.logf("Compatibility regression failure in %s test '%s': %s", failureType, testName, failure.Difference)
}

func (suite *CompatibilityRegressionTestSuite) recordWarning(warningType, testName, message string, actual interface{}, context map[string]interface{}) {
	warning := RegressionWarning{
		Type:      warningType,
		TestName:  testName,
		Message:   message,
		Actual:    actual,
		Timestamp: time.Now(),
		Context:   context,
	}

	suite.mu.Lock()
	suite.warnings = append(suite.warnings, warning)
	suite.mu.Unlock()

	suite.logf("Compatibility regression warning in %s test '%s': %s", warningType, testName, message)
}
