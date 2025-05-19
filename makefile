bash:
	export CGO_ENABLED=1
	export CC=x86_64-w64-mingw32-gcc
	GOOS=windows GOARCH=amd64 go build   -o ./dist/screensage.exe   -ldflags="-s -w"   ./cmd/screensage