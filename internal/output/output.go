package output

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ── Spinner (Bubbletea Tier 2) ───────────────────────────────────────

type doneMsg struct {
	result string
	failed bool
}

type spinnerModel struct {
	spinner spinner.Model
	msg     string
	done    bool
	result  string
	failed  bool
}

func (m spinnerModel) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m spinnerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC {
			m.done = true
			m.result = "Cancelled"
			m.failed = true
			return m, tea.Quit
		}
	case doneMsg:
		m.done = true
		m.result = msg.result
		m.failed = msg.failed
		return m, tea.Quit
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m spinnerModel) View() string {
	if m.done {
		if m.failed {
			return Red.Render("✗") + " " + m.result + "\n"
		}
		return Green.Render("✓") + " " + m.result + "\n"
	}
	return m.spinner.View() + " " + m.msg
}

type Spinner struct {
	program    *tea.Program
	wg         sync.WaitGroup
	cancelled  bool
	CancelledC chan struct{}
}

func NewSpinner(msg string) *Spinner {
	s := spinner.New(spinner.WithSpinner(spinner.Dot))
	s.Style = lipgloss.NewStyle().Foreground(ColorPrimary)

	m := spinnerModel{spinner: s, msg: msg}
	p := tea.NewProgram(m, tea.WithOutput(os.Stderr))
	sp := &Spinner{program: p, CancelledC: make(chan struct{})}

	sp.wg.Add(1)
	go func() {
		defer sp.wg.Done()
		finalModel, _ := p.Run()
		if m, ok := finalModel.(spinnerModel); ok && m.result == "Cancelled" {
			sp.cancelled = true
			close(sp.CancelledC)
		}
	}()

	return sp
}

func (s *Spinner) Stop(result string) {
	s.program.Send(doneMsg{result: result, failed: false})
	s.wg.Wait()
}

func (s *Spinner) Fail(result string) {
	s.program.Send(doneMsg{result: result, failed: true})
	s.wg.Wait()
}

func (s *Spinner) Cancelled() bool {
	s.wg.Wait()
	return s.cancelled
}

// ── Table Output with Format Switching ───────────────────────────────

// PrintTable renders data as a styled table (TTY) or JSON (scripting).
func PrintTable(headers []string, rows [][]string, jsonOutput bool) {
	if jsonOutput {
		printJSON(headers, rows)
		return
	}

	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = len(h)
	}
	for _, row := range rows {
		for i, cell := range row {
			if i < len(widths) && len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}

	cols := make([]table.Column, len(headers))
	for i, h := range headers {
		cols[i] = table.Column{Title: h, Width: widths[i]}
	}

	tableRows := make([]table.Row, len(rows))
	for i, r := range rows {
		tableRows[i] = table.Row(r)
	}

	s := table.DefaultStyles()
	s.Header = s.Header.Bold(true).BorderBottom(true)
	s.Selected = lipgloss.NewStyle()

	t := table.New(
		table.WithColumns(cols),
		table.WithRows(tableRows),
		table.WithHeight(len(rows)),
		table.WithStyles(s),
	)

	fmt.Println(t.View())
}

func printJSON(headers []string, rows [][]string) {
	result := make([]map[string]string, len(rows))
	for i, row := range rows {
		item := make(map[string]string)
		for j, h := range headers {
			if j < len(row) {
				item[h] = row[j]
			}
		}
		result[i] = item
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(result) //nolint:errcheck
}

// ── JSON output ───────────────────────────────────────────────────────

func PrintJSON(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

// ── Time formatting ───────────────────────────────────────────────────

func FormatTimeAgo(t time.Time) string {
	if t.IsZero() {
		return "—"
	}
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	default:
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	}
}
