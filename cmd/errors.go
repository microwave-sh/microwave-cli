package cmd

import "github.com/microwave-sh/microwave-cli/internal/output"

// RenderError wraps an error in a styled banner for TTY stderr display.
func RenderError(err error) string {
	return output.ErrorBanner("Error: " + err.Error())
}
