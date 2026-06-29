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
dcvix is an orchestrator for the Amazon DCV remote desktop system focused on
simplicity, lightweight operation, and security. It enables small to medium-sized
organizations to manage the creation of sessions and server pools.

This package provides the launcher GUI that allows users to authenticate,
select a DCV server, and launch the DCV viewer.

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
* Thu Jun 04 2026 Diego Cortassa <diego@cortassa.net> - 0.1.0-1
- Initial package
