package cmd

// newTestGlobals returns a Globals wired to the given test server URL.
func newTestGlobals(serverURL string) *Globals {
	return &Globals{
		Token:  "test-token",
		APIURL: serverURL,
	}
}
