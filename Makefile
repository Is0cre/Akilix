PREFIX ?= /usr
VERSION ?= 0.0.1-m0

.PHONY: build test check completion-check install install-completion rpm kiwi kiwi-iso schema-check vm-check

build:
	go build -trimpath -buildvcs=false -o pensuse ./cmd/pensuse

test:
	go test ./cmd/... ./internal/...

check: test
	go vet ./cmd/... ./internal/...
	$(MAKE) schema-check
	$(MAKE) build
	$(MAKE) completion-check

completion-check: build
	./pensuse completion bash | bash -n
	@if command -v zsh >/dev/null 2>&1; then ./pensuse completion zsh | zsh -n; fi

schema-check:
	@for schema in schemas/*.json; do python3 -m json.tool "$$schema" >/dev/null || exit 1; done

vm-check:
	sh scripts/check-m0-platform.sh

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


kiwi-iso: build
	: "$${PENSUSE_LIVE_PASSWORD:?Set PENSUSE_LIVE_PASSWORD for the live operator account}"
	install -Dm0755 pensuse image/kiwi-iso/root/usr/bin/pensuse
	install -Dm0755 scripts/check-m0-platform.sh image/kiwi-iso/root/usr/bin/pensuse-m0-check
	mkdir -p image/kiwi-iso/root/usr/share/zsh/site-functions image/kiwi-iso/root/usr/share/bash-completion/completions
	./pensuse completion zsh > image/kiwi-iso/root/usr/share/zsh/site-functions/_pensuse
	./pensuse completion bash > image/kiwi-iso/root/usr/share/bash-completion/completions/pensuse
	hash=$$(openssl passwd -6 "$$PENSUSE_LIVE_PASSWORD"); \
	python3 scripts/render-live-config.py image/kiwi-iso/config.xml image/kiwi-iso/config.generated.xml "$$hash"; \
	trap 'rm -f image/kiwi-iso/config.generated.xml' EXIT; \
	kiwi-ng system build --description image/kiwi-iso --kiwi-file=config.generated.xml --target-dir build/kiwi-iso
