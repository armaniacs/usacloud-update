# Makefile for sacloud-update
# Usage:
#   make build         # ビルド
#   make test          # 通常テスト（golden比較）
#   make golden        # 期待値(golden)を最新出力で上書き更新
#   make verify-sample # サンプルを実行して期待値とdiff確認
#   make tidy fmt vet  # 開発補助
#   make clean         # 生成物掃除

GO       ?= go
BINARY   := sacloud-update
BIN_DIR  := bin
CMD_PKG  := ./cmd/$(BINARY)
PKGS     := ./...

IN_SAMPLE  := testdata/sample_v0_v1_mixed.sh
OUT_SAMPLE := /tmp/out.sh
GOLDEN     := testdata/expected_v1_1.sh

.PHONY: all build run test golden verify-sample tidy fmt vet clean

all: build

build: tidy fmt
	$(GO) build -o $(BIN_DIR)/$(BINARY) $(CMD_PKG)

run: build
	$(BIN_DIR)/$(BINARY) --in $(IN_SAMPLE) --out $(OUT_SAMPLE)

test:
	$(GO) test $(PKGS)

# Goldenファイル(期待値)を「仕様変更後の正しい出力」で更新
golden:
	$(GO) test -run Golden -update $(PKGS)

# サンプル入力を変換して期待値と比較（手元確認）
verify-sample: run
	diff -u $(GOLDEN) $(OUT_SAMPLE) || true

tidy:
	$(GO) mod tidy

fmt:
	$(GO) fmt $(PKGS)

vet:
	$(GO) vet $(PKGS)

clean:
	rm -rf $(BIN_DIR)

