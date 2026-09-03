.PHONY: all clean windows mac linux dev
### バージョンの定義
VERSION     := "v2.0.0"
COMMIT      := $(shell git rev-parse --short HEAD)

### コマンドの定義
TAGS     := wgpu24
LDFLAGS  = "-s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT)"

### PHONY ターゲットのビルドルール
all: windows mac linux

clean:
	rm -rf build/bin/TWLogAIAN*

windows:
	wails build  -platform windows/amd64 -clean -tags $(TAGS) -ldflags $(LDFLAGS)

windebug:
	wails build  -platform windows/amd64 -clean -tags $(TAGS) -debug

mac:
	wails build  -platform darwin/universal -clean -tags $(TAGS) -ldflags $(LDFLAGS)

linux:
	wails build  -platform linux/amd64 -clean -tags "webkit2_41,$(TAGS)" -ldflags $(LDFLAGS)

dev:
	wails dev -tags $(TAGS)

