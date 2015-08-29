.PHONY: all
all: lovebeat

dashboard-assets:
	$(MAKE) -C dashboard

dependencies: dashboard-assets
	go get -t ./...

# Explicitly listing version.go since it might not exist when
# the makefile is parsed.
GO_FILES := $(shell find . -name "*.go" -print) version.go
lovebeat: $(GO_FILES) dependencies dashboard-assets
	go build

# Generate version.go based on the "git describe" output so that
# Lovebeat's reported version number is always descriptive and useful.
# This rule must always run but the target file is only updated if
# there's an actual change in the version number.
.FORCE:
.PRECIOUS: version.go
version.go: .FORCE
	TMPFILE=$$(mktemp $@.XXXX) && \
	    echo "package main" >> $$TMPFILE && \
	    echo "const VERSION = \"$$(git describe --tags --always)\"" \
	            >> $$TMPFILE && \
	    gofmt -w $$TMPFILE && \
	    if ! cmp --quiet $$TMPFILE $@ ; then \
	        mv $$TMPFILE $@ ; \
	    fi && \
	    rm -f $$TMPFILE

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

docker-build:
	docker run --rm -v $(shell pwd):/src -v /var/run/docker.sock:/var/run/docker.sock centurylink/golang-builder
	docker tag -f lovebeat:latest boivie/lovebeat:latest

docker-upload:
	docker push boivie/lovebeat:latest
