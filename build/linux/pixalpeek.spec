Name:           pixalpeek
Version:        0.1.5
Release:        1%{?dist}
Summary:        QR code scanner & generator
License:        MIT
URL:            https://rkriad585.github.io/PixalPeek
Source0:        %{name}-%{version}.tar.gz

%description
PixalPeek is a dual-mode QR code tool with a GUI (Wails v3)
and a full-featured CLI. Supports scanning, generating, batch
processing, camera scan, clipboard decode, and more.

%prep
%setup -q

%build
go build -tags desktop,production -ldflags "-s -w" -o %{name} .

%install
mkdir -p %{buildroot}/usr/local/bin
install -m 755 %{name} %{buildroot}/usr/local/bin/%{name}

mkdir -p %{buildroot}/usr/share/applications
cat > %{buildroot}/usr/share/applications/%{name}.desktop <<EOF
[Desktop Entry]
Name=PixalPeek
Comment=QR code scanner & generator
Exec=%{name}
Icon=%{name}
Type=Application
Categories=Utility;
Terminal=false
EOF

%files
/usr/local/bin/%{name}
/usr/share/applications/%{name}.desktop

%changelog
* Mon Aug 25 2026 rkriad585 <rkriad585@gmail.com> - 0.1.5-1
- Initial RPM release
