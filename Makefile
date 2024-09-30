build/linux/arm64:
	$(info Building debug tool for linux arm64)
	CGO_ENABLED=1 CC=aarch64-linux-gnu-gcc GOOS=linux GOARCH=arm64 wails build -o ftb-debug-linux-arm64 -trimpath -ldflags "-s -w -X 'main.GitCommit=$(GITHUB_SHA_SHORT)' -X 'main.Version=$(GITHUB_REF_NAME)'"

build/linux/amd64:
	$(info Building debug tool for linux amd64)
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 wails build -o ftb-debug-linux-amd64 -trimpath -ldflags "-s -w -X 'main.GitCommit=$(GITHUB_SHA_SHORT)' -X 'main.Version=$(GITHUB_REF_NAME)'"

build/windows/arm64:
	$(info Building debug tool for windows arm64)
	CGO_ENABLED=1 CC=aarch64-linux-gnu-gcc GOOS=windows GOARCH=arm64 wails build -o ftb-debug-windows-arm64 -trimpath -ldflags "-s -w -X 'main.GitCommit=$(GITHUB_SHA_SHORT)' -X 'main.Version=$(GITHUB_REF_NAME)'"

build/windows/amd64:
	$(info Building debug tool for windows amd64)
	CGO_ENABLED=1 GOOS=windows GOARCH=amd64 wails build -o ftb-debug-windows-amd64 -trimpath -ldflags "-s -w -X 'main.GitCommit=$(GITHUB_SHA_SHORT)' -X 'main.Version=$(GITHUB_REF_NAME)'"

build/darwin/amd64:
	$(info Building debug tool for darwin amd64)
	CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 wails build -o ftb-debug-darwin-arm64 -trimpath -ldflags "-s -w -X 'main.GitCommit=$(GITHUB_SHA_SHORT)' -X 'main.Version=$(GITHUB_REF_NAME)'"

build/darwin/arm64:
	$(info Building debug tool for darwin arm64)
	CGO_ENABLED=1 CC=aarch64-linux-gnu-gcc GOOS=darwin GOARCH=arm64 wails build -o ftb-debug-darwin-amd64 -trimpath -ldflags "-s -w -X 'main.GitCommit=$(GITHUB_SHA_SHORT)' -X 'main.Version=$(GITHUB_REF_NAME)'"

build_all: build/linux/arm64 build/linux/amd64 build/windows/arm64 build/windows/amd64 build/darwin/amd64 build/darwin/arm64