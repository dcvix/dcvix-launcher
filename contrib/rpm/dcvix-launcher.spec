%define _prefix /usr
%global debug_package %{nil}

Name: dcvix-launcher
Version: %{version}
Release: %{release}%{?dist}
Summary: DCV remote desktop client launcher

License: MIT
URL: https://github.com/dcvix/dcvix-launcher
Source0: %{name}-v%{version}-linux-amd64.tar.gz
Source1: dcvix-launcher.desktop
Source2: Icon.png

BuildRequires: desktop-file-utils

Requires: desktop-file-utils

%description
dcvix is a session broker and server-pool manager
for Amazon DCV. It provides centralized authentication,
desktop session lifecycle management,
and automatic allocation of DCV servers.
This package provides the launcher, a GUI client that runs on the
user's computer. It authenticates users against the director,
displays available DCV servers, and launches the DCV viewer.

.

%prep
%setup -q -n %{name}-v%{version}-linux-amd64
cp %{SOURCE1} .
cp %{SOURCE2} .

%build
# Binary is pre-built

%install
mkdir -p %{buildroot}%{_bindir}
install -m 755 %{name} %{buildroot}%{_bindir}/%{name}

mkdir -p %{buildroot}%{_datadir}/applications
desktop-file-install --dir=%{buildroot}%{_datadir}/applications \
  --set-key=Exec --set-value=%{_bindir}/%{name} \
  %{name}.desktop

mkdir -p %{buildroot}%{_datadir}/pixmaps
install -m 644 Icon.png %{buildroot}%{_datadir}/pixmaps/%{name}.png

%post
desktop-file-validate %{_datadir}/applications/%{name}.desktop || :

%files
%{_bindir}/%{name}
%{_datadir}/applications/%{name}.desktop
%{_datadir}/pixmaps/%{name}.png
%license LICENSE.md
%doc README.md

%changelog
* Thu Jun 04 2026 Diego Cortassa <diego@cortassa.net> - %{version}-%{release}
- Initial package
