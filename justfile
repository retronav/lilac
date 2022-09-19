GIT_VERSION := `git rev-parse --verify HEAD`
check:
    go fmt ./...
    go test -v ./...
dev:
    go run .
build:
    go build -ldflags="-X 'karawale.in/go/lilac/version.Version=git:{{ GIT_VERSION }}'" -o build/lilacd .
