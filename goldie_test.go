package goldie

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
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

		filename := g.GoldenFileName(t, test.name)
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
			err = f.Close()
			assert.Nil(t, err)
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
		err := g.Update(t, test.name, test.data)
		assert.Equal(t, test.err, err)

		data, err := ioutil.ReadFile(g.GoldenFileName(t, test.name))
		assert.Nil(t, err)
		assert.Equal(t, test.data, data)

		err = os.RemoveAll(g.fixtureDir)
		assert.Nil(t, err)
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

func TestCleanFunction(t *testing.T) {
	savedCleanState := *clean
	*clean = false
	savedUpdateState := *update
	*update = true
	ts = time.Now()

	sampleData := []byte("sample data")
	fixtureDir := "test-fixtures"
	fixtureSubDirA := fixtureDir + "/a"
	fixtureSubDirB := fixtureDir + "/b"
	suffix := ".golden"

	// The first time running go test, with -update, without -clean
	firstTests := []struct {
		fixtureDirWithSub string
		filePrefix        string
	}{
		{fixtureDirWithSub: fixtureSubDirA, filePrefix: "example-a1"},
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

		fullPath := fmt.Sprintf("%s%s",
			filepath.Join(tt.fixtureDirWithSub, tt.filePrefix),
			suffix,
		)

		_, err := os.Stat(fullPath)
		assert.Nil(t, err)
	}

	// make sure output files of the first run doesnt exist
	for _, tt := range firstTests {
		fullPath := fmt.Sprintf("%s%s",
			filepath.Join(tt.fixtureDirWithSub, tt.filePrefix),
			suffix,
		)

		_, err := os.Stat(fullPath)
		assert.Error(t, err)
		assert.True(t, os.IsNotExist(err))
	}

	err := os.RemoveAll(fixtureDir)
	assert.Nil(t, err)
	*clean = savedCleanState
	*update = savedUpdateState
}
