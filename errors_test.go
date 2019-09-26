package goldie

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrFixtureNotFound(t *testing.T) {
	expected := "Golden fixture not found. Try running with -update flag."
	err := newErrFixtureNotFound()

	assert.Equal(t, expected, err.Error())
	assert.IsType(t, &errFixtureNotFound{}, err)
}

func TestErrFixtureMismatch(t *testing.T) {
	message := "example message"
	err := newErrFixtureMismatch(message)

	assert.Equal(t, message, err.Error())
	assert.IsType(t, &errFixtureMismatch{}, err)
}

func TestErrFixtureDirectoryIsFile(t *testing.T) {
	message := "fixture folder is a file: some/location/thing"
	location := "some/location/thing"
	err := newErrFixtureDirectoryIsFile(location)

	assert.Equal(t, message, err.Error())
	assert.IsType(t, &errFixtureDirectoryIsFile{}, err)
}
