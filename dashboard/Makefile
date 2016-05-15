GO_BINDATA := $(if $(GOBIN),$(GOBIN),$(GOPATH)/bin)/go-bindata
GULP := node_modules/.bin/gulp
BOWER := node_modules/.bin/bower

DESTDIR ?= /f
ASSETS := $(shell find assets/ -print)
JS_FILES := $(shell find js/ -print)
BINDATA_DEBUG ?=

assets.go: $(ASSETS) assets/bundle.js
	go get -u github.com/jteeuwen/go-bindata/...
	$(GO_BINDATA) $(BINDATA_DEBUG) -pkg=dashboard -prefix "assets/" -o assets.go assets/...

assets/bundle.js: node_modules/webpack/bin/webpack.js $(JS_FILES)
	./node_modules/webpack/bin/webpack.js

node_modules/webpack/bin/webpack.js:
	npm install
