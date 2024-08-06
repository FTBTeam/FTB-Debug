build/linux/arm64:
	$(info Building debug tool for linux arm64)
	GOOS=linux GOARCH=arm64 go build -o out/ftb-debug-linux-arm64 -trimpath -buildvcs=false -ldflags "-s -w -X 'main.GitCommit=$(GITHUB_SHA_SHORT)' -X 'main.Version=$(GITHUB_REF_NAME)'"

build/linux/amd64:
	$(info Building debug tool for linux amd64)
	GOOS=linux GOARCH=amd64 go build -o out/ftb-debug-linux-amd64 -trimpath -buildvcs=false -ldflags "-s -w -X 'main.GitCommit=$(GITHUB_SHA_SHORT)' -X 'main.Version=$(GITHUB_REF_NAME)'"

build/windows/arm64:
	$(info Building debug tool for windows arm64)
	GOOS=windows GOARCH=arm64 go build -o out/ftb-debug-windows-arm64.exe -trimpath -buildvcs=false -ldflags "-s -w -X 'main.GitCommit=$(GITHUB_SHA_SHORT)' -X 'main.Version=$(GITHUB_REF_NAME)'"

build/windows/amd64:
	$(info Building debug tool for windows amd64)
	GOOS=windows GOARCH=amd64 go build -o out/ftb-debug-windows-amd64.exe -trimpath -buildvcs=false -ldflags "-s -w -X 'main.GitCommit=$(GITHUB_SHA_SHORT)' -X 'main.Version=$(GITHUB_REF_NAME)'"

build/darwin/amd64:
	$(info Building debug tool for darwin amd64)
	GOOS=darwin GOARCH=amd64 go build -o out/ftb-debug-darwin-amd64 -trimpath -buildvcs=false -ldflags "-s -w -X 'main.GitCommit=$(GITHUB_SHA_SHORT)' -X 'main.Version=$(GITHUB_REF_NAME)'"

build/darwin/arm64:
	$(info Building debug tool for darwin arm64)
	GOOS=darwin GOARCH=arm64 go build -o out/ftb-debug-darwin-arm64 -trimpath -buildvcs=false -ldflags "-s -w -X 'main.GitCommit=$(GITHUB_SHA_SHORT)' -X 'main.Version=$(GITHUB_REF_NAME)'"

build_all: build/linux/arm build/linux/arm64 build/linux/amd64 build/windows/arm64 build/windows/amd64 build/darwin/amd64 build/darwin/arm64