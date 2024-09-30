build/linux/arm64:
	$(info Building debug tool for linux arm64)
	wails build --platform linux/arm64 -o ftb-debug-linux-arm64 -trimpath -ldflags "-s -w -X 'main.GitCommit=$(GITHUB_SHA_SHORT)' -X 'main.Version=$(GITHUB_REF_NAME)'"
	#GOOS=linux GOARCH=arm64 go build -o out/ftb-debug-linux-arm64 -trimpath -buildvcs=false -ldflags "-s -w -X 'main.GitCommit=$(GITHUB_SHA_SHORT)' -X 'main.Version=$(GITHUB_REF_NAME)'"

build/linux/amd64:
	$(info Building debug tool for linux amd64)
	wails build --platform linux/amd64 -o ftb-debug-linux-amd64 -trimpath -ldflags "-s -w -X 'main.GitCommit=$(GITHUB_SHA_SHORT)' -X 'main.Version=$(GITHUB_REF_NAME)'"
	#GOOS=linux GOARCH=amd64 go build -o out/ftb-debug-linux-amd64 -trimpath -buildvcs=false -ldflags "-s -w -X 'main.GitCommit=$(GITHUB_SHA_SHORT)' -X 'main.Version=$(GITHUB_REF_NAME)'"

build/windows/arm64:
	$(info Building debug tool for windows arm64)
	wails build --platform windows/arm64 -o ftb-debug-windows-arm64 -trimpath -ldflags "-s -w -X 'main.GitCommit=$(GITHUB_SHA_SHORT)' -X 'main.Version=$(GITHUB_REF_NAME)'"
	#GOOS=windows GOARCH=arm64 go build -o out/ftb-debug-windows-arm64.exe -trimpath -buildvcs=false -ldflags "-s -w -X 'main.GitCommit=$(GITHUB_SHA_SHORT)' -X 'main.Version=$(GITHUB_REF_NAME)'"

build/windows/amd64:
	$(info Building debug tool for windows amd64)
	wails build --platform windows/amd64 -o ftb-debug-windows-amd64 -trimpath -ldflags "-s -w -X 'main.GitCommit=$(GITHUB_SHA_SHORT)' -X 'main.Version=$(GITHUB_REF_NAME)'"
	#GOOS=windows GOARCH=amd64 go build -o out/ftb-debug-windows-amd64.exe -trimpath -buildvcs=false -ldflags "-s -w -X 'main.GitCommit=$(GITHUB_SHA_SHORT)' -X 'main.Version=$(GITHUB_REF_NAME)'"

build/darwin/amd64:
	$(info Building debug tool for darwin amd64)
	wails build --platform darwin/arm64 -o ftb-debug-darwin-arm64 -trimpath -ldflags "-s -w -X 'main.GitCommit=$(GITHUB_SHA_SHORT)' -X 'main.Version=$(GITHUB_REF_NAME)'"
	#GOOS=darwin GOARCH=amd64 go build -o out/ftb-debug-darwin-amd64 -trimpath -buildvcs=false -ldflags "-s -w -X 'main.GitCommit=$(GITHUB_SHA_SHORT)' -X 'main.Version=$(GITHUB_REF_NAME)'"

build/darwin/arm64:
	$(info Building debug tool for darwin arm64)
	wails build --platform darwin/amd64 -o ftb-debug-darwin-amd64 -trimpath -ldflags "-s -w -X 'main.GitCommit=$(GITHUB_SHA_SHORT)' -X 'main.Version=$(GITHUB_REF_NAME)'"
	#GOOS=darwin GOARCH=arm64 go build -o out/ftb-debug-darwin-arm64 -trimpath -buildvcs=false -ldflags "-s -w -X 'main.GitCommit=$(GITHUB_SHA_SHORT)' -X 'main.Version=$(GITHUB_REF_NAME)'"

build_all: build/linux/arm64 build/linux/amd64 build/windows/arm64 build/windows/amd64 build/darwin/amd64 build/darwin/arm64