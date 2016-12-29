package goldie

import (
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
