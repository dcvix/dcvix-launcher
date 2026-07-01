MAIN_NAME=dcvix-launcher

DIST_DIR=dist

# Version information
VERSION?=$(shell (git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0-dev") | sed 's/^v//')
RELEASE=1
COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Build variables
BINARY_NAME=$(MAIN_NAME)
GO=$(shell which go)
LDFLAGS="-X github.com/dcvix/$(MAIN_NAME)/internal/version.Version=$(VERSION) \
         -X github.com/dcvix/$(MAIN_NAME)/internal/version.Commit=$(COMMIT) \
         -X github.com/dcvix/$(MAIN_NAME)/internal/version.BuildTime=$(BUILD_TIME)"

# Platform-specific variables
LINUX_AMD64_BINARY=$(MAIN_NAME)
LINUX_AMD64_DIR=$(MAIN_NAME)-v$(VERSION)-linux-amd64
WINDOWS_BINARY=$(BINARY_NAME).exe
WINDOWS_AMD64_DIR=$(MAIN_NAME)-v$(VERSION)-windows-amd64
DARWIN_AMD64_DIR=$(MAIN_NAME)-v$(VERSION)-darwin-amd64

# Build all platforms and packages
.PHONY: build
build: build-linux build-windows-cross build-darwin

# Build all packages
.PHONY: build-packages
build-packages: rpm deb installer

# Build for Linux
.PHONY: build-linux
build-linux: update-toml
	mkdir -p $(DIST_DIR)/$(LINUX_AMD64_DIR)
	GOOS=linux GOARCH=amd64 $(GO) build \
		-trimpath -ldflags $(LDFLAGS) \
		-o $(DIST_DIR)/$(LINUX_AMD64_DIR)/$(LINUX_AMD64_BINARY) ./cmd/$(MAIN_NAME)
	cp README.md LICENSE.md $(DIST_DIR)/$(LINUX_AMD64_DIR)/
	cd $(DIST_DIR) && tar czf $(LINUX_AMD64_DIR).tar.gz $(LINUX_AMD64_DIR)

# Build for Windows
.PHONY: build-windows
build-windows: update-toml windows-resource
	mkdir -p $(DIST_DIR)/$(WINDOWS_AMD64_DIR)
	GOOS=windows GOARCH=amd64 $(GO) build \
		-trimpath -ldflags $(LDFLAGS) \
		-o $(DIST_DIR)/$(WINDOWS_AMD64_DIR)/$(WINDOWS_AMD64_BINARY) ./cmd/$(MAIN_NAME)
	cp README.md LICENSE.md $(DIST_DIR)/$(WINDOWS_AMD64_DIR)/
	cd $(DIST_DIR) && 7z a -bd -r $(WINDOWS_AMD64_DIR).zip $(WINDOWS_AMD64_DIR)

# Build for Windows cross compile
.PHONY: build-windows-cross
build-windows-cross: update-toml windows-resource
	mkdir -p $(DIST_DIR)/$(WINDOWS_AMD64_DIR)	
	CGO_ENABLED=1 GOOS=windows GOARCH=amd64 CC=x86_64-w64-mingw32-gcc $(GO) build \
		-trimpath -ldflags $(LDFLAGS) \
		-o $(DIST_DIR)/$(WINDOWS_AMD64_DIR)/$(WINDOWS_AMD64_BINARY) ./cmd/$(MAIN_NAME)
	cp README.md LICENSE.md $(DIST_DIR)/$(WINDOWS_AMD64_DIR)/
	cd $(DIST_DIR) && 7z a -bd -r $(WINDOWS_AMD64_DIR).zip $(WINDOWS_AMD64_DIR)

# Compile resource file for version info and icon
.PHONY: windows-resource
windows-resource:
	go-winres simply \
		--product-version $(VERSION).0 \
		--file-version $(VERSION).0 \
		--file-description "Graphical interface to easily launch the DCV viewer" \
		--product-name "dcvix Launcher" \
		--copyright "Diego Cortassa" \
		--original-filename "$(WINDOWS_AMD64_BINARY)" \
		--icon Icon.png


# Build for macOS
.PHONY: build-darwin
build-darwin: update-toml
	mkdir -p $(DIST_DIR)/$(DARWIN_AMD64_DIR)
	CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 $(GO) build -trimpath -ldflags $(LDFLAGS) -o $(DIST_DIR)/$(DARWIN_AMD64_DIR)/$(MAIN_NAME) ./cmd/$(MAIN_NAME)
	cp README.md LICENSE.md $(DIST_DIR)/$(DARWIN_AMD64_DIR)/
	cd $(DIST_DIR) && zip -r $(DARWIN_AMD64_DIR).zip $(DARWIN_AMD64_DIR)

# Build RPM package
.PHONY: rpm
rpm: build-linux
	mkdir -p $(HOME)/rpmbuild/SOURCES $(HOME)/rpmbuild/SPECS
	cp $(DIST_DIR)/$(LINUX_AMD64_DIR).tar.gz $(HOME)/rpmbuild/SOURCES/
	sed -e 's/Icon\.png/$(MAIN_NAME).png/' contrib/dcvix-launcher.desktop > $(HOME)/rpmbuild/SOURCES/dcvix-launcher.desktop
	cp Icon.png $(HOME)/rpmbuild/SOURCES/
	rpmbuild -ba \
		--define "_topdir $(HOME)/rpmbuild" \
		--define "version $(VERSION)" \
		--define "release $(RELEASE)" \
		contrib/rpm/dcvix-launcher.spec
	cp $(HOME)/rpmbuild/RPMS/x86_64/$(MAIN_NAME)-$(VERSION)-$(RELEASE)*.rpm $(DIST_DIR)/
	cp $(HOME)/rpmbuild/SRPMS/$(MAIN_NAME)-$(VERSION)-$(RELEASE)*.src.rpm $(DIST_DIR)/

.PHONY: rpm-clean
rpm-clean:
	rm -rf $(HOME)/rpmbuild

# Build DEB package
.PHONY: deb
deb: build-linux
	mkdir -p $(DIST_DIR)/deb/$(MAIN_NAME)_$(VERSION)-$(RELEASE)_amd64/DEBIAN
	mkdir -p $(DIST_DIR)/deb/$(MAIN_NAME)_$(VERSION)-$(RELEASE)_amd64/usr/bin
	mkdir -p $(DIST_DIR)/deb/$(MAIN_NAME)_$(VERSION)-$(RELEASE)_amd64/usr/share/applications
	mkdir -p $(DIST_DIR)/deb/$(MAIN_NAME)_$(VERSION)-$(RELEASE)_amd64/usr/share/pixmaps
	sed -e "s/@VERSION@/$(VERSION)/g" -e "s/@RELEASE@/$(RELEASE)/g" contrib/deb/control > $(DIST_DIR)/deb/$(MAIN_NAME)_$(VERSION)-$(RELEASE)_amd64/DEBIAN/control
	cp contrib/deb/copyright $(DIST_DIR)/deb/$(MAIN_NAME)_$(VERSION)-$(RELEASE)_amd64/DEBIAN/
	cp $(DIST_DIR)/$(LINUX_AMD64_DIR)/$(MAIN_NAME) $(DIST_DIR)/deb/$(MAIN_NAME)_$(VERSION)-$(RELEASE)_amd64/usr/bin/
	cp contrib/dcvix-launcher.desktop $(DIST_DIR)/deb/$(MAIN_NAME)_$(VERSION)-$(RELEASE)_amd64/usr/share/applications/
	cp Icon.png $(DIST_DIR)/deb/$(MAIN_NAME)_$(VERSION)-$(RELEASE)_amd64/usr/share/pixmaps/$(MAIN_NAME).png
	dpkg-deb --build $(DIST_DIR)/deb/$(MAIN_NAME)_$(VERSION)-$(RELEASE)_amd64
	mv $(DIST_DIR)/deb/$(MAIN_NAME)_$(VERSION)-$(RELEASE)_amd64.deb $(DIST_DIR)/$(MAIN_NAME)_$(VERSION)-$(RELEASE)_amd64.deb

.PHONY: deb-clean
deb-clean:
	rm -rf $(DIST_DIR)/deb

# Build NSIS Windows installer
.PHONY: installer
installer: build-windows-cross
	makensis -DVERSION=$(VERSION) -DNAME=$(MAIN_NAME) -DSRCDIR="$(CURDIR)" contrib/nsis/installer.nsi

.PHONY: installer-clean
installer-clean:
	rm -f $(DIST_DIR)/$(MAIN_NAME)-v$(VERSION)-windows-amd64-setup.exe

## audit: run quality control checks
.PHONY: audit
audit:
	$(GO) mod tidy -diff
	$(GO) mod verify
	test -z "$(shell gofmt -l .)"
	$(GO) vet ./...
	$(GO) run honnef.co/go/tools/cmd/staticcheck@latest -checks=all ./...
	$(GO) run golang.org/x/vuln/cmd/govulncheck@latest ./...

# Show version
.PHONY: version
version:
	@echo $(VERSION)

# Create a new version tag
.PHONY: tag
tag: update-toml
	git add FyneApp.toml
	git commit -m "chore: update version to $(VERSION)" || true
	git tag -a v$(VERSION) -m "Version $(VERSION)"
	@echo "Tagged v$(VERSION). Push with: git push origin v$(VERSION)"

# update-toml
.PHONY: update-toml
update-toml:
	sed -e "s/^  Version = \".*\"/  Version = \"$(VERSION)\"/" \
	    -e "s/^  Build = .*/  Build = $(RELEASE)/" FyneApp.toml > FyneApp.toml.tmp
	mv FyneApp.toml.tmp FyneApp.toml

.PHONY: clean
clean:
	rm -rf $(DIST_DIR)
	rm -rf fyne-cross
	rm -f *.syso

.PHONY: run-debug
run-debug:
	go run -tags debug cmd/$(MAIN_NAME)/main.go ;

.PHONY: run
run:
	go run cmd/$(MAIN_NAME)/main.go ;
