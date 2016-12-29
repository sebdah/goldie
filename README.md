# goldie

[![GoDoc](https://godoc.org/github.com/sebdah/goldie?status.svg)](https://godoc.org/github.com/sebdah/goldie)

`goldie` is a golden file test utility for Go projects. It's typically used for testing responses with larger data bodies.

The concept is straight forward. Valid response data is stored in a "golden file". The actual response data will be byte compared with the golden file and the test will fail if there is a difference.

Updating the golden file can be done by running `go test -update ./...`.

## Example usage

The below example fetches data from a REST API. The last line in the test is the
actual usage of `goldie`. It takes the HTTP response body and asserts that it's
what is present in the golden test file.

```
func TestExample(t *testing.T) {
    recorder := httptest.NewRecorder()

    req, err := http.NewRequest("GET", "/example", nil)
    assert.Nil(t, err)

    handler := http.HandlerFunc(ExampleHandler)
    handler.ServeHTTP()

    goldie.Assert(t, "example", recorder.Body.Bytes())
}
```

## API

### Assert(*testing.T, string, []byte)

The `Assert` function takes the test object, a test name (should only contain letters valid in a filename) and a byte slice for comparing with the golden file.

### Update(string, []byte)

The `Update` function is used to write the actual data to the golden fixture.
This method is exposed so that it can be called from code. But the more common
approach would be to update the fixtures using the `-update` parameter to `go
test`.

## Configuration

The following public variables can be used to control the behaviour of `goldie`:

- `FixtureDir` - Dir to store the golden test files in. It will be relative to
	where the tests are executed. Default: `fixtures`.
- `FileNameSuffix` - Suffix to use for the golden test files. Default:
	`.golden`.
- `FlagName` - Name of the command line argument for updating the golden test
	files. Default: `update`.
- `FilePerms` - Permissions to set on the golden fixture files. Default: `0644`.
- `DirPerms` - Permissions to set on the golden fixture folder. Default: `0755`.

## FAQ

### Do you need any help in the project?

Yes, please! Pull requests are most welcome. On the wish list:

- Unit tests.
- Better output for failed tests. A diff of some sort would be great.

### Why the name `goldie`?

The name comes from the fact that it's for Go and handles golden file testing. But yes, it may not be the best name in the world.

## License

MIT

Copyright 2016 Sebastian Dahlgren <sebastian.dahlgren@gmail.com>

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
