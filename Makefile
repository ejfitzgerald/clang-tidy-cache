default:
	go build github.com/ejfitzgerald/clang-tidy-cache

release:
	GOOS=darwin GOARCH=amd64 go build -o clang-tidy-cache-darwin-amd64 github.com/ejfitzgerald/clang-tidy-cache
	GOOS=linux GOARCH=amd64 go build -o clang-tidy-cache-linux-amd64 github.com/ejfitzgerald/clang-tidy-cache
	GOOS=windows GOARCH=amd64 go build -o clang-tidy-cache-windows-amd64.exe github.com/ejfitzgerald/clang-tidy-cache

.PHONY: default release