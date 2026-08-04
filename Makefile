PREFIX ?= /usr
VERSION ?= 0.0.1-m0

.PHONY: build test install rpm kiwi schema-check

build:
	go build -trimpath -buildvcs=false -o pensuse ./cmd/pensuse

test:
	go test ./...

schema-check:
	@for schema in schemas/*.json; do python3 -m json.tool "$$schema" >/dev/null || exit 1; done

install: build
	install -Dm0755 pensuse $(DESTDIR)$(PREFIX)/bin/pensuse

rpm:
	mkdir -p build/rpm/SOURCES
	tar --exclude=build --exclude=pensuse.tar.gz --exclude=.git --transform='s,^\./,pensuse/,' -czf build/rpm/SOURCES/pensuse.tar.gz .
	rpmbuild -bb --define '_topdir $(CURDIR)/build/rpm' --define '_tmppath /tmp' packaging/rpm/pensuse.spec

kiwi:
	kiwi-ng system build --description image/kiwi --target-dir build/kiwi
