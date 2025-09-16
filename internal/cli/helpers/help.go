package helpers

import (
	"fmt"
	"os"

	"github.com/fatih/color"
)

// GetHelpContent returns the main help content
func GetHelpContent(version string) string {
	return fmt.Sprintf(`usacloud-update v%s

概要:
  usacloud v0、v1.0、v1.1の記述が混在したbashスクリプトを、v1.1で動作するように自動変換します。
  廃止されたオプション、変更されたリソース名、新しいコマンド引数形式などを自動更新し、
  変換できない箇所は適切なコメントと共に手動対応を促します。

  --sandboxオプションでSakura Cloudサンドボックス環境での実際のコマンド実行が可能です。

使用方法:
  usacloud-update [オプション]

基本的な使用例:
  # パイプラインで使用
  cat input.sh | usacloud-update > output.sh

  # ファイルを指定して変換
  usacloud-update --in script.sh --out updated_script.sh

  # 変更統計のみ確認（出力は破棄）
  usacloud-update --in script.sh --out /dev/null

  # 統計出力を無効にして変換
  usacloud-update --in script.sh --out updated.sh --stats=false

サンドボックス機能の使用例:
  # インタラクティブTUIでサンドボックス実行
  usacloud-update --sandbox --in script.sh

  # ドライランモード（実行せずに結果確認）
  usacloud-update --sandbox --dry-run --in script.sh

  # バッチモード（全コマンド自動実行）
  usacloud-update --sandbox --batch --in script.sh

  # TUIなしで直接バッチ実行
  usacloud-update --sandbox --interactive=false --batch --in script.sh

環境設定:
  サンドボックス機能を使用するには設定ファイルまたは環境変数が必要です:

  【推奨】設定ファイル方式:
    usacloud-update.conf.sample を参考に ~/.config/usacloud-update/usacloud-update.conf を作成
    初回実行時に対話的に作成することも可能

    設定ファイルディレクトリのカスタマイズ:
      USACLOUD_UPDATE_CONFIG_DIR=/path/to/config - カスタム設定ディレクトリを指定

  環境変数方式（レガシー）:
    SAKURACLOUD_ACCESS_TOKEN、SAKURACLOUD_ACCESS_TOKEN_SECRET`, version)
}

// GetOptionsContent returns the options help content
func GetOptionsContent() string {
	return `

オプション:
  --batch
        バッチモード: 選択した全コマンドを自動実行
  --color
        カラー出力を有効にする (default true)
  --config string
        設定ファイルパス（指定しない場合はデフォルト設定を使用）
  --dry-run
        実際の実行を行わず変換結果のみ表示
  --help
        ヘルプメッセージを表示
  --help-mode string
        ヘルプモード (basic/enhanced/interactive) (default "enhanced")
  --in string
        入力ファイルパス ('-'で標準入力) (default "-")
  --interactive
        インタラクティブTUIモード (sandboxとの組み合わせで使用) (default true)
  --interactive-mode
        インタラクティブ検証・修正モード
  --language string
        言語設定 (ja/en) (default "ja")
  --out string
        出力ファイルパス ('-'で標準出力) (default "-")
  --sandbox
        サンドボックス環境での実際のコマンド実行
  --skip-deprecated
        廃止コマンド警告をスキップ
  --stats
        変更の統計情報を標準エラー出力に表示 (default true)
  --strict-validation
        厳格検証モード（エラー発生時に処理を停止）
  --suggestion-level int
        提案レベル設定 (1-5) (default 3)
  --validate-only
        検証のみ実行（変換は行わない）
  --version
        バージョン情報を表示

`
}

// GetFooterContent returns the footer help content
func GetFooterContent() string {
	return `詳細な使用方法とルールについては README-Usage.md を参照してください。

バグ報告・機能要望: https://github.com/armaniacs/usacloud-update/issues
`
}

// FatalError prints an error message in red and exits with code 1
func FatalError(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, color.RedString(format), args...)
	if len(format) > 0 && format[len(format)-1] != '\n' {
		fmt.Fprint(os.Stderr, "\n")
	}
	os.Exit(1)
}

// PrintError prints an error message in red to stderr without exiting
func PrintError(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, color.RedString(format), args...)
	if len(format) > 0 && format[len(format)-1] != '\n' {
		fmt.Fprint(os.Stderr, "\n")
	}
}

// PrintWarning prints a warning message in yellow to stderr
func PrintWarning(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, color.YellowString(format), args...)
	if len(format) > 0 && format[len(format)-1] != '\n' {
		fmt.Fprint(os.Stderr, "\n")
	}
}

// PrintSuccess prints a success message in green to stderr
func PrintSuccess(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, color.GreenString(format), args...)
	if len(format) > 0 && format[len(format)-1] != '\n' {
		fmt.Fprint(os.Stderr, "\n")
	}
}

// PrintInfo prints an informational message in cyan to stderr
func PrintInfo(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, color.CyanString(format), args...)
	if len(format) > 0 && format[len(format)-1] != '\n' {
		fmt.Fprint(os.Stderr, "\n")
	}
}
