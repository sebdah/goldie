package goldie

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

var (
	// BasePath is the folder name for where the fixtures are stored. It's
	// relative to the "go test" path.
	BasePath = "golden-fixtures"

	// FileNameSuffix is the suffix appended to the fixtures. Set to empty
	// string to disable file name suffixes.
	FileNameSuffix = ".golden"

	// FlagName is the name of the command line flag for go test.
	FlagName = "update"

	// update determines if the actual received data should be written to the
	// golden files or not. This should be true when you need to update the
	// data, but false when actually running the tests.
	update = flag.Bool(FlagName, false, "Update golden test file fixture")
)

// Assert compares the actual data received with the expected data in the
// golden files. If the update flag is set, it will also update the golden
// file.
func Assert(t *testing.T, name string, actualData []byte) {
	goldenFilePath := filepath.Join(
		BasePath,
		fmt.Sprintf("%s%s", name, FileNameSuffix))

	if *update {
		err := ensureBasePath()
		if err != nil {
			t.Error(err)
			t.FailNow()
		}

		err := ioutil.WriteFile(goldenFilePath, actualData, 0644)
		if err != nil {
			t.Error(err)
			t.FailNow()
		}
	}

	expectedData, err := ioutil.ReadFile(goldenFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			t.Error("Golden fixture not found. Try running with -update flag.")
			t.FailNow()
		} else {
			t.Errorf("Expected %s to be nil", err.Error())
		}
	}

	if !bytes.Equal(actualData, expectedData) {
		t.Errorf("Result did not match the golden file")
	}
}

// ensureBasePath will create the fixture folder if it does not already exist.
func ensureBasePath() error {
	_, err := os.Stat(BasePath)
	if err == nil {
		return nil
	}

	if os.IsNotExist(err) {
		err = os.Mkdir(BasePath, 0755)
		if err != nil {
			return err
		}
	}

	return err
}
