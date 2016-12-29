package goldie

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
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

func TestEnsureFixtureDir(t *testing.T) {
	tests := []struct {
		dir         string
		shouldExist bool
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
			dir:         "\"24348q0980fd/&&**D&S**SS:",
			shouldExist: false,
			err:         &os.PathError{},
		},
	}

	for _, test := range tests {
		oldFixtureDir := FixtureDir
		FixtureDir = test.dir

		if test.shouldExist {
			err := os.Mkdir(test.dir, 0755)
			assert.Nil(t, err)
		}

		err := ensureFixtureDir()
		assert.IsType(t, test.err, err)

		if err == nil {
			err = os.RemoveAll(test.dir)
			assert.Nil(t, err)
		}

		FixtureDir = oldFixtureDir
	}
}
