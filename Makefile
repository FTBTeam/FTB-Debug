#build/linux/arm64:
#	$(info Building debug tool for linux arm64)
#	wails build -clean -platform linux/arm64 -trimpath -ldflags "-s -w -X 'main.GitCommit=$(GITHUB_SHA_SHORT)' -X 'main.Version=$(GITHUB_REF_NAME)'"
#	mkdir -p ./out
#	mv ./build/bin/ftb-debug-ui ./out/ftb-debug-linux-arm64
#
#build/linux/amd64:
#	$(info Building debug tool for linux amd64)
#	wails build -clean -platform linux/amd64 -trimpath -ldflags "-s -w -X 'main.GitCommit=$(GITHUB_SHA_SHORT)' -X 'main.Version=$(GITHUB_REF_NAME)'"
#	mkdir -p ./out
#	mv ./build/bin/ftb-debug-ui ./out/ftb-debug-windows-amd64

build/windows/arm64:
	$(info Building debug tool for windows arm64)
	wails build -clean -platform windows/arm64 -trimpath -ldflags "-s -w -X 'main.GitCommit=$(GITHUB_SHA_SHORT)' -X 'main.Version=$(GITHUB_REF_NAME)'"
	mkdir -p ./out
	mv ./build/bin/ftb-debug-ui.exe ./out/ftb-debug-windows-arm64.exe

build/windows/amd64:
	$(info Building debug tool for windows amd64)
	wails build -clean -platform windows/amd64 -trimpath -ldflags "-s -w -X 'main.GitCommit=$(GITHUB_SHA_SHORT)' -X 'main.Version=$(GITHUB_REF_NAME)'"
	mkdir -p ./out
	mv ./build/bin/ftb-debug-ui.exe ./out/ftb-debug-windows-amd64.exe

build/darwin/amd64:
	$(info Building debug tool for darwin amd64)
	wails build -clean -platform darwin/amd64 -trimpath -ldflags "-s -w -X 'main.GitCommit=$(GITHUB_SHA_SHORT)' -X 'main.Version=$(GITHUB_REF_NAME)'"
	mkdir -p ./out
	cd ./build/bin; zip -r ./out/ftb-debug-darwin-amd64.zip ./build/bin/ftb-debug-ui.app

build/darwin/arm64:
	$(info Building debug tool for darwin arm64)
	wails build -clean -platform darwin/arm64 -trimpath -ldflags "-s -w -X 'main.GitCommit=$(GITHUB_SHA_SHORT)' -X 'main.Version=$(GITHUB_REF_NAME)'"
	mkdir -p ./out
	cd ./build/bin; zip -r ./out/ftb-debug-darwin-arm64.zip ./build/bin/ftb-debug-ui.app
build_all: build/linux/arm64 build/linux/amd64 build/windows/arm64 build/windows/amd64 build/darwin/amd64 build/darwin/arm64