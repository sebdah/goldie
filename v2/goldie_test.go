package goldie

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGoldenFileName(t *testing.T) {
	tests := map[string]struct {
		name     string
		options  []Option
		expected string
	}{
		"using defaults": {
			name:     "example",
			expected: fmt.Sprintf("%s/%s%s", defaultFixtureDir, "example", defaultFileNameSuffix),
		},
		"with custom suffix": {
			name: "example",
			options: []Option{
				WithNameSuffix(".txt"),
			},
			expected: fmt.Sprintf("%s/%s%s", defaultFixtureDir, "example", ".txt"),
		},
		"with custom fixture dir": {
			name: "example",
			options: []Option{
				WithFixtureDir("fixtures"),
			},
			expected: fmt.Sprintf("%s/%s%s", "fixtures", "example", defaultFileNameSuffix),
		},
		"using test name for dir": {
			name: "example",
			options: []Option{
				WithTestNameForDir(true),
			},
			expected: fmt.Sprintf("%s/%s/%s%s", defaultFixtureDir, t.Name(), "example", defaultFileNameSuffix),
		},
		"using sub test name for dir": {
			name: "example",
			options: []Option{
				WithSubTestNameForDir(true),
			},
			expected: fmt.Sprintf("%s/%s/%s%s", defaultFixtureDir, "using_sub_test_name_for_dir", "example", defaultFileNameSuffix),
		},
		"using sub/sub test name for dir": {
			name: "example",
			options: []Option{
				WithSubTestNameForDir(true),
			},
			expected: fmt.Sprintf("%s/%s/%s/%s%s", defaultFixtureDir, "using_sub", "sub_test_name_for_dir", "example", defaultFileNameSuffix),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			g := New(t, test.options...)
			assert.Equal(t, test.expected, g.GoldenFileName(t, test.name))
		})
	}
}

func TestEnsureDir(t *testing.T) {
	tests := map[string]struct {
		dir         string
		shouldExist bool
		fileExist   bool
		err         interface{}
	}{
		"with existing directory": {
			dir:         "example1",
			shouldExist: true,
			err:         nil,
		},
		"without existing directory": {
			dir:         "example2",
			shouldExist: false,
			fileExist:   false,
			err:         nil,
		},
		"with existing deep directory structure": {
			dir:         "now/still/works",
			shouldExist: true,
			err:         nil,
		},
		"error, fixture directory is a file": {
			dir:         "this/will/not",
			shouldExist: false,
			fileExist:   true,
			err:         newErrFixtureDirectoryIsFile(filepath.Join(os.TempDir(), "this/will/not")),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			g := New(t)
			target := filepath.Join(os.TempDir(), test.dir)

			if test.shouldExist {
				err := os.MkdirAll(target, g.dirPerms)
				assert.Nil(t, err)
			}

			if test.fileExist {
				err := os.MkdirAll(filepath.Dir(target), g.dirPerms)
				assert.Nil(t, err)

				f, err := os.Create(target)
				require.NoError(t, err)
				err = f.Close()
				assert.Nil(t, err)
			}

			err := g.ensureDir(target)
			assert.Equal(t, test.err, err)
			if err != nil {
				return
			}

			s, err := os.Stat(target)
			assert.Nil(t, err)
			assert.True(t, s.IsDir())
		})
	}
}

func TestMeta(t *testing.T) {
	tests := []struct {
		name     string
		data     interface{}
		expected map[string]string
	}{
		{
			name: "successful update",
			data: map[string]interface{}{
				"NAME": "example",
			},
			expected: map[string]string{
				"example": ".NAME",
			},
		},
		{
			name: "replace most specific keys first",
			data: map[string]string{
				"FOO":    "foo",
				"FOOBAR": "foobar",
			},
			expected: map[string]string{
				"foo":    ".FOO",
				"foobar": ".FOOBAR",
			},
		},
		{
			name: "int",
			data: 123,
			expected: map[string]string{
				"123": ".",
			},
		},
		{
			name: "scalar types in a map",
			data: map[string]interface{}{
				"NAME":   "example",
				"NUMBER": 123,
				"BOOL":   true,
				"FLOAT":  123.456,
			},
			expected: map[string]string{
				"example": ".NAME",
				"123":     ".NUMBER",
				"true":    ".BOOL",
				"123.456": ".FLOAT",
			},
		},
		{
			name: "scalar types in a struct",
			data: struct {
				Name   string
				Number int
				Bool   bool
				Float  float64
			}{
				Name:   "example",
				Number: 123,
				Bool:   true,
				Float:  123.456,
			},
			expected: map[string]string{
				"example": ".Name",
				"123":     ".Number",
				"true":    ".Bool",
				"123.456": ".Float",
			},
		},
		{
			name: "scalar types in a slice",
			data: []interface{}{
				"example",
				123,
				true,
				123.456,
			},
			expected: map[string]string{
				"example": "index (.) 0",
				"123":     "index (.) 1",
				"true":    "index (.) 2",
				"123.456": "index (.) 3",
			},
		},
		{
			name: "nested maps",
			data: map[string]interface{}{
				"FOO": map[string]interface{}{
					"BAR": "bar",
					"BAZ": "baz",
				},
			},
			expected: map[string]string{
				"bar": ".FOO.BAR",
				"baz": ".FOO.BAZ",
			},
		},
		{
			name: "nested structs",
			data: struct {
				Foo struct {
					Bar string
					Baz string
				}
			}{
				Foo: struct {
					Bar string
					Baz string
				}{
					Bar: "bar",
					Baz: "baz",
				},
			},
			expected: map[string]string{
				"bar": ".Foo.Bar",
				"baz": ".Foo.Baz",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, meta(test.data))
		})
	}
}

// TODO: This test could use a little <3. It should test some more negative
// cases.
func TestUpdate(t *testing.T) {
	tests := map[string]struct {
		name       string
		actualData []byte
		err        error
	}{
		"successful update": {
			name:       "abc",
			actualData: []byte("some example data"),
			err:        nil,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			g := New(t)
			err := g.Update(t, test.name, test.actualData)
			assert.Equal(t, test.err, err)

			data, err := os.ReadFile(g.GoldenFileName(t, test.name))
			assert.Nil(t, err)
			assert.Equal(t, string(test.actualData), string(data))

			err = os.RemoveAll(g.fixtureDir)
			assert.Nil(t, err)
		})
	}
}

func TestUpdateWithTemplate(t *testing.T) {
	tests := []struct {
		name         string
		data         interface{}
		actualData   []byte
		expectedData []byte
		err          error
	}{
		{
			name: "successful update",
			data: map[string]interface{}{
				"NAME": "example",
			},
			actualData:   []byte("abc example"),
			expectedData: []byte("abc {{.NAME}}"),
			err:          nil,
		},
		{
			name: "replace most specific keys first",
			data: map[string]string{
				"FOO":    "foo",
				"FOOBAR": "foobar",
			},
			actualData:   []byte("testing foo and foobar"),
			expectedData: []byte("testing {{.FOO}} and {{.FOOBAR}}"),
			err:          nil,
		},
		{
			name:         "int",
			data:         123,
			actualData:   []byte("abc 123"),
			expectedData: []byte("abc {{.}}"),
			err:          nil,
		},
		{
			name: "scalar types in a map",
			data: map[string]interface{}{
				"NAME":   "example",
				"NUMBER": 123,
				"BOOL":   true,
				"FLOAT":  123.456,
			},
			actualData:   []byte("abc example 123 true 123.456"),
			expectedData: []byte("abc {{.NAME}} {{.NUMBER}} {{.BOOL}} {{.FLOAT}}"),
			err:          nil,
		},
		{
			name: "scalar types in a struct",
			data: struct {
				Name   string
				Number int
				Bool   bool
				Float  float64
			}{
				Name:   "example",
				Number: 123,
				Bool:   true,
				Float:  123.456,
			},
			actualData:   []byte("abc example 123 true 123.456"),
			expectedData: []byte("abc {{.Name}} {{.Number}} {{.Bool}} {{.Float}}"),
			err:          nil,
		},
		{
			name: "scalar types in a slice",
			data: []interface{}{
				"example",
				123,
				true,
				123.456,
			},
			actualData:   []byte("abc example 123 true 123.456"),
			expectedData: []byte("abc {{index (.) 0}} {{index (.) 1}} {{index (.) 2}} {{index (.) 3}}"),
			err:          nil,
		},
		{
			name: "nested maps",
			data: map[string]interface{}{
				"FOO": map[string]interface{}{
					"BAR": "bar",
					"BAZ": "baz",
				},
			},
			actualData:   []byte("abc bar baz"),
			expectedData: []byte("abc {{.FOO.BAR}} {{.FOO.BAZ}}"),
			err:          nil,
		},
		{
			name: "nested structs",
			data: struct {
				Foo struct {
					Bar string
					Baz string
				}
			}{
				Foo: struct {
					Bar string
					Baz string
				}{
					Bar: "bar",
					Baz: "baz",
				},
			},
			actualData:   []byte("abc bar baz"),
			expectedData: []byte("abc {{.Foo.Bar}} {{.Foo.Baz}}"),
			err:          nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := New(t)
			err := g.UpdateWithTemplate(t, test.name, test.data, test.actualData)
			assert.Equal(t, test.err, err)

			got, err := os.ReadFile(g.GoldenFileName(t, test.name))
			assert.Nil(t, err)
			assert.Equal(t, string(test.expectedData), string(got))

			err = os.RemoveAll(g.fixtureDir)
			assert.Nil(t, err)
		})
	}
}

func TestDiffEngines(t *testing.T) {
	type engine struct {
		engine DiffEngine
		diff   string
	}

	tests := map[string]struct {
		actual   string
		expected string
		engine   engine
	}{
		"simple": {
			actual:   "Lorem ipsum dolor.",
			expected: "Lorem dolor sit amet.",
			engine: engine{
				engine: Simple,
				diff: `Expected: Lorem dolor sit amet.
Got: Lorem ipsum dolor.`},
		},
		"classic": {
			actual:   "Lorem ipsum dolor.",
			expected: "Lorem dolor sit amet.",
			engine: engine{
				engine: ClassicDiff,
				diff: `--- Expected
+++ Actual
@@ -1 +1 @@
-Lorem dolor sit amet.
+Lorem ipsum dolor.
`},
		},
		"colored": {
			actual:   "Lorem ipsum dolor.",
			expected: "Lorem dolor sit amet.",
			engine: engine{
				engine: ColoredDiff,
				diff:   "Lorem \x1b[31mipsum \x1b[0mdolor\x1b[32m sit amet\x1b[0m.",
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(
				t,
				test.engine.diff,
				Diff(test.engine.engine, test.actual, test.expected),
			)
		})
	}

}

func TestCleanFunction(t *testing.T) {
	savedCleanState := *clean
	*clean = false
	savedUpdateState := *update
	*update = true
	t.Cleanup(func() {
		*clean = savedCleanState
		*update = savedUpdateState
	})

	ts = time.Now()

	sampleData := []byte("sample data")
	fixtureDir := "test-fixtures"
	fixtureSubDirA := fixtureDir + "/a"
	fixtureSubDirB := fixtureDir + "/b"
	fixtureNestedDirAC := fixtureSubDirA + "/c"
	suffix := ".golden"

	t.Cleanup(func() {
		err := os.RemoveAll(fixtureDir)
		assert.Nil(t, err)
	})

	// The first time running go test, with -update, without -clean
	firstTests := []struct {
		fixtureDirWithSub string
		filePrefix        string
	}{
		{fixtureDirWithSub: fixtureSubDirA, filePrefix: "example-a1"},
		{fixtureDirWithSub: fixtureNestedDirAC, filePrefix: "example-ac1"},
		{fixtureDirWithSub: fixtureSubDirA, filePrefix: "example-a2"},
		{fixtureDirWithSub: fixtureSubDirB, filePrefix: "example-b1"},
		{fixtureDirWithSub: fixtureSubDirB, filePrefix: "example-b2"},
	}

	for i, tt := range firstTests {
		g := New(t,
			WithFixtureDir(tt.fixtureDirWithSub),
			WithNameSuffix(suffix),
		)

		t.Run(fmt.Sprint(i), func(t *testing.T) {
			g.Assert(t, tt.filePrefix, sampleData)
		})

		fullPath := fmt.Sprintf("%s%s",
			filepath.Join(tt.fixtureDirWithSub, tt.filePrefix),
			suffix,
		)

		_, err := os.Stat(fullPath)
		assert.Nil(t, err)
	}

	*clean = true
	ts = time.Now()

	// The second time running go test, with -update and -clean
	secondTests := []struct {
		fixtureDirWithSub string
		filePrefix        string
	}{
		{fixtureDirWithSub: fixtureSubDirA, filePrefix: "example-a3"},
		{fixtureDirWithSub: fixtureNestedDirAC, filePrefix: "example-ac2"},
		{fixtureDirWithSub: fixtureSubDirA, filePrefix: "example-a4"},
		{fixtureDirWithSub: fixtureSubDirB, filePrefix: "example-b3"},
		{fixtureDirWithSub: fixtureSubDirB, filePrefix: "example-b4"},
	}

	for i, tt := range secondTests {
		g := New(t,
			WithFixtureDir(tt.fixtureDirWithSub),
			WithNameSuffix(suffix),
		)

		t.Run(fmt.Sprint(i), func(t *testing.T) {
			g.Assert(t, tt.filePrefix, sampleData)
		})
	}

	// make sure all the output files from the second run now exist
	for _, tt := range secondTests {
		_, err := os.Stat(fmt.Sprintf("%s%s",
			filepath.Join(tt.fixtureDirWithSub, tt.filePrefix),
			suffix,
		))
		assert.Nil(t, err)
	}

	// make sure output files' of the first run doesn't exist
	for _, tt := range firstTests {
		fullPath := fmt.Sprintf("%s%s",
			filepath.Join(tt.fixtureDirWithSub, tt.filePrefix),
			suffix,
		)

		_, err := os.Stat(fullPath)
		assert.Error(t, err)
		assert.True(t, os.IsNotExist(err))
	}
}
