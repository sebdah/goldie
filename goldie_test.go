package goldie

import (
	"io/ioutil"
	"os"
	"path/filepath"
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
		oldFixtureDir := FixtureDir
		oldFileNameSuffix := FileNameSuffix

		FixtureDir = test.dir
		FileNameSuffix = test.suffix

		filename := goldenFileName(test.name)
		assert.Equal(t, test.expected, filename)

		FixtureDir = oldFixtureDir
		FileNameSuffix = oldFileNameSuffix
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

		err := ensureDir(target)
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

	for _, test := range tests {
		err := Update(test.name, test.data)
		assert.Equal(t, test.err, err)

		data, err := ioutil.ReadFile(goldenFileName(test.name))
		assert.Nil(t, err)
		assert.Equal(t, test.data, data)

		err = os.RemoveAll(FixtureDir)
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
			err:          errFixtureNotFound{},
		},
		{
			name:         "example",
			actualData:   []byte("bc"),
			expectedData: []byte("abc"),
			update:       true,
			err:          errFixtureMismatch{},
		},
	}

	for _, test := range tests {
		if test.update {
			err := Update(test.name, test.expectedData)
			assert.Nil(t, err)
		}

		err := compare(test.name, test.actualData)
		assert.IsType(t, test.err, err)

		err = os.RemoveAll(FixtureDir)
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
			err:          errFixtureNotFound{},
		},
		{
			name:         "example",
			actualData:   []byte("bc example"),
			expectedData: []byte("abc {{ .Name }}"),
			data:         nil,
			update:       true,
			err:          errFixtureMismatch{},
		},
	}

	for _, test := range tests {
		if test.update {
			err := Update(test.name, test.expectedData)
			assert.Nil(t, err)
		}

		err := compareTemplate(test.name, test.data, test.actualData)
		assert.IsType(t, test.err, err)

		err = os.RemoveAll(FixtureDir)
		assert.Nil(t, err)
	}
}
