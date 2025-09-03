package transform

import (
	"regexp"
	"strings"
)

type Change struct {
	RuleName string
	Before   string
	After    string
}

type Result struct {
	Line    string
	Changed bool
	Changes []Change
}

type Rule interface {
	Name() string
	Apply(line string) (string, bool, string, string)
}

type Engine struct{ rules []Rule }

func NewDefaultEngine() *Engine {
	return &Engine{rules: DefaultRules()}
}

func (e *Engine) Apply(line string) Result {
	// コメント/空行はスキップ
	trim := strings.TrimSpace(line)
	if trim == "" || strings.HasPrefix(trim, "#") {
		return Result{Line: line}
	}

	changed := false
	var changes []Change
	cur := line
	for _, r := range e.rules {
		after, ok, beforeFrag, afterFrag := r.Apply(cur)
		if ok {
			changed = true
			changes = append(changes, Change{RuleName: r.Name(), Before: beforeFrag, After: afterFrag})
			cur = after
		}
	}
	return Result{Line: cur, Changed: changed, Changes: changes}
}

// utilities
var reSpaces = regexp.MustCompile(`\s+`)
