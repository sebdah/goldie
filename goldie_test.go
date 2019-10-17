package goldie

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGoldenFileName(t *testing.T) {
	tests := []struct {
		dir      string
		name     string
		suffix   string
		expected string
	}{
		{
			dir:      "fixtures",
			name:     "example-name",
			suffix:   ".suffix",
			expected: "fixtures/example-name.suffix",
		},
		{
			dir:      "",
			name:     "example-name",
			suffix:   ".suffix",
			expected: "example-name.suffix",
		},
		{
			dir:      "fixtures",
			name:     "",
			suffix:   ".suffix",
			expected: "fixtures/.suffix",
		},
		{
			dir:      "fixtures",
			name:     "example-name",
			suffix:   "",
			expected: "fixtures/example-name",
		},
	}

	for _, test := range tests {
		g := New(t,
			WithFixtureDir(test.dir),
			WithNameSuffix(test.suffix),
		)

		filename := g.goldenFileName(test.name)
		assert.Equal(t, test.expected, filename)
	}
}

func TestEnsureDir(t *testing.T) {
	tests := []struct {
		dir         string
		shouldExist bool
		fileExist   bool
		err         interface{}
	}{
		{
			dir:         "example1",
			shouldExist: true,
			err:         nil,
		},
		{
			dir:         "example2",
			shouldExist: false,
			err:         nil,
		},
		{
			dir:         "now/still/works",
			shouldExist: true,
			err:         nil,
		},
		{
			dir:         "this/will/not",
			shouldExist: false,
			fileExist:   true,
			err:         newErrFixtureDirectoryIsFile(""),
		},
	}

	g := New(t)

	for _, test := range tests {
		target := filepath.Join(os.TempDir(), test.dir)

		if test.shouldExist {
			err := os.MkdirAll(target, 0755)
			assert.Nil(t, err)
		}

		if test.fileExist {
			err := os.MkdirAll(filepath.Dir(target), 0755)
			assert.Nil(t, err)

			f, err := os.Create(target)
			require.NoError(t, err)
			f.Close()
		}

		err := g.ensureDir(target)
		assert.IsType(t, test.err, err)
	}
}

// TODO: This test could use a little <3. It should test some more negative
// cases.
func TestUpdate(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		err  error
	}{
		{
			name: "abc",
			data: []byte("some example data"),
			err:  nil,
		},
	}

	g := New(t)

	for _, test := range tests {
		err := g.Update(test.name, test.data)
		assert.Equal(t, test.err, err)

		data, err := ioutil.ReadFile(g.goldenFileName(test.name))
		assert.Nil(t, err)
		assert.Equal(t, test.data, data)

		err = os.RemoveAll(g.fixtureDir)
		assert.Nil(t, err)
	}
}

func TestCompare(t *testing.T) {
	tests := []struct {
		name         string
		actualData   []byte
		expectedData []byte
		update       bool
		err          error
	}{
		{
			name:         "example",
			actualData:   []byte("abc"),
			expectedData: []byte("abc"),
			update:       true,
			err:          nil,
		},
		{
			name:         "example",
			actualData:   []byte("abc"),
			expectedData: []byte("abc"),
			update:       false,
			err:          &errFixtureNotFound{},
		},
		{
			name:         "example",
			actualData:   []byte("bc"),
			expectedData: []byte("abc"),
			update:       true,
			err:          &errFixtureMismatch{},
		},
		{
			name:         "nil",
			actualData:   nil,
			expectedData: nil,
			update:       true,
			err:          nil,
		},
	}

	g := New(t)

	for _, test := range tests {
		if test.update {
			err := g.Update(test.name, test.expectedData)
			assert.Nil(t, err)
		}

		err := g.compare(test.name, test.actualData)
		assert.IsType(t, test.err, err)

		g.goldenFileName(test.name)
		err = os.RemoveAll(filepath.Dir(g.goldenFileName(test.name)))
		assert.Nil(t, err)
	}
}

func TestCompareTemplate(t *testing.T) {
	data := struct {
		Name string
	}{
		Name: "example",
	}

	tests := []struct {
		name         string
		actualData   []byte
		expectedData []byte
		data         interface{}
		update       bool
		err          error
	}{
		{
			name:         "example",
			actualData:   []byte("abc example"),
			expectedData: []byte("abc {{ .Name }}"),
			data:         data,
			update:       true,
			err:          nil,
		},
		{
			name:         "example",
			actualData:   []byte("abc example"),
			expectedData: []byte("abc {{ .Name }}"),
			data:         nil,
			update:       false,
			err:          &errFixtureNotFound{},
		},
		{
			name:         "example",
			actualData:   []byte("bc example"),
			expectedData: []byte("abc {{ .Name }}"),
			data:         data,
			update:       true,
			err:          &errFixtureMismatch{},
		},
		{
			name:         "example",
			actualData:   []byte("bc example"),
			expectedData: []byte("abc {{ .Name }}"),
			data:         nil,
			update:       true,
			err:          &errMissingKey{},
		}}

	g := New(t)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.update {
				err := g.Update(test.name, test.expectedData)
				assert.Nil(t, err)
			}

			err := g.compareTemplate(test.name, test.data, test.actualData)
			assert.IsType(t, test.err, err)

			err = os.RemoveAll(g.fixtureDir)
			assert.Nil(t, err)
		})
	}
}

func TestNormalizeLF(t *testing.T) {
	tests := []struct {
		name         string
		inputData    []byte
		expectedData []byte
	}{
		{"windows-style", []byte("Hello\r\nWorld"), []byte("Hello\nWorld")},
		{"mac-style", []byte("Hello\rWorld"), []byte("Hello\nWorld")},
		{"unix-style", []byte("Hello\nWorld"), []byte("Hello\nWorld")},
		{"empty-slice", []byte(""), []byte{}},
		{"nil-input", nil, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if actualData := normalizeLF(tt.inputData); !reflect.DeepEqual(actualData, tt.expectedData) {
				t.Errorf("normalizeLF() = %v, want %v", actualData, tt.expectedData)
			}
		})
	}
}

func TestDiffEngines(t *testing.T) {
	type engine struct {
		engine DiffEngine
		diff   string
	}

	tests := []struct {
		name     string
		actual   string
		expected string
		engines  []engine
	}{
		{
			name:     "lorem",
			actual:   "Lorem ipsum dolor.",
			expected: "Lorem dolor sit amet.",
			engines: []engine{
				{engine: ClassicDiff, diff: `--- Expected
+++ Actual
@@ -1 +1 @@
-Lorem dolor sit amet.
+Lorem ipsum dolor.
`},
				{engine: ColoredDiff, diff: "Lorem \x1b[31mipsum \x1b[0mdolor\x1b[32m sit amet\x1b[0m."},
			},
		},
	}

	for _, tt := range tests {
		for _, e := range tt.engines {
			diff := Diff(e.engine, tt.actual, tt.expected)
			assert.Equal(t, e.diff, diff)
		}
	}

}

func TestNewExample(t *testing.T) {
	tests := map[string]struct {
		fixtureDir string // This will get removed from the file system for each test
		suffix     string
		filePrefix string
	}{
		"with-custom-fixtureDir-prefix-and-suffix": {
			fixtureDir: "test-fixtures",
			suffix:     ".golden.json",
			filePrefix: "example",
		},
		"with-prefix-and-suffix": {
			suffix:     ".golden.json",
			filePrefix: "example",
		},
	}

	sampleData := []byte("sample data")

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			g := New(
				t,
				WithFixtureDir(test.fixtureDir),
				WithNameSuffix(test.suffix),
				WithTestNameForDir(true),
				WithSubTestNameForDir(true),
			)

			g.Update(test.filePrefix, sampleData)
			g.Assert(test.filePrefix, sampleData)

			fullpath := fmt.Sprintf("%s%s",
				filepath.Join(
					test.fixtureDir,
					"TestNewExample",
					name,
					test.filePrefix,
				),
				test.suffix,
			)

			_, err := os.Stat(fullpath)
			assert.Nil(t, err)

			os.RemoveAll(test.fixtureDir)
			assert.Nil(t, err)
		})
	}
}
