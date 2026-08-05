PREFIX ?= /usr
VERSION ?= 0.0.1-m0

.PHONY: build test check install install-completion rpm kiwi schema-check

build:
	go build -trimpath -buildvcs=false -o pensuse ./cmd/pensuse

test:
	go test ./...

check: test
	go vet ./...
	$(MAKE) schema-check
	$(MAKE) build

schema-check:
	@for schema in schemas/*.json; do python3 -m json.tool "$$schema" >/dev/null || exit 1; done

install: build
	install -Dm0755 pensuse $(DESTDIR)$(PREFIX)/bin/pensuse

install-completion: build
	zsh_tmp=$$(mktemp); bash_tmp=$$(mktemp); trap 'rm -f "$$zsh_tmp" "$$bash_tmp"' EXIT; ./pensuse completion zsh >"$$zsh_tmp"; ./pensuse completion bash >"$$bash_tmp"; install -Dm0644 "$$zsh_tmp" $(DESTDIR)$(PREFIX)/share/zsh/site-functions/_pensuse; install -Dm0644 "$$bash_tmp" $(DESTDIR)$(PREFIX)/share/bash-completion/completions/pensuse

rpm:
	mkdir -p build/rpm/SOURCES
	tar --exclude=build --exclude=pensuse.tar.gz --exclude=.git --transform='s,^\./,pensuse/,' -czf build/rpm/SOURCES/pensuse.tar.gz .
	rpmbuild -bb --define '_topdir $(CURDIR)/build/rpm' --define '_tmppath /tmp' packaging/rpm/pensuse.spec

kiwi:
	kiwi-ng system build --description image/kiwi --target-dir build/kiwi
