package transform

import (
	"bufio"
	"flag"
	"os"
	"strings"
	"testing"
)

var update = flag.Bool("update", false, "update golden files")

// applyFile reads a bash script, applies the transform engine line-by-line,
// and returns the final output string with the generated header and trailing newline.
func applyFile(t *testing.T, inPath string) string {
	t.Helper()

	f, err := os.Open(inPath)
	if err != nil {
		t.Fatalf("open input %s: %v", inPath, err)
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	// allow long lines
	sc.Buffer(make([]byte, 0, 1024*1024), 1024*1024)

	eng := NewDefaultEngine()
	var outLines []string
	for sc.Scan() {
		res := eng.Apply(sc.Text())
		outLines = append(outLines, res.Line)
	}
	if err := sc.Err(); err != nil {
		t.Fatalf("scan %s: %v", inPath, err)
	}

	// Join with LF and ensure terminating newline
	return strings.Join(append([]string{GeneratedHeader()}, outLines...), "\n") + "\n"
}

func TestGolden_SampleMixed(t *testing.T) {
	inPath := "../../testdata/sample_v0_v1_mixed.sh"
	wantPath := "../../testdata/expected_v1_1.sh"

	got := applyFile(t, inPath)

	if *update {
		if err := os.WriteFile(wantPath, []byte(got), 0o644); err != nil {
			t.Fatalf("update golden %s: %v", wantPath, err)
		}
		return
	}

	wantBytes, err := os.ReadFile(wantPath)
	if err != nil {
		t.Fatalf("open expected %s: %v", wantPath, err)
	}
	want := string(wantBytes)

	if got != want {
		t.Errorf("golden mismatch.\n--- want ---\n%s\n--- got ---\n%s", want, got)
	}
}
