module github.com/microwave-sh/microwave-cli

go 1.26.1

replace github.com/microwave-sh/microwave-go => ../microwave-go

require (
	github.com/alecthomas/kong v1.14.0
	github.com/charmbracelet/lipgloss v1.1.0
	github.com/microwave-sh/microwave-go v0.0.0-00010101000000-000000000000
	github.com/muesli/termenv v0.16.0
)
