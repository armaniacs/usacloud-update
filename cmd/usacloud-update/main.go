package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/armaniacs/usacloud-update/internal/transform"
	"github.com/fatih/color"
)

const version = "1.1.0"

var (
	inFile      = flag.String("in", "-", "input file path ('-' for stdin)")
	outFile     = flag.String("out", "-", "output file path ('-' for stdout)")
	stats       = flag.Bool("stats", true, "print summary of changes to stderr")
	showVersion = flag.Bool("version", false, "show version")
)

func main() {
	flag.Parse()

	if *showVersion {
		fmt.Printf("usacloud-update v%s\n", version)
		os.Exit(0)
	}

	var r io.Reader = os.Stdin
	if *inFile != "-" {
		f, e := os.Open(*inFile)
		if e != nil {
			panic(e)
		}
		defer f.Close()
		r = f
	}

	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 1024*1024), 1024*1024)

	eng := transform.NewDefaultEngine()
	var outLines []string
	for lineNum := 1; scanner.Scan(); lineNum++ {
		line := scanner.Text()
		res := eng.Apply(line)
		if res.Changed {
			for _, c := range res.Changes {
				if *stats {
					fmt.Fprintf(os.Stderr, color.YellowString("#L%-5d %s => %s [%s]\n"), lineNum, c.Before, c.After, c.RuleName)
				}
			}
		}
		outLines = append(outLines, res.Line)
	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}

	output := strings.Join(append([]string{transform.GeneratedHeader()}, outLines...), "\n") + "\n"

	var w io.Writer = os.Stdout
	if *outFile != "-" {
		f, e := os.Create(*outFile)
		if e != nil {
			panic(e)
		}
		defer f.Close()
		w = f
	}
	if _, err := io.WriteString(w, output); err != nil {
		panic(err)
	}
}
