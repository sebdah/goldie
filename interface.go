// Package goldie provides test assertions based on golden files. It's typically
// used for testing responses with larger data bodies.
//
// The concept is straight forward. Valid response data is stored in a "golden
// file". The actual response data will be byte compared with the golden file
// and the test will fail if there is a difference.
//
// Updating the golden file can be done by running `go test -update ./...`.
package goldie

import (
	"flag"
	"fmt"
	"os"
	"testing"
)

const (
	// FixtureDir is the folder name for where the fixtures are stored. It's
	// relative to the "go test" path.
	defaultFixtureDir = "testdata"

	// FileNameSuffix is the suffix appended to the fixtures. Set to empty
	// string to disable file name suffixes.
	defaultFileNameSuffix = ".golden"

	// FilePerms is used to set the permissions on the golden fixture files.
	defaultFilePerms os.FileMode = 0644

	// DirPerms is used to set the permissions on the golden fixture folder.
	defaultDirPerms os.FileMode = 0755
)

var (
	// update determines if the actual received data should be written to the
	// golden files or not. This should be true when you need to update the
	// data, but false when actually running the tests.
	update = flag.Bool("update", false, "Update golden test file fixture")
)

type Option func(OptionProcessor) error

type Tester interface {
	Assert(t *testing.T, name string, actualData []byte)
	AssertJson(t *testing.T, name string, actualJsonData interface{})
	AssertWithTemplate(t *testing.T, name string, data interface{}, actualData []byte)
	Update(t *testing.T, name string, actualData []byte) error
}

// DiffFn takes in an actual and expected and will return a diff string
// representing the differences between the two
type DiffFn func(actual string, expected string) string
type DiffProcessor int

const (
	UndefinedDiff DiffProcessor = iota
	ClassicDiff
	ColoredDiff
)

type OptionProcessor interface {
	WithFixtureDir(dir string) error
	WithNameSuffix(suffix string) error
	WithFilePerms(mode os.FileMode) error
	WithDirPerms(mode os.FileMode) error

	WithDiffEngine(engine DiffProcessor) error
	WithDiffFn(fn DiffFn) error
	WithIgnoreTemplateErrors(ignoreErrors bool) error
	WithTestNameForDir(use bool) error
	WithSubTestNameForDir(use bool) error
}

func New(t *testing.T, options ...Option) *goldie {
	g := goldie{
		fixtureDir:     defaultFixtureDir,
		fileNameSuffix: defaultFileNameSuffix,
		filePerms:      defaultFilePerms,
		dirPerms:       defaultDirPerms,
	}

	var err error
	for _, option := range options {
		err = option(&g)
		if err != nil {
			t.Error(fmt.Errorf("Could not apply option: %w", err))
		}
	}

	return &g
}

// === OptionProcessor ===============================

func WithFixtureDir(dir string) Option {
	return func(o OptionProcessor) error {
		return o.WithFixtureDir(dir)
	}
}

func WithNameSuffix(suffix string) Option {
	return func(o OptionProcessor) error {
		return o.WithNameSuffix(suffix)
	}
}

func WithFilePerms(mode os.FileMode) Option {
	return func(o OptionProcessor) error {
		return o.WithFilePerms(mode)
	}
}

func WithDirPerms(mode os.FileMode) Option {
	return func(o OptionProcessor) error {
		return o.WithDirPerms(mode)
	}
}

func WithDiffEngine(engine DiffProcessor) Option {
	return func(o OptionProcessor) error {
		return o.WithDiffEngine(engine)
	}
}

func WithIgnoreTemplateErrors(ignoreErrors bool) Option {
	return func(o OptionProcessor) error {
		return o.WithIgnoreTemplateErrors(ignoreErrors)
	}
}

func WithTestNameForDir(use bool) Option {
	return func(o OptionProcessor) error {
		return o.WithTestNameForDir(use)
	}
}

func WithSubTestNameForDir(use bool) Option {
	return func(o OptionProcessor) error {
		return o.WithSubTestNameForDir(use)
	}
}
