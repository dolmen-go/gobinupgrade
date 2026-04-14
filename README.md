# gobinupgrade

`gobinupgrade` is a tool to upgrade Go binaries installed in `$GOPATH/bin` (or `$GOBIN`) to their latest versions.

It works by reading the embedded build information from the binary to determine its original package path, build tags, and environment variables (like `GOOS`, `GOARCH`, `CGO_ENABLED`), and then runs `go install <path>@latest` with those same settings.

## Installation

```bash
go install github.com/dolmen-go/gobinupgrade@latest
```

## Usage

```bash
gobinupgrade [-n] [-v] <binary_name>[@<version>...
```

Default version is 'latest'. To keep the same version as already installed use just '<binary_name>@'.

### Options

- `-n`: Dry run. Show what would be done without actually performing the upgrade.
- `-v`: Verbose output. Show detailed build information.

### Examples

Show information about all $GOPATH/bin binaries:
```bash
gobinupgrade -n $(go env GOPATH)/bin/*
```

Upgrade `gobinupgrade` itself:
```bash
gobinupgrade gobinupgrade
```

Upgrade multiple tools:
```bash
gobinupgrade staticcheck golangci-lint
```

## How it works

1.  Locates the binary in `$GOBIN` or `$GOPATH/bin`.
2.  Uses [`debug/buildinfo`](https://pkg.go.dev/debug/buildinfo) (available since Go 1.18) to extract:
    *   The main package path.
    *   The module path.
    *   Build tags (`-tags`).
    *   Build environment variables (`CGO_ENABLED`, etc.).
3.  Executes `go install -buildvcs=true <package>@latest` with the recovered tags and environment variables.

## Trivia

`gobinupgrade` was created years ago (2018?) in the GOPATH era, as a shell script called `go-bin-update`.
I ([dolmen](https://github.com/dolmen)) used to run it after each upgrade of Go to rebuild with the new compiler.

## License

[Apache License, Version 2.0](LICENSE)
