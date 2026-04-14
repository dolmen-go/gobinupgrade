// gobinupgrade upgrades Go binaries installed in $GOPATH/bin.
//
// Usage: gobinupgrade [-n] [-v] <file>[@<version>]...
//
// Options:
//
//	-n  don't perform update (dry run)
//	-v  verbose output
//
// Example:
//
//   - Upgrade gobinupgrade itself:
//     gobinupgrade gobinupgrade
package main

import (
	"bufio"
	"debug/buildinfo"
	"flag"
	"fmt"
	"go/build"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	_ "embed"
)

//go:embed main.go
var source string

var (
	noUpdate = flag.Bool("n", false, "don't perform update (dry run)")
	verbose  = flag.Bool("v", false, "verbose output")
)

func main() {
	flag.Usage = func() {
		// Extract package documentation from source
		scanner := bufio.NewScanner(strings.NewReader(source))
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "package ") {
				break
			}
			line = strings.TrimPrefix(line, "// ")
			line = strings.TrimPrefix(line, "//")
			fmt.Fprintln(os.Stderr, line)
		}
	}
	flag.Parse()

	if flag.NArg() == 0 {
		flag.Usage()
		os.Exit(1)
	}

	gobin := os.Getenv("GOBIN")
	if gobin == "" {
		gobin = filepath.Join(filepath.SplitList(build.Default.GOPATH)[0], "bin")
	}

	// Each arg is the name of a tool in $GOPATH/bin
	for _, f := range flag.Args() {
		var version string
		var ok bool
		if f, version, ok = strings.Cut(f, "@"); !ok {
			version = "latest"
		}
		if runtime.GOOS == "windows" {
			if filepath.Ext(f) == "" {
				f += ".exe"
			}
		}
		if dir, _ := filepath.Split(f); dir == "" {
			f = filepath.Join(gobin, f)
		}
		process(f, version)
	}
}

func process(binPath string, version string) {
	info, err := buildinfo.ReadFile(binPath)
	if err != nil || info.Path == "" {
		fmt.Fprintf(os.Stderr, "%s: no go module information embeded in binary\n", binPath)
		return
	}

	fmt.Printf("%s: path=%s module=%s@%s\n", filepath.Base(binPath), info.Path, info.Main.Path, info.Main.Version)

	if *verbose {
		fmt.Print(info.String())
	}

	if *noUpdate {
		return
	}

	args := []string{"install", "-buildvcs=true"}
	env := os.Environ()
	for _, s := range info.Settings {
		if s.Key == "" {
			continue
		}
		if s.Key == "-tags" {
			args = append(args, s.Key, s.Value)
		}
		// Assume that all uppercase keys are build environment variables
		if s.Key == strings.ToUpper(s.Key) {
			// If not explicit in environment...
			if _, override := os.LookupEnv(s.Key); !override {
				// Use the value from the binary
				env = append(env, s.Key+"="+s.Value)
			}
		}
	}
	if version == "" {
		version = info.Main.Version
	}
	args = append(args, info.Path+"@"+version)

	cmd := exec.Command("go", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = env

	if *verbose {
		fmt.Println("go", args) // FIXME formatting
	}

	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to install %s: %v\n", info.Path, err)
		return
	}

	if *verbose {
		// After update, read again
		info, err = buildinfo.ReadFile(binPath)
		if err == nil {
			fmt.Print(info.String())
		}
	}
}
