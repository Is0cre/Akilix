Name:           akilix
Version:        0.0.1
Release:        0.m0%{?dist}
Summary:        Akilix security engineering workstation foundation
License:        Apache-2.0
URL:            https://github.com/Is0cre/Akilix
BuildRequires:  golang
Source0:        akilix.tar.gz

%description
The minimal Akilix command-line foundation for openSUSE Leap.

%prep
%setup -q -n akilix

%build
go build -trimpath -buildvcs=false -o akilix ./cmd/akilix
go build -trimpath -buildvcs=false -o akilix-udev-handler ./cmd/akilix-udev-handler

%install
install -Dpm0755 akilix %{buildroot}%{_bindir}/akilix
install -Dpm0755 akilix-udev-handler %{buildroot}%{_bindir}/akilix-udev-handler
install -Dpm0644 image/kiwi-iso/root/etc/udev/rules.d/99-akilix-forensic-block.rules %{buildroot}%{_udevrulesdir}/99-akilix-forensic-block.rules
install -Dpm0644 image/kiwi-iso/root/usr/lib/tmpfiles.d/akilix-device-events.conf %{buildroot}%{_tmpfilesdir}/akilix-device-events.conf
install -d %{buildroot}%{_datadir}/akilix/profiles
install -m0644 profiles/*.yaml %{buildroot}%{_datadir}/akilix/profiles/
install -Dpm0644 repositories/repositories.json %{buildroot}%{_datadir}/akilix/repositories.json
install -d %{buildroot}%{_datadir}/zsh/site-functions %{buildroot}%{_datadir}/bash-completion/completions
./akilix completion zsh > %{buildroot}%{_datadir}/zsh/site-functions/_akilix
./akilix completion bash > %{buildroot}%{_datadir}/bash-completion/completions/akilix

%files
%{_bindir}/akilix
%{_bindir}/akilix-udev-handler
%{_udevrulesdir}/99-akilix-forensic-block.rules
%{_tmpfilesdir}/akilix-device-events.conf
%{_datadir}/akilix/profiles/*.yaml
%{_datadir}/akilix/repositories.json
%{_datadir}/zsh/site-functions/_akilix
%{_datadir}/bash-completion/completions/akilix

%changelog
* Tue Aug 04 2026 Akilix Contributors - 0.0.1-0.m0
- Initial M0 foundation package
