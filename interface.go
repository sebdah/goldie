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

	"github.com/pmezard/go-difflib/difflib"
	"github.com/sergi/go-diff/diffmatchpatch"
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

// Option defines the signature of a functional option method that can apply
// options to an OptionProcessor.
type Option func(OptionProcessor) error

// Tester defines the methods that any golden tester should support.
type Tester interface {
	Assert(t *testing.T, name string, actualData []byte)
	AssertJson(t *testing.T, name string, actualJsonData interface{})
	AssertWithTemplate(t *testing.T, name string, data interface{}, actualData []byte)
	Update(t *testing.T, name string, actualData []byte) error
}

// DiffFn takes in an actual and expected and will return a diff string
// representing the differences between the two.
type DiffFn func(actual string, expected string) string

// DiffEngine is used to enumerate the diff engine processors that are
// available.
type DiffEngine int

const (
	// UndefinedDiff represents any undefined diff processor.  If a new diff
	// engine is implemented, it should be added to this enumeration and to the
	// `diff` helper function.
	UndefinedDiff DiffEngine = iota

	// ClassicDiff produces a diff similar to what the `diff` tool would
	// produce.
	//		+++ Actual
	//		@@ -1 +1 @@
	//		-Lorem dolor sit amet.
	//		+Lorem ipsum dolor.
	//
	ClassicDiff

	// ColoredDiff produces a diff that will use red and green colors to
	// distinguish the diffs between the two values.
	ColoredDiff
)

// OptionProcessor defines the functions that can be called to set values for
// a tester.  To expand this list, add a function to this interface and then
// implement the generic option setter below.
type OptionProcessor interface {
	// WithFixtureDir sets the directory that will be used to store the
	// fixtures.
	//
	// Defaults to `testdata`.
	WithFixtureDir(dir string) error
	WithNameSuffix(suffix string) error
	WithFilePerms(mode os.FileMode) error
	WithDirPerms(mode os.FileMode) error

	WithDiffEngine(engine DiffEngine) error
	WithDiffFn(fn DiffFn) error
	WithIgnoreTemplateErrors(ignoreErrors bool) error
	WithTestNameForDir(use bool) error
	WithSubTestNameForDir(use bool) error
}

// === OptionProcessor ===============================

func WithFixtureDir(dir string) Option {
	return func(o OptionProcessor) error {
		return o.WithFixtureDir(dir)
	}
}

// WithNameSuffix sets the file suffix to be used for the golden file.
//
// Defaults to `.golden`
func WithNameSuffix(suffix string) Option {
	return func(o OptionProcessor) error {
		return o.WithNameSuffix(suffix)
	}
}

// WithFilePerms sets the file permissions on the golden files that are
// created.
//
// Defaults to 0644.
func WithFilePerms(mode os.FileMode) Option {
	return func(o OptionProcessor) error {
		return o.WithFilePerms(mode)
	}
}

// WithDirPerms sets the directory permissions for the directories in which the
// golden files are created.
//
// Defaults to 0755.
func WithDirPerms(mode os.FileMode) Option {
	return func(o OptionProcessor) error {
		return o.WithDirPerms(mode)
	}
}

// WithDiffEngine sets the `diff` engine that will be used to generate the
// `diff` text.
func WithDiffEngine(engine DiffEngine) Option {
	return func(o OptionProcessor) error {
		return o.WithDiffEngine(engine)
	}
}

// WithDiffFn sets the `diff` engine to be a function that implements the
// DiffFn signature. This allows for any customized diff logic you would like
// to create.
func WithDiffFn(fn DiffFn) Option {
	return func(o OptionProcessor) error {
		return o.WithDiffFn(fn)
	}
}

// WithIgnoreTemplateErrors allows template processing to ignore any variables
// in the template that do not have corresponding data values passed in.
//
// Default value is false.
func WithIgnoreTemplateErrors(ignoreErrors bool) Option {
	return func(o OptionProcessor) error {
		return o.WithIgnoreTemplateErrors(ignoreErrors)
	}
}

// WithTestNameForDir will create a directory with the test's name in the
// fixture directory to store all the golden files.
func WithTestNameForDir(use bool) Option {
	return func(o OptionProcessor) error {
		return o.WithTestNameForDir(use)
	}
}

// WithSubTestNameForDir will create a directory with the sub test's name to
// store all the golden files. If WithTestNameForDir is enabled, it will be in
// the test name's directory. Otherwise, it will be in the fixture directory.
func WithSubTestNameForDir(use bool) Option {
	return func(o OptionProcessor) error {
		return o.WithSubTestNameForDir(use)
	}
}

// === Create new testers ==================================

// New creates a new golden file tester. If there is an issue with applying any
// of the options, an error will be reported and t.FailNow() will be called.
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
			t.FailNow()
		}
	}

	return &g
}

// Diff generates a string that shows the difference between the actual and the
// expected. This method could be called in your own DiffFn in case you want
// to leverage any of the engines defined.
func Diff(engine DiffEngine, actual string, expected string) string {
	var diff string
	switch engine {
	case ClassicDiff:
		diff, _ = difflib.GetUnifiedDiffString(difflib.UnifiedDiff{
			A:        difflib.SplitLines(expected),
			B:        difflib.SplitLines(actual),
			FromFile: "Expected",
			FromDate: "",
			ToFile:   "Actual",
			ToDate:   "",
			Context:  1,
		})

	case ColoredDiff:
		dmp := diffmatchpatch.New()
		diffs := dmp.DiffMain(actual, expected, false)
		diff = dmp.DiffPrettyText(diffs)
	}
	return diff
}
