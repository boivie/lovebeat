.PHONY: all
all: lovebeat

ASSETS := $(shell find dashboard/assets/ -print)
BINDATA_DEBUG ?=
GO_BINDATA := $(if $(GOBIN),$(GOBIN),$(GOPATH)/bin)/go-bindata
DESTDIR ?= /

.PHONY: rebuild-assets
rebuild-assets: $(ASSETS)
	go install github.com/jteeuwen/go-bindata/go-bindata
	$(GO_BINDATA) $(BINDATA_DEBUG) -pkg=dashboard -prefix "dashboard/assets/" -o dashboard/assets.go dashboard/assets/...

GO_FILES := $(shell find . -name "*.go" -print)
lovebeat-dev: rebuild-assets $(GO_FILES)
	go build

lovebeat: $(GO_FILES)
	go build

.PHONY: clean
clean:
	rm -f lovebeat

.PHONY: install
install: lovebeat
	mkdir -p $(DESTDIR)/usr/sbin
	install -m 0755 --strip lovebeat $(DESTDIR)/usr/sbin
	mkdir -p $(DESTDIR)/etc/lovebeat.conf.d
	install -m 0644 lovebeat.cfg $(DESTDIR)/etc/lovebeat.conf.d

.PHONY: deb
deb:
	debuild --preserve-envvar GOPATH -uc -us
