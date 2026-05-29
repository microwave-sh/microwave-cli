package output

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// ── Semantic colors (shared across all Mataki CLIs) ──────────────────

var (
	ColorSuccess   = lipgloss.Color("#2dd4bf") // green — success, complete
	ColorError     = lipgloss.Color("#ef4444") // red — error, failed
	ColorWarning   = lipgloss.Color("#f59e0b") // yellow — warning, in-progress
	ColorPending   = lipgloss.Color("#64748b") // gray — pending, scheduled
	ColorCancelled = lipgloss.Color("#a78bfa") // magenta — cancelled
)

// ── Product brand colors (customize per CLI) ─────────────────────────

var (
	ColorPrimary = lipgloss.Color("#f66018") // product primary — headers, borders
	ColorAccent  = lipgloss.Color("#171717") // product accent — subtle borders
)

// ── Reusable styles ──────────────────────────────────────────────────

var (
	Green  = lipgloss.NewStyle().Foreground(ColorSuccess)
	Red    = lipgloss.NewStyle().Foreground(ColorError)
	Yellow = lipgloss.NewStyle().Foreground(ColorWarning)
	Bold   = lipgloss.NewStyle().Bold(true)
	Dim    = lipgloss.NewStyle().Faint(true)

	HeaderStyle = lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary)
	KeyStyle    = lipgloss.NewStyle().Bold(true).Foreground(ColorPending)
)

// ── Status color mapping ─────────────────────────────────────────────

var statusColors = map[string]lipgloss.Color{
	"pending":     ColorPending,
	"scheduled":   ColorPending,
	"running":     ColorWarning,
	"in_progress": ColorWarning,
	"completed":   ColorSuccess,
	"success":     ColorSuccess,
	"failed":      ColorError,
	"error":       ColorError,
	"cancelled":   ColorCancelled,
}

func ColorStatus(val string) string {
	c, ok := statusColors[strings.ToLower(val)]
	if !ok {
		return val
	}
	return lipgloss.NewStyle().Foreground(c).Render(val)
}

// ── Banners ──────────────────────────────────────────────────────────

func SuccessBanner(msg string) string {
	return lipgloss.NewStyle().
		BorderLeft(true).
		BorderStyle(lipgloss.ThickBorder()).
		BorderForeground(ColorSuccess).
		PaddingLeft(1).
		Render(msg)
}

func ErrorBanner(msg string) string {
	return lipgloss.NewStyle().
		BorderLeft(true).
		BorderStyle(lipgloss.ThickBorder()).
		BorderForeground(ColorError).
		PaddingLeft(1).
		Render(msg)
}
