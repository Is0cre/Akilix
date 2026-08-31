PREFIX ?= /usr
VERSION ?= 0.0.1-m0

.PHONY: build test check branding-check branding-stage completion-check manifest-check install install-completion rpm rpm-check kiwi kiwi-iso kiwi-iso-prompt schema-check vm-check

build:
	go build -trimpath -buildvcs=false -o pensuse ./cmd/pensuse

test:
	go test ./cmd/... ./internal/...

check: test
	go vet ./cmd/... ./internal/...
	$(MAKE) schema-check
	$(MAKE) branding-check
	$(MAKE) build
	$(MAKE) completion-check
	$(MAKE) manifest-check

completion-check: build
	./pensuse completion bash | bash -n
	@if command -v zsh >/dev/null 2>&1; then ./pensuse completion zsh | zsh -n; fi

schema-check:
	@for schema in schemas/*.json; do python3 -m json.tool "$$schema" >/dev/null || exit 1; done

manifest-check: build
	sh scripts/check-manifests.sh

branding-check:
	python3 scripts/check-branding.py

branding-stage: branding-check
	install -Dm0644 branding/os/grub/background-1920x1080.png image/kiwi-iso/root/usr/share/grub2/themes/PenSUSE/background.png
	install -Dm0644 branding/os/grub/logo.png image/kiwi-iso/root/usr/share/pensuse/branding/logo.png
	install -Dm0644 branding/os/plymouth/logo.png image/kiwi-iso/root/usr/share/plymouth/themes/pensuse/logo.png
	install -Dm0644 branding/os/wallpaper/pensuse-3840x2160.png image/kiwi-iso/root/usr/share/backgrounds/pensuse/pensuse-3840x2160.png
	install -Dm0644 branding/LICENSE image/kiwi-iso/root/usr/share/licenses/pensuse-branding/LICENSE

vm-check:
	sh scripts/check-m0-platform.sh

install: build
	install -Dm0755 pensuse $(DESTDIR)$(PREFIX)/bin/pensuse
	install -d $(DESTDIR)$(PREFIX)/share/pensuse/profiles
	install -m0644 profiles/*.yaml $(DESTDIR)$(PREFIX)/share/pensuse/profiles/
	install -Dm0644 repositories/repositories.json $(DESTDIR)$(PREFIX)/share/pensuse/repositories.json

install-completion: build
	zsh_tmp=$$(mktemp); bash_tmp=$$(mktemp); trap 'rm -f "$$zsh_tmp" "$$bash_tmp"' EXIT; ./pensuse completion zsh >"$$zsh_tmp"; ./pensuse completion bash >"$$bash_tmp"; install -Dm0644 "$$zsh_tmp" $(DESTDIR)$(PREFIX)/share/zsh/site-functions/_pensuse; install -Dm0644 "$$bash_tmp" $(DESTDIR)$(PREFIX)/share/bash-completion/completions/pensuse

rpm:
	mkdir -p build/rpm/SOURCES
	tar --exclude=build --exclude=pensuse.tar.gz --exclude=.git --transform='s,^\./,pensuse/,' -czf build/rpm/SOURCES/pensuse.tar.gz .
	rpmbuild -bb --define '_topdir $(CURDIR)/build/rpm' --define '_tmppath /tmp' packaging/rpm/pensuse.spec

rpm-check: rpm
	sh scripts/check-rpm.sh

kiwi:
	kiwi-ng system build --description image/kiwi --target-dir build/kiwi


kiwi-iso: build branding-stage
	: "$${PENSUSE_LIVE_PASSWORD:?Set PENSUSE_LIVE_PASSWORD for the live operator account}"
	install -Dm0755 pensuse image/kiwi-iso/root/usr/bin/pensuse
	install -Dm0755 scripts/check-m0-platform.sh image/kiwi-iso/root/usr/bin/pensuse-m0-check
	mkdir -p image/kiwi-iso/root/usr/share/zsh/site-functions image/kiwi-iso/root/usr/share/bash-completion/completions
	./pensuse completion zsh > image/kiwi-iso/root/usr/share/zsh/site-functions/_pensuse
	./pensuse completion bash > image/kiwi-iso/root/usr/share/bash-completion/completions/pensuse
	hash=$$(openssl passwd -6 "$$PENSUSE_LIVE_PASSWORD"); \
	python3 scripts/render-live-config.py image/kiwi-iso/config.xml image/kiwi-iso/config.generated.xml "$$hash"; \
	trap 'rm -f image/kiwi-iso/config.generated.xml' EXIT; \
	kiwi-ng --kiwi-file=config.generated.xml system build --description image/kiwi-iso --target-dir build/kiwi-iso

kiwi-iso-prompt:
	sh scripts/build-live-iso.sh
