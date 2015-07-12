GO_BINDATA := $(if $(GOBIN),$(GOBIN),$(GOPATH)/bin)/go-bindata
GULP := node_modules/.bin/gulp
BOWER := node_modules/.bin/bower

DESTDIR ?= /f
ASSETS := $(shell find assets/ -print)
BINDATA_DEBUG ?=

assets.go: $(ASSETS) assets/lovebeat.js assets/lovebeat.css
	go install github.com/jteeuwen/go-bindata/go-bindata
	$(GO_BINDATA) $(BINDATA_DEBUG) -pkg=dashboard -prefix "assets/" -o assets.go assets/...

assets/lovebeat.js: js/app.js js/controllers.js js/services.js bower_components/ | $(GULP)
	$(GULP) scripts

assets/lovebeat.css: css/application.scss bower_components/ | $(GULP)
	$(GULP) css

$(GULP):
	npm install

$(BOWER):
	npm install

bower_components/: $(BOWER)
	$(BOWER) install
