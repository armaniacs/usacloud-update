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

const version = "1.2.0"

var (
	inFile      = flag.String("in", "-", "入力ファイルパス ('-'で標準入力)")
	outFile     = flag.String("out", "-", "出力ファイルパス ('-'で標準出力)")
	stats       = flag.Bool("stats", true, "変更の統計情報を標準エラー出力に表示")
	showVersion = flag.Bool("version", false, "バージョン情報を表示")
)

func init() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `usacloud-update v%s

概要:
  usacloud v0、v1.0、v1.1の記述が混在したbashスクリプトを、v1.1で動作するように自動変換します。
  廃止されたオプション、変更されたリソース名、新しいコマンド引数形式などを自動更新し、
  変換できない箇所は適切なコメントと共に手動対応を促します。

使用方法:
  usacloud-update [オプション]

使用例:
  # パイプラインで使用
  cat input.sh | usacloud-update > output.sh

  # ファイルを指定して変換
  usacloud-update --in script.sh --out updated_script.sh

  # 変更統計のみ確認（出力は破棄）
  usacloud-update --in script.sh --out /dev/null

  # 統計出力を無効にして変換
  usacloud-update --in script.sh --out updated.sh --stats=false

オプション:
`, version)
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, `
詳細な使用方法とルールについては README-Usage.md を参照してください。
`)
	}
}

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
