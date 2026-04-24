# midimap - MIDI remapper scripted in Lua

BINARY   := midimap
PKG      := .
BUILD_DIR := bin
SCRIPT   ?=

# バージョン情報をビルドに埋め込む(git 管理下でなくても落ちないように)
VERSION  := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS  := -ldflags "-s -w -X main.version=$(VERSION)"

# デフォルト引数(make run で使用)。上書き可: make run IN=0 OUT=1
IN       ?=
OUT      ?=
ARGS     ?=

.DEFAULT_GOAL := help

.PHONY: help build run list install clean fmt vet tidy test deps dev all

## help: このヘルプを表示
help:
	@echo "midimap - Makefile targets"
	@echo ""
	@awk '/^## / { sub(/^## /, "  "); print }' $(MAKEFILE_LIST)
	@echo ""
	@echo "Examples:"
	@echo "  make build"
	@echo "  make run IN=0 OUT=1 SCRIPT=luascripts/example.lua"
	@echo "  make dev IN=0 OUT=1 SCRIPT=luascripts/example.lua"
	@echo "  make list"

## build: バイナリを bin/ にビルド
build:
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY) $(PKG)
	@echo "Built: $(BUILD_DIR)/$(BINARY) ($(VERSION))"

## run: ビルドして実行 (IN, OUT, SCRIPT で引数指定)
run: build
	@if [ -z "$(IN)" ] || [ -z "$(OUT)" ] || [ -z "$(SCRIPT)" ]; then \
		echo "Usage: make run IN=<port> OUT=<port> SCRIPT=<script.lua> [ARGS=\"-v\"]"; \
		echo "Run 'make list' to see available ports."; \
		exit 1; \
	fi
	./$(BUILD_DIR)/$(BINARY) -i "$(IN)" -o "$(OUT)" -s $(SCRIPT) $(ARGS)

## list: MIDI ポート一覧を表示
list: build
	./$(BUILD_DIR)/$(BINARY) -l

## dev: verbose 付きで実行 (make dev IN=0 OUT=1)
dev: ARGS := -v
dev: run

## install: $GOBIN (または $GOPATH/bin) にインストール
install:
	go install $(LDFLAGS) $(PKG)

## fmt: gofmt で整形
fmt:
	go fmt ./...

## vet: go vet で静的解析
vet:
	go vet ./...

## tidy: go.mod / go.sum を整理
tidy:
	go mod tidy

## deps: 依存を取得
deps:
	go mod download

## test: テストを実行
test:
	go test -v ./...

## clean: ビルド成果物を削除
clean:
	rm -rf $(BUILD_DIR)

## all: fmt, vet, build を順に実行
all: fmt vet build