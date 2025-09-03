package transform

import (
	"fmt"
	"regexp"
	"strings"
)

type simpleRule struct {
	name   string
	re     *regexp.Regexp
	repl   func([]string) string
	reason string
	url    string
}

func (r *simpleRule) Name() string { return r.name }

func (r *simpleRule) Apply(line string) (string, bool, string, string) {
	m := r.re.FindStringSubmatch(line)
	if m == nil {
		return line, false, "", ""
	}
	after := r.re.ReplaceAllString(line, r.repl(m))
	comment := fmt.Sprintf(" # sacloud-update: %s (%s)", r.reason, r.url)
	if !strings.Contains(after, "# sacloud-update:") {
		after += comment
	}
	beforeFrag := strings.TrimSpace(m[0])
	afterFrag := strings.TrimSpace(r.repl(m))
	return after, true, beforeFrag, afterFrag
}

// helper to build rule
func mk(name, pattern string, repl func([]string) string, reason, url string) Rule {
	return &simpleRule{name: name, re: regexp.MustCompile(pattern), repl: repl, reason: reason, url: url}
}
