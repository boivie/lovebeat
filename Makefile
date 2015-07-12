.PHONY: all
all: lovebeat

dashboard-assets:
	$(MAKE) -C dashboard

dependencies: dashboard-assets
	go get -t ./...

GO_FILES := $(shell find . -name "*.go" -print)
lovebeat: $(GO_FILES) dependencies dashboard-assets
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

docker-build:
	docker run --rm -v $(shell pwd):/src -v /var/run/docker.sock:/var/run/docker.sock centurylink/golang-builder
	docker tag -f lovebeat:latest boivie/lovebeat:latest

docker-upload:
	docker push boivie/lovebeat:latest
