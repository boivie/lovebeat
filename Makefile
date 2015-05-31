all: lovebeat

ASSETS := $(shell find data -print)
BINDATA_DEBUG ?=

dashboard/assets.go: $(ASSETS)
	go-bindata $(BINDATA_DEBUG) -pkg=dashboard -o dashboard/assets.go data/...

GO_FILES := $(shell find . -name "*.go" -print)
lovebeat: dashboard/assets.go $(GO_FILES)
	go build

clean:
	rm -f lovebeat dashboard/assets.go
