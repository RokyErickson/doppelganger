package main

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/RokyErickson/doppelganger/cmd"
	"github.com/RokyErickson/doppelganger/pkg/agent"
	"github.com/RokyErickson/doppelganger/pkg/doppelganger"
	"github.com/polydawn/gosh"
	"github.com/spf13/pflag"
)

const (
	agentPackage                 = "github.com/RokyErickson/doppelganger/cmd/doppelganger-agent"
	cliPackage                   = "github.com/RokyErickson/doppelganger/cmd/doppelganger"
	buildDirectoryName           = "build"
	agentBuildSubdirectoryName   = "agent"
	cliBuildSubdirectoryName     = "cli"
	releaseBuildSubdirectoryName = "release"
	agentBaseName                = "doppelganger-agent"
	cliBaseName                  = "doppelganger"
	minimumMacOSVersion          = "10.10"
	minimumARMSupport            = "5"
)

type Target struct {
	GOOS   string
	GOARCH string
}

func (t Target) String() string {
	return fmt.Sprintf("%s/%s", t.GOOS, t.GOARCH)
}

func (t Target) Name() string {
	return fmt.Sprintf("%s_%s", t.GOOS, t.GOARCH)
}

func (t Target) ExecutableName(base string) string {

	if t.GOOS == "windows" {
		return fmt.Sprintf("%s.exe", base)
	}

	return base
}

func (t Target) goEnv() ([]string, error) {

	result := os.Environ()

	result = append(result, "GO111MODULE=on")

	result = append(result, fmt.Sprintf("GOOS=%s", t.GOOS))
	result = append(result, fmt.Sprintf("GOARCH=%s", t.GOARCH))

	if t.GOOS == "darwin" && t.GOARCH == "amd64" {
		result = append(result, fmt.Sprintf("CGO_CFLAGS=-mmacosx-version-min=%s", minimumMacOSVersion))
		result = append(result, fmt.Sprintf("CGO_LDFLAGS=-mmacosx-version-min=%s", minimumMacOSVersion))
	} else {
		result = append(result, "CGO_ENABLED=0")
	}

	if t.GOARCH == "arm" {
		result = append(result, fmt.Sprintf("GOARM=%s", minimumARMSupport))
	}

	return result, nil
}

func (t Target) Cross() bool {
	return t.GOOS != runtime.GOOS || t.GOARCH != runtime.GOARCH
}

func (t Target) Build(url, output string) gosh.Proc {

	builder := gosh.Gosh("go", "build", "-o", output, "-ldflags=-s -w", url)

	return builder.Run()
}

var targets = []Target{

	{"darwin", "amd64"},

	// {"darwin", "arm64"},
	{"dragonfly", "amd64"},
	{"freebsd", "386"},
	{"freebsd", "amd64"},
	{"freebsd", "arm"},
	{"linux", "386"},
	{"linux", "amd64"},
	{"linux", "arm"},
	{"linux", "arm64"},
	{"linux", "ppc64"},
	{"linux", "ppc64le"},
	{"linux", "mips"},
	{"linux", "mipsle"},
	{"linux", "mips64"},
	{"linux", "mips64le"},
	{"linux", "s390x"},
	{"netbsd", "386"},
	{"netbsd", "amd64"},
	{"netbsd", "arm"},
	{"openbsd", "386"},
	{"openbsd", "amd64"},
	{"openbsd", "arm"},
	{"solaris", "amd64"},
	{"windows", "386"},
	{"windows", "amd64"},
}

const archiveBuilderCopyBufferSize = 32 * 1024

type ArchiveBuilder struct {
	file       *os.File
	compressor *gzip.Writer
	archiver   *tar.Writer
	copyBuffer []byte
}

func NewArchiveBuilder(bundlePath string) (*ArchiveBuilder, error) {

	file, err := os.Create(bundlePath)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create target file")
	}

	compressor, err := gzip.NewWriterLevel(file, gzip.BestCompression)
	if err != nil {
		file.Close()
		return nil, errors.Wrap(err, "unable to create compressor")
	}

	return &ArchiveBuilder{
		file:       file,
		compressor: compressor,
		archiver:   tar.NewWriter(compressor),
		copyBuffer: make([]byte, archiveBuilderCopyBufferSize),
	}, nil
}

func (b *ArchiveBuilder) Close() error {

	if err := b.archiver.Close(); err != nil {
		b.compressor.Close()
		b.file.Close()
		return errors.Wrap(err, "unable to close archiver")
	} else if err := b.compressor.Close(); err != nil {
		b.file.Close()
		return errors.Wrap(err, "unable to close compressor")
	} else if err := b.file.Close(); err != nil {
		return errors.Wrap(err, "unable to close file")
	}

	return nil
}

func (b *ArchiveBuilder) Add(name, path string, mode int64) error {

	if name == "" {
		name = filepath.Base(path)
	}

	file, err := os.Open(path)
	if err != nil {
		return errors.Wrap(err, "unable to open file")
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return errors.Wrap(err, "unable to determine file size")
	}
	size := stat.Size()

	header := &tar.Header{
		Name:    name,
		Mode:    mode,
		Size:    size,
		ModTime: time.Now(),
	}
	if err := b.archiver.WriteHeader(header); err != nil {
		return errors.Wrap(err, "unable to write archive header")
	}

	if _, err := io.CopyBuffer(b.archiver, file, b.copyBuffer); err != nil {
		return errors.Wrap(err, "unable to write archive entry")
	}

	return nil
}

func buildAgentForTargetInTesting(target Target) bool {
	return !target.Cross() ||
		target.GOOS == "darwin" ||
		target.GOOS == "windows" ||
		(target.GOOS == "linux" && (target.GOARCH == "amd64" || target.GOARCH == "arm")) ||
		(target.GOOS == "freebsd" && target.GOARCH == "amd64")
}

var usage = `usage: build [-h|--help] [-m|--mode=<mode>] [-s|--skip-bundles]

The mode flag takes three values: 'slim', 'testing', and 'release'.
`

func main() {
	flagSet := pflag.NewFlagSet("build", pflag.ContinueOnError)
	flagSet.SetOutput(ioutil.Discard)
	var mode string
	var skipBundles bool
	flagSet.StringVarP(&mode, "mode", "m", "slim", "specify the build mode")
	flagSet.BoolVarP(&skipBundles, "skip-bundles", "s", false, "skip release bundle building")
	if err := flagSet.Parse(os.Args[1:]); err != nil {
		if err == pflag.ErrHelp {
			fmt.Fprint(os.Stdout, usage)
			return
		} else {
			cmd.Fatal(errors.Wrap(err, "unable to parse command line"))
		}
	}
	if mode != "slim" && mode != "testing" && mode != "release" {
		cmd.Fatal(errors.New("invalid build mode"))
	}

	if runtime.GOOS != "darwin" {
		if mode == "release" {
			cmd.Fatal(errors.New("macOS required for release builds"))
		} else if mode == "testing" {
			cmd.Warning("macOS agents will be built without cgo support")
		}
	}

	doppelgangerSourcePath, err := doppelganger.SourceTreePath()
	if err != nil {
		cmd.Fatal(errors.Wrap(err, "unable to compute Doppelganger source tree path"))
	}

	workingDirectory, err := os.Getwd()
	if err != nil {
		cmd.Fatal(errors.Wrap(err, "unable to compute working directory"))
	}
	workingDirectoryRelativePath, err := filepath.Rel(doppelgangerSourcePath, workingDirectory)
	if err != nil {
		cmd.Fatal(errors.Wrap(err, "unable to determine working directory relative path"))
	}
	if strings.Contains(workingDirectoryRelativePath, "..") {
		cmd.Fatal(errors.New("build script run outside Doppelganger source tree"))
	}

	buildPath := filepath.Join(doppelgangerSourcePath, buildDirectoryName)
	if err := os.MkdirAll(buildPath, 0700); err != nil {
		cmd.Fatal(errors.Wrap(err, "unable to create build directory"))
	}

	agentBuildSubdirectoryPath := filepath.Join(buildPath, agentBuildSubdirectoryName)
	cliBuildSubdirectoryPath := filepath.Join(buildPath, cliBuildSubdirectoryName)
	releaseBuildSubdirectoryPath := filepath.Join(buildPath, releaseBuildSubdirectoryName)
	if err := os.MkdirAll(agentBuildSubdirectoryPath, 0700); err != nil {
		cmd.Fatal(errors.Wrap(err, "unable to create agent build subdirectory"))
	}
	if err := os.MkdirAll(cliBuildSubdirectoryPath, 0700); err != nil {
		cmd.Fatal(errors.Wrap(err, "unable to create CLI build subdirectory"))
	}
	if mode == "release" && !skipBundles {
		if err := os.MkdirAll(releaseBuildSubdirectoryPath, 0700); err != nil {
			cmd.Fatal(errors.Wrap(err, "unable to create release build subdirectory"))
		}
	}

	log.Println("Building agent bundle...")
	agentBundlePath := filepath.Join(buildPath, agent.BundleName)
	agentBundle, err := NewArchiveBuilder(agentBundlePath)
	if err != nil {
		cmd.Fatal(errors.Wrap(err, "unable to create agent archive builder"))
	}
	for _, target := range targets {

		if mode == "slim" && target.Cross() {
			continue
		} else if mode == "testing" && !buildAgentForTargetInTesting(target) {
			continue
		}

		log.Println("Building agent for", target)

		agentBuildPath := filepath.Join(agentBuildSubdirectoryPath, target.Name())

		target.Build(agentPackage, agentBuildPath)

		if err := agentBundle.Add(target.Name(), agentBuildPath, 0700); err != nil {
			agentBundle.Close()
			cmd.Fatal(errors.Wrap(err, "unable to add agent to bundle"))
		}
	}
	if err := agentBundle.Close(); err != nil {
		cmd.Fatal(errors.Wrap(err, "unable to finalize agent bundle"))
	}

	for _, target := range targets {

		if mode != "release" && target.Cross() {
			continue
		}

		log.Println("Building CLI for", target)

		cliBuildPath := filepath.Join(cliBuildSubdirectoryPath, target.Name())

		target.Build(cliPackage, cliBuildPath)

		if mode == "release" && !skipBundles {

			log.Println("Building release bundle for", target)

			bundlePath := filepath.Join(
				releaseBuildSubdirectoryPath,
				fmt.Sprintf("doppelganger_%s_v%s.tar.gz", target.Name(), doppelganger.Version),
			)

			bundle, err := NewArchiveBuilder(bundlePath)
			if err != nil {
				cmd.Fatal(errors.Wrap(err, "unable to create release bundle"))
			}

			if err := bundle.Add(target.ExecutableName(cliBaseName), cliBuildPath, 0700); err != nil {
				bundle.Close()
				cmd.Fatal(errors.Wrap(err, "unable to bundle CLI"))
			}
			if err := bundle.Add("", agentBundlePath, 0600); err != nil {
				bundle.Close()
				cmd.Fatal(errors.Wrap(err, "unable to bundle agent bundle"))
			}

			if err := bundle.Close(); err != nil {
				cmd.Fatal(errors.Wrap(err, "unable to finalize release bundle"))
			}
		}

		if !target.Cross() {

			log.Println("Relocating binary for testing")

			targetPath := filepath.Join(buildPath, target.ExecutableName(cliBaseName))
			if err := os.Rename(cliBuildPath, targetPath); err != nil {
				cmd.Fatal(errors.Wrap(err, "unable to relocate platform CLI"))
			}
		}
	}
}
