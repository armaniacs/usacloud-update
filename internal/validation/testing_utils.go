package validation

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

type TestHelper struct {
	t       *testing.T
	dataDir string
	tempDir string
}

func NewTestHelper(t *testing.T) *TestHelper {
	return &TestHelper{
		t:       t,
		dataDir: "testdata",
		tempDir: t.TempDir(),
	}
}

func (th *TestHelper) LoadTestData(filename string, v interface{}) {
	path := filepath.Join(th.dataDir, filename)
	data, err := os.ReadFile(path)
	if err != nil {
		th.t.Fatalf("テストデータ読み込みエラー %s: %v", path, err)
	}

	if err := json.Unmarshal(data, v); err != nil {
		th.t.Fatalf("JSONパースエラー %s: %v", path, err)
	}
}

func (th *TestHelper) CreateTestData(filename string, data interface{}) {
	path := filepath.Join(th.tempDir, filename)
	dir := filepath.Dir(path)

	if err := os.MkdirAll(dir, 0755); err != nil {
		th.t.Fatalf("テストディレクトリ作成エラー %s: %v", dir, err)
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		th.t.Fatalf("JSON生成エラー: %v", err)
	}

	if err := os.WriteFile(path, jsonData, 0644); err != nil {
		th.t.Fatalf("テストファイル作成エラー %s: %v", path, err)
	}
}

func (th *TestHelper) AssertEqual(got, want interface{}, msgAndArgs ...interface{}) {
	th.t.Helper()
	if got != want {
		msg := fmt.Sprintf("Expected: %v, Got: %v", want, got)
		if len(msgAndArgs) > 0 {
			msg = fmt.Sprintf(msgAndArgs[0].(string), msgAndArgs[1:]...) + " - " + msg
		}
		th.t.Error(msg)
	}
}

func (th *TestHelper) AssertNotEqual(got, want interface{}, msgAndArgs ...interface{}) {
	th.t.Helper()
	if got == want {
		msg := fmt.Sprintf("Expected values to be different, but both were: %v", got)
		if len(msgAndArgs) > 0 {
			msg = fmt.Sprintf(msgAndArgs[0].(string), msgAndArgs[1:]...) + " - " + msg
		}
		th.t.Error(msg)
	}
}

func (th *TestHelper) AssertTrue(condition bool, msgAndArgs ...interface{}) {
	th.t.Helper()
	if !condition {
		msg := "Expected condition to be true"
		if len(msgAndArgs) > 0 {
			msg = fmt.Sprintf(msgAndArgs[0].(string), msgAndArgs[1:]...)
		}
		th.t.Error(msg)
	}
}

func (th *TestHelper) AssertFalse(condition bool, msgAndArgs ...interface{}) {
	th.t.Helper()
	if condition {
		msg := "Expected condition to be false"
		if len(msgAndArgs) > 0 {
			msg = fmt.Sprintf(msgAndArgs[0].(string), msgAndArgs[1:]...)
		}
		th.t.Error(msg)
	}
}

func (th *TestHelper) AssertContains(haystack, needle string, msgAndArgs ...interface{}) {
	th.t.Helper()
	if !strings.Contains(haystack, needle) {
		msg := fmt.Sprintf("Expected '%s' to contain '%s'", haystack, needle)
		if len(msgAndArgs) > 0 {
			msg = fmt.Sprintf(msgAndArgs[0].(string), msgAndArgs[1:]...) + " - " + msg
		}
		th.t.Error(msg)
	}
}

func (th *TestHelper) AssertNotContains(haystack, needle string, msgAndArgs ...interface{}) {
	th.t.Helper()
	if strings.Contains(haystack, needle) {
		msg := fmt.Sprintf("Expected '%s' to not contain '%s'", haystack, needle)
		if len(msgAndArgs) > 0 {
			msg = fmt.Sprintf(msgAndArgs[0].(string), msgAndArgs[1:]...) + " - " + msg
		}
		th.t.Error(msg)
	}
}

func (th *TestHelper) AssertNoError(err error, msgAndArgs ...interface{}) {
	th.t.Helper()
	if err != nil {
		msg := fmt.Sprintf("Expected no error, got: %v", err)
		if len(msgAndArgs) > 0 {
			msg = fmt.Sprintf(msgAndArgs[0].(string), msgAndArgs[1:]...) + " - " + msg
		}
		th.t.Error(msg)
	}
}

func (th *TestHelper) AssertError(err error, msgAndArgs ...interface{}) {
	th.t.Helper()
	if err == nil {
		msg := "Expected an error, got nil"
		if len(msgAndArgs) > 0 {
			msg = fmt.Sprintf(msgAndArgs[0].(string), msgAndArgs[1:]...)
		}
		th.t.Error(msg)
	}
}

func (th *TestHelper) AssertErrorContains(err error, expectedMsg string, msgAndArgs ...interface{}) {
	th.t.Helper()
	if err == nil {
		msg := fmt.Sprintf("Expected error containing '%s', got nil", expectedMsg)
		if len(msgAndArgs) > 0 {
			msg = fmt.Sprintf(msgAndArgs[0].(string), msgAndArgs[1:]...) + " - " + msg
		}
		th.t.Error(msg)
		return
	}

	if !strings.Contains(err.Error(), expectedMsg) {
		msg := fmt.Sprintf("Expected error containing '%s', got: %v", expectedMsg, err)
		if len(msgAndArgs) > 0 {
			msg = fmt.Sprintf(msgAndArgs[0].(string), msgAndArgs[1:]...) + " - " + msg
		}
		th.t.Error(msg)
	}
}

func (th *TestHelper) AssertSliceEqual(got, want []string, msgAndArgs ...interface{}) {
	th.t.Helper()
	if len(got) != len(want) {
		msg := fmt.Sprintf("Expected slice length %d, got %d", len(want), len(got))
		if len(msgAndArgs) > 0 {
			msg = fmt.Sprintf(msgAndArgs[0].(string), msgAndArgs[1:]...) + " - " + msg
		}
		th.t.Error(msg)
		return
	}

	for i, v := range want {
		if got[i] != v {
			msg := fmt.Sprintf("Expected slice[%d]='%s', got '%s'", i, v, got[i])
			if len(msgAndArgs) > 0 {
				msg = fmt.Sprintf(msgAndArgs[0].(string), msgAndArgs[1:]...) + " - " + msg
			}
			th.t.Error(msg)
		}
	}
}

func (th *TestHelper) AssertSliceContains(slice []string, element string, msgAndArgs ...interface{}) {
	th.t.Helper()
	for _, v := range slice {
		if v == element {
			return
		}
	}

	msg := fmt.Sprintf("Expected slice to contain '%s', got: %v", element, slice)
	if len(msgAndArgs) > 0 {
		msg = fmt.Sprintf(msgAndArgs[0].(string), msgAndArgs[1:]...) + " - " + msg
	}
	th.t.Error(msg)
}

func (th *TestHelper) AssertGreater(got, want int, msgAndArgs ...interface{}) {
	th.t.Helper()
	if got <= want {
		msg := fmt.Sprintf("Expected %d > %d", got, want)
		if len(msgAndArgs) > 0 {
			msg = fmt.Sprintf(msgAndArgs[0].(string), msgAndArgs[1:]...) + " - " + msg
		}
		th.t.Error(msg)
	}
}

func (th *TestHelper) AssertLess(got, want int, msgAndArgs ...interface{}) {
	th.t.Helper()
	if got >= want {
		msg := fmt.Sprintf("Expected %d < %d", got, want)
		if len(msgAndArgs) > 0 {
			msg = fmt.Sprintf(msgAndArgs[0].(string), msgAndArgs[1:]...) + " - " + msg
		}
		th.t.Error(msg)
	}
}

func (th *TestHelper) AssertGreaterOrEqual(got, want int, msgAndArgs ...interface{}) {
	th.t.Helper()
	if got < want {
		msg := fmt.Sprintf("Expected %d >= %d", got, want)
		if len(msgAndArgs) > 0 {
			msg = fmt.Sprintf(msgAndArgs[0].(string), msgAndArgs[1:]...) + " - " + msg
		}
		th.t.Error(msg)
	}
}

func (th *TestHelper) AssertWithinDuration(expected, actual time.Time, delta time.Duration, msgAndArgs ...interface{}) {
	th.t.Helper()
	diff := actual.Sub(expected)
	if diff < 0 {
		diff = -diff
	}

	if diff > delta {
		msg := fmt.Sprintf("Expected time difference within %v, got %v", delta, diff)
		if len(msgAndArgs) > 0 {
			msg = fmt.Sprintf(msgAndArgs[0].(string), msgAndArgs[1:]...) + " - " + msg
		}
		th.t.Error(msg)
	}
}

type BenchmarkHelper struct {
	dataCache map[string]interface{}
}

func NewBenchmarkHelper() *BenchmarkHelper {
	return &BenchmarkHelper{
		dataCache: make(map[string]interface{}),
	}
}

func (bh *BenchmarkHelper) LoadBenchmarkData(filename string) interface{} {
	if data, exists := bh.dataCache[filename]; exists {
		return data
	}

	path := filepath.Join("testdata", filename)
	fileData, err := os.ReadFile(path)
	if err != nil {
		return nil
	}

	var data interface{}
	if err := json.Unmarshal(fileData, &data); err != nil {
		return nil
	}

	bh.dataCache[filename] = data
	return data
}

func (bh *BenchmarkHelper) GetTestCommands() []string {
	data := bh.LoadBenchmarkData("commands/benchmark_commands.json")
	if commands, ok := data.([]interface{}); ok {
		result := make([]string, len(commands))
		for i, cmd := range commands {
			if str, ok := cmd.(string); ok {
				result[i] = str
			}
		}
		return result
	}
	return []string{"server", "disk", "database", "loadbalancer"}
}

func (bh *BenchmarkHelper) GetTestInputs() []string {
	return []string{
		"serv", "dsk", "databse", "loadbalancer",
		"invalidcommand", "very-long-invalid-command-name",
		"", "x", "ab", "xyz",
	}
}

type MockTime struct {
	currentTime time.Time
}

func NewMockTime(t time.Time) *MockTime {
	return &MockTime{currentTime: t}
}

func (m *MockTime) Now() time.Time {
	return m.currentTime
}

func (m *MockTime) Advance(d time.Duration) {
	m.currentTime = m.currentTime.Add(d)
}

type TestCase struct {
	Name        string
	Input       interface{}
	Expected    interface{}
	ShouldError bool
	ErrorMsg    string
}

func RunTestCases(t *testing.T, testCases []TestCase, testFunc func(interface{}) (interface{}, error)) {
	helper := NewTestHelper(t)

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			result, err := testFunc(tc.Input)

			if tc.ShouldError {
				helper.AssertError(err, "Test case '%s' should return error", tc.Name)
				if tc.ErrorMsg != "" {
					helper.AssertErrorContains(err, tc.ErrorMsg, "Error message should contain expected text")
				}
			} else {
				helper.AssertNoError(err, "Test case '%s' should not return error", tc.Name)
				helper.AssertEqual(result, tc.Expected, "Test case '%s' result", tc.Name)
			}
		})
	}
}

type PerformanceChecker struct {
	maxDuration time.Duration
}

func NewPerformanceChecker(maxDuration time.Duration) *PerformanceChecker {
	return &PerformanceChecker{maxDuration: maxDuration}
}

func (pc *PerformanceChecker) CheckPerformance(t *testing.T, name string, fn func()) {
	start := time.Now()
	fn()
	duration := time.Since(start)

	if duration > pc.maxDuration {
		t.Errorf("Performance check '%s' exceeded maximum duration: %v > %v",
			name, duration, pc.maxDuration)
	}
}

type TableTest struct {
	tests []TestCase
}

func NewTableTest() *TableTest {
	return &TableTest{tests: make([]TestCase, 0)}
}

func (tt *TableTest) Add(name string, input, expected interface{}) *TableTest {
	tt.tests = append(tt.tests, TestCase{
		Name:     name,
		Input:    input,
		Expected: expected,
	})
	return tt
}

func (tt *TableTest) AddError(name string, input interface{}, errorMsg string) *TableTest {
	tt.tests = append(tt.tests, TestCase{
		Name:        name,
		Input:       input,
		ShouldError: true,
		ErrorMsg:    errorMsg,
	})
	return tt
}

func (tt *TableTest) Run(t *testing.T, testFunc func(interface{}) (interface{}, error)) {
	RunTestCases(t, tt.tests, testFunc)
}
