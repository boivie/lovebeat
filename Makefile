all: lovebeat-go

ASSETS := $(shell find data -print)

dashboard/assets.go: $(ASSETS)
	go-bindata -pkg=dashboard -o dashboard/assets.go data/...

GO_FILES := $(shell find . -name "*.go" -print)
lovebeat-go: dashboard/assets.go $(GO_FILES)
	go build

clean:
	rm -f lovebeat-go dashboard/assets.go
