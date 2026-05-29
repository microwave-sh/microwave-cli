package cmd

import "github.com/microwave-sh/microwave-cli/internal/output"

// RenderError wraps an error in a styled banner for TTY stderr display.
func RenderError(err error) string {
	return output.ErrorBanner("Error: " + err.Error())
}

// notImplementedError is a placeholder used by resource skeletons until Tasks 7–11.
type notImplementedError struct{ what string }

func (e notImplementedError) Error() string { return e.what + ": not implemented yet" }
