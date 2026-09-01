PREFIX ?= /usr
VERSION ?= 0.0.1-m0

.PHONY: build test check branding-check branding-stage boot-diagnostics-check network-image-check operator-freedom-check device-policy-check iso-payload-check weathr-stage ghidra-stage memtest-stage usb-ids-stage completion-check manifest-check install install-completion rpm rpm-check kiwi kiwi-iso kiwi-iso-prompt schema-check vm-check

build:
	go build -trimpath -buildvcs=false -o akilix ./cmd/akilix
	go build -trimpath -buildvcs=false -o akilix-udev-handler ./cmd/akilix-udev-handler

test:
	go test ./cmd/... ./internal/...

check: test
	go vet ./cmd/... ./internal/...
	$(MAKE) schema-check
	$(MAKE) branding-check
	$(MAKE) boot-diagnostics-check
	$(MAKE) network-image-check
	$(MAKE) operator-freedom-check
	$(MAKE) device-policy-check
	$(MAKE) build
	$(MAKE) completion-check
	$(MAKE) manifest-check

completion-check: build
	./akilix completion bash | bash -n
	@if command -v zsh >/dev/null 2>&1; then ./akilix completion zsh | zsh -n; fi

schema-check:
	@for schema in schemas/*.json; do python3 -m json.tool "$$schema" >/dev/null || exit 1; done

manifest-check: build
	sh scripts/check-manifests.sh

branding-check:
	python3 scripts/check-branding.py

boot-diagnostics-check:
	sh scripts/check-boot-diagnostics.sh

network-image-check:
	sh scripts/check-network-image.sh

operator-freedom-check:
	sh scripts/check-operator-freedom.sh

device-policy-check:
	sh scripts/check-device-policy.sh

iso-payload-check:
	: "$${ISO_PATH:?Set ISO_PATH to the generated ISO}"
	sh scripts/check-iso-boot-payload.sh "$$ISO_PATH"

branding-stage: branding-check
	install -Dm0644 branding/os/grub/background-1920x1080.png image/kiwi-iso/root/usr/share/grub2/themes/Akilix/background.png
	install -Dm0644 branding/os/grub/logo.png image/kiwi-iso/root/usr/share/akilix/branding/logo.png
	install -Dm0644 branding/os/plymouth/logo.png image/kiwi-iso/root/usr/share/plymouth/themes/akilix/logo.png
	install -Dm0644 branding/os/wallpaper/akilix-3840x2160.png image/kiwi-iso/root/usr/share/backgrounds/akilix/akilix-3840x2160.png
	install -Dm0644 branding/LICENSE image/kiwi-iso/root/usr/share/licenses/akilix-branding/LICENSE

weathr-stage:
	sh scripts/stage-weathr.sh

ghidra-stage:
	sh scripts/stage-ghidra.sh

memtest-stage:
	sh scripts/stage-memtest86plus.sh

usb-ids-stage:
	sh scripts/stage-usb-ids.sh

vm-check:
	sh scripts/check-m0-platform.sh

install: build
	install -Dm0755 akilix $(DESTDIR)$(PREFIX)/bin/akilix
	install -Dm0755 akilix-udev-handler $(DESTDIR)$(PREFIX)/bin/akilix-udev-handler
	install -d $(DESTDIR)$(PREFIX)/share/akilix/profiles
	install -m0644 profiles/*.yaml $(DESTDIR)$(PREFIX)/share/akilix/profiles/
	install -Dm0644 repositories/repositories.json $(DESTDIR)$(PREFIX)/share/akilix/repositories.json

install-completion: build
	zsh_tmp=$$(mktemp); bash_tmp=$$(mktemp); trap 'rm -f "$$zsh_tmp" "$$bash_tmp"' EXIT; ./akilix completion zsh >"$$zsh_tmp"; ./akilix completion bash >"$$bash_tmp"; install -Dm0644 "$$zsh_tmp" $(DESTDIR)$(PREFIX)/share/zsh/site-functions/_akilix; install -Dm0644 "$$bash_tmp" $(DESTDIR)$(PREFIX)/share/bash-completion/completions/akilix

rpm:
	mkdir -p build/rpm/SOURCES
	tar --exclude=build --exclude=akilix.tar.gz --exclude=.git --transform='s,^\./,akilix/,' -czf build/rpm/SOURCES/akilix.tar.gz .
	rpmbuild -bb --define '_topdir $(CURDIR)/build/rpm' --define '_tmppath /tmp' packaging/rpm/akilix.spec

rpm-check: rpm
	sh scripts/check-rpm.sh

kiwi:
	kiwi-ng system build --description image/kiwi --target-dir build/kiwi


kiwi-iso: build branding-stage weathr-stage ghidra-stage memtest-stage usb-ids-stage
	: "$${AKILIX_LIVE_PASSWORD:?Set AKILIX_LIVE_PASSWORD for the live operator account}"
	@if [ -e build/kiwi-iso ]; then \
		archive="build/kiwi-iso.previous-$$(date -u +%Y%m%dT%H%M%SZ)"; \
		test ! -e "$$archive"; \
		mv build/kiwi-iso "$$archive"; \
		printf '%s\n' "preserved previous ISO build as $$archive"; \
	fi
	install -Dm0755 akilix image/kiwi-iso/root/usr/bin/akilix
	install -Dm0755 akilix-udev-handler image/kiwi-iso/root/usr/bin/akilix-udev-handler
	install -Dm0755 scripts/check-m0-platform.sh image/kiwi-iso/root/usr/bin/akilix-m0-check
	mkdir -p image/kiwi-iso/root/usr/share/zsh/site-functions image/kiwi-iso/root/usr/share/bash-completion/completions
	./akilix completion zsh > image/kiwi-iso/root/usr/share/zsh/site-functions/_akilix
	./akilix completion bash > image/kiwi-iso/root/usr/share/bash-completion/completions/akilix
	build_id=$${AKILIX_BUILD_ID:-$$(date -u +%Y%m%dT%H%M%SZ)}; \
	commit=$${AKILIX_GIT_COMMIT:-$$(git rev-parse --short=12 HEAD)}; \
	sh scripts/stage-build-identity.sh "$$build_id" "$$commit" "$(VERSION)" image/kiwi-iso/root/etc/akilix-build; \
	hash=$$(openssl passwd -6 "$$AKILIX_LIVE_PASSWORD"); \
	python3 scripts/render-live-config.py image/kiwi-iso/config.xml image/kiwi-iso/config.generated.xml "$$hash"; \
	trap 'rm -f image/kiwi-iso/config.generated.xml image/kiwi-iso/root/etc/akilix-build' EXIT; \
	kiwi-ng --kiwi-file=config.generated.xml system build --description image/kiwi-iso --target-dir build/kiwi-iso; \
	iso=build/kiwi-iso/akilix-m0-iso.x86_64-0.0.1.iso; \
	sh scripts/check-iso-boot-payload.sh "$$iso"; \
	release=build/kiwi-iso/akilix-$(VERSION)-$$build_id-$$commit.x86_64.iso; \
	ln "$$iso" "$$release"; \
	(cd build/kiwi-iso && sha256sum "$$(basename "$$release")" > "$$(basename "$$release").sha256"); \
	printf '%s\n' "$$release" > build/kiwi-iso/LATEST-ISO

kiwi-iso-prompt:
	sh scripts/build-live-iso.sh
