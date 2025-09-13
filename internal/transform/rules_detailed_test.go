package transform

import (
	"regexp"
	"strings"
	"testing"
)

func TestSimpleRule_Name(t *testing.T) {
	rule := &simpleRule{name: "test-rule"}
	if rule.Name() != "test-rule" {
		t.Errorf("Expected name 'test-rule', got %s", rule.Name())
	}
}

func TestSimpleRule_Apply_NoMatch(t *testing.T) {
	rule := &simpleRule{
		name: "test-rule",
		re:   regexp.MustCompile(`test-pattern`),
		repl: func(m []string) string { return "replacement" },
	}

	line, changed, before, after := rule.Apply("no match here")

	if changed {
		t.Errorf("Expected no change for non-matching line")
	}
	if line != "no match here" {
		t.Errorf("Expected unchanged line, got %s", line)
	}
	if before != "" || after != "" {
		t.Errorf("Expected empty before/after for non-matching line")
	}
}

func TestSimpleRule_Apply_WithMatch(t *testing.T) {
	rule := &simpleRule{
		name: "test-rule",
		re:   regexp.MustCompile(`(test)-pattern`),
		repl: func(m []string) string {
			return strings.Replace(m[0], "test", "new", 1)
		},
		reason: "test replacement",
		url:    "https://example.com",
	}

	line, changed, before, after := rule.Apply("this is test-pattern here")

	if !changed {
		t.Errorf("Expected change for matching line")
	}

	expectedLine := "this is new-pattern here # usacloud-update: test replacement (https://example.com)"
	if line != expectedLine {
		t.Errorf("Expected line %s, got %s", expectedLine, line)
	}

	if before != "test-pattern" {
		t.Errorf("Expected before fragment 'test-pattern', got %s", before)
	}

	if after != "new-pattern" {
		t.Errorf("Expected after fragment 'new-pattern', got %s", after)
	}
}

func TestSimpleRule_Apply_PreservesExistingComment(t *testing.T) {
	rule := &simpleRule{
		name:   "test-rule",
		re:     regexp.MustCompile(`test-pattern`),
		repl:   func(m []string) string { return "new-pattern" },
		reason: "test replacement",
		url:    "https://example.com",
	}

	input := "test-pattern # usacloud-update: existing comment"
	line, changed, _, _ := rule.Apply(input)

	if !changed {
		t.Errorf("Expected change for matching line")
	}

	// Should not add duplicate comment
	commentCount := strings.Count(line, "# usacloud-update:")
	if commentCount != 1 {
		t.Errorf("Expected exactly 1 usacloud-update comment, got %d in line: %s", commentCount, line)
	}
}

func TestMkFunction(t *testing.T) {
	rule := mk("test-rule", `test-(\w+)`, func(m []string) string {
		return "new-" + m[1]
	}, "test reason", "https://example.com")

	if rule.Name() != "test-rule" {
		t.Errorf("Expected rule name 'test-rule', got %s", rule.Name())
	}

	line, changed, before, after := rule.Apply("test-value")

	if !changed {
		t.Errorf("Expected change for matching pattern")
	}

	if before != "test-value" {
		t.Errorf("Expected before fragment 'test-value', got %s", before)
	}

	if after != "new-value" {
		t.Errorf("Expected after fragment 'new-value', got %s", after)
	}

	expectedLine := "new-value # usacloud-update: test reason (https://example.com)"
	if line != expectedLine {
		t.Errorf("Expected line %s, got %s", expectedLine, line)
	}
}

func TestRuleWithComplexReplacement(t *testing.T) {
	rule := mk("complex-rule",
		`usacloud\s+(\w+)\s+list\s+--output-type=(\w+)`,
		func(m []string) string {
			if m[2] == "csv" || m[2] == "tsv" {
				return strings.Replace(m[0], m[2], "json", 1)
			}
			return m[0]
		},
		"output format change",
		"https://docs.example.com")

	// Test CSV replacement
	line, changed, _, _ := rule.Apply("usacloud server list --output-type=csv")
	if !changed {
		t.Errorf("Expected change for CSV input")
	}
	if !strings.Contains(line, "--output-type=json") {
		t.Errorf("Expected CSV to be replaced with json, got: %s", line)
	}

	// Test TSV replacement
	line, changed, _, _ = rule.Apply("usacloud server list --output-type=tsv")
	if !changed {
		t.Errorf("Expected change for TSV input")
	}
	if !strings.Contains(line, "--output-type=json") {
		t.Errorf("Expected TSV to be replaced with json, got: %s", line)
	}

	// Test JSON input - rule will still match but replacement function returns same value
	line, changed, _, _ = rule.Apply("usacloud server list --output-type=json")
	// The rule pattern matches but replacement function returns the same value
	// This still counts as "changed" in the rule system even if value is same
	if !changed {
		t.Errorf("Rule matches JSON input but replacement returns same value")
	}
}

func TestRuleChaining(t *testing.T) {
	// Test that multiple rules can be applied sequentially
	rules := []Rule{
		mk("rule1", `oldval`, func(m []string) string { return "midval" }, "first change", "https://example1.com"),
		mk("rule2", `midval`, func(m []string) string { return "newval" }, "second change", "https://example2.com"),
	}

	engine := &Engine{rules: rules}
	result := engine.Apply("this has oldval in it")

	if !result.Changed {
		t.Errorf("Expected change when applying chained rules")
	}

	if !strings.Contains(result.Line, "newval") {
		t.Errorf("Expected final value 'newval' in result: %s", result.Line)
	}

	if len(result.Changes) != 2 {
		t.Errorf("Expected 2 changes recorded, got %d", len(result.Changes))
	}

	// Verify change details
	if result.Changes[0].RuleName != "rule1" {
		t.Errorf("Expected first change rule name 'rule1', got %s", result.Changes[0].RuleName)
	}
	if result.Changes[1].RuleName != "rule2" {
		t.Errorf("Expected second change rule name 'rule2', got %s", result.Changes[1].RuleName)
	}
}
