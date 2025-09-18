# Makefile for usacloud-update
# Usage:
#   make build         # ビルド
#   make test          # 通常テスト（TUI無効、golden比較、30秒タイムアウト）
#   make test-long     # 長時間テスト（TUI無効、30分タイムアウト）
#   make test-tui      # TUI関連テスト（インタラクティブ環境向け）
#   make bdd           # BDDテスト（サンドボックス機能、30秒タイムアウト）
#   make golden        # 期待値(golden)を最新出力で上書き更新（TUI無効）
#   make verify-sample # サンプルを実行して期待値とdiff確認
#   make install       # $GOPATH/bin にインストール
#   make uninstall     # $GOPATH/bin から削除
#   make tidy fmt vet  # 開発補助
#   make clean         # 生成物掃除
#   TEST_TIMEOUT=600s make test # カスタムタイムアウト設定例

GO       ?= go
BINARY   := usacloud-update
BIN_DIR  := bin
CMD_PKG  := ./cmd/$(BINARY)
PKGS     := ./...

# テストタイムアウト設定（環境変数で上書き可能）
TEST_TIMEOUT ?= 30s

IN_SAMPLE  := testdata/sample_v0_v1_mixed.sh
OUT_SAMPLE := /tmp/out.sh
GOLDEN     := testdata/expected_v1_1.sh

IN_MIXED   := testdata/mixed_with_non_usacloud.sh
OUT_MIXED  := /tmp/out_mixed.sh
GOLDEN_MIXED := testdata/expected_mixed_non_usacloud.sh

.PHONY: all build run test test-long test-tui bdd golden verify-sample verify-mixed install uninstall tidy fmt vet clean

all: build

build: tidy fmt
	$(GO) build -o $(BIN_DIR)/$(BINARY) $(CMD_PKG)

run: build
	$(BIN_DIR)/$(BINARY) --in $(IN_SAMPLE) --out $(OUT_SAMPLE)

test:
	USACLOUD_UPDATE_NO_TUI=true $(GO) test -timeout $(TEST_TIMEOUT) $(PKGS)

# 長時間テスト用（30分タイムアウト）
test-long:
	USACLOUD_UPDATE_NO_TUI=true $(GO) test -timeout 1800s $(PKGS)

# TUI関連テスト（インタラクティブ環境向け）
test-tui:
	$(GO) test -timeout $(TEST_TIMEOUT) -run TUI $(PKGS)

# BDD機能テスト（サンドボックス機能のBDD）
bdd:
	$(GO) test -timeout $(TEST_TIMEOUT) ./internal/bdd -godog.format=pretty -godog.paths=features

# Goldenファイル(期待値)を「仕様変更後の正しい出力」で更新
golden:
	USACLOUD_UPDATE_NO_TUI=true $(GO) test -timeout $(TEST_TIMEOUT) -run Golden -update $(PKGS)

# サンプル入力を変換して期待値と比較（手元確認）
verify-sample: run
	diff -u $(GOLDEN) $(OUT_SAMPLE) || true

# 非usacloud行混在テストの実行と期待値比較
verify-mixed: build
	$(BIN_DIR)/$(BINARY) --in $(IN_MIXED) --out $(OUT_MIXED)
	diff -u $(GOLDEN_MIXED) $(OUT_MIXED) || true

tidy:
	$(GO) mod tidy

fmt:
	$(GO) fmt $(PKGS)

vet:
	$(GO) vet $(PKGS)

# ユーザーのGOPATH/binにインストール
install:
	$(GO) install $(CMD_PKG)
	@echo "✅ $(BINARY) を $(shell go env GOPATH)/bin にインストールしました"

# アンインストール
uninstall:
	rm -f $(shell go env GOPATH)/bin/$(BINARY)
	@echo "🗑️  $(BINARY) をアンインストールしました"

clean:
	rm -rf $(BIN_DIR)

