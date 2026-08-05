Name:           pensuse
Version:        0.0.1
Release:        0.m0%{?dist}
Summary:        PenSUSE security engineering workstation foundation
License:        Apache-2.0
URL:            https://github.com/pensuse/pensuse
BuildRequires:  golang
Source0:        pensuse.tar.gz

%description
The minimal PenSUSE command-line foundation for openSUSE Leap.

%prep
%setup -q -n pensuse

%build
go build -trimpath -buildvcs=false -o pensuse ./cmd/pensuse

%install
install -Dpm0755 pensuse %{buildroot}%{_bindir}/pensuse
install -d %{buildroot}%{_datadir}/pensuse/profiles
install -m0644 profiles/*.yaml %{buildroot}%{_datadir}/pensuse/profiles/
install -d %{buildroot}%{_datadir}/zsh/site-functions %{buildroot}%{_datadir}/bash-completion/completions
./pensuse completion zsh > %{buildroot}%{_datadir}/zsh/site-functions/_pensuse
./pensuse completion bash > %{buildroot}%{_datadir}/bash-completion/completions/pensuse

%files
%{_bindir}/pensuse
%{_datadir}/pensuse/profiles/*.yaml
%{_datadir}/zsh/site-functions/_pensuse
%{_datadir}/bash-completion/completions/pensuse

%changelog
* Tue Aug 04 2026 PenSUSE Contributors - 0.0.1-0.m0
- Initial M0 foundation package
