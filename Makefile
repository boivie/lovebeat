.PHONY: all
all: lovebeat

ASSETS := $(shell find data -print)
BINDATA_DEBUG ?=
GO_BINDATA := $(if $(GOBIN),$(GOBIN),$(GOPATH)/bin)/go-bindata

dashboard/assets.go: $(ASSETS)
	go install github.com/jteeuwen/go-bindata/go-bindata
	$(GO_BINDATA) $(BINDATA_DEBUG) -pkg=dashboard -o dashboard/assets.go data/...

GO_FILES := $(shell find . -name "*.go" -print)
lovebeat: dashboard/assets.go $(GO_FILES)
	go build

.PHONY: clean
clean:
	rm -f lovebeat dashboard/assets.go
