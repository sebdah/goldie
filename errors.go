package goldie

// errFixtureNotFound is thrown when the fixture file could not be found.
type errFixtureNotFound struct {
	message string
}

// newErrFixtureNotFound returns a new instance of the error.
func newErrFixtureNotFound() errFixtureNotFound {
	return errFixtureNotFound{
		message: "Golden fixture not found. Try running with -update flag.",
	}
}

// Error returns the error message.
func (e errFixtureNotFound) Error() string {
	return e.message
}

// errFixtureMismatch is thrown when the actual and expected data is not
// matching.
type errFixtureMismatch struct {
	message string
}

// newErrFixtureMismatch returns a new instance of the error.
func newErrFixtureMismatch(message string) errFixtureMismatch {
	return errFixtureMismatch{
		message: message,
	}
}

func (e errFixtureMismatch) Error() string {
	return e.message
}
