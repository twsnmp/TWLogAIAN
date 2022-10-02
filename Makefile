.PHONY: all clean windows mac dev
### バージョンの定義
VERSION     := "v1.6.0"
COMMIT      := $(shell git rev-parse --short HEAD)

### コマンドの定義
LDFLAGS  = "-s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT)"

### PHONY ターゲットのビルドルール
all: windows mac

clean:
	rm -rf build/bin/TWLogAIAN*

windows:
	wails build  -platform windows -ldflags $(LDFLAGS)

windebug:
	wails build  -platform windows -debug

mac:
	wails build  -platform darwin -ldflags $(LDFLAGS)

dev:
	wails dev -e svelte,go

