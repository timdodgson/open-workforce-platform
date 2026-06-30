// Package cli provides terminal presentation helpers for the OWP CLI.
// No business logic. Only formatting, colour, and table rendering.
package cli

import (
	"fmt"
	"strings"
)

// --- Display Options ---

// Options controls colour, emoji and formatting behaviour.
type Options struct {
	Colour bool
	Emoji  bool
}

// DefaultOptions returns options with colour and emoji enabled.
func DefaultOptions() Options {
	return Options{Colour: true, Emoji: true}
}

// PlainOptions returns options with colour and emoji disabled (for CI/logs).
func PlainOptions() Options {
	return Options{Colour: false, Emoji: false}
}

// --- ANSI Colours ---

const (
	ansiReset  = "\033[0m"
	ansiRed    = "\033[31m"
	ansiGreen  = "\033[32m"
	ansiYellow = "\033[33m"
	ansiBlue   = "\033[34m"
	ansiCyan   = "\033[36m"
	ansiGrey   = "\033[90m"
	ansiBold   = "\033[1m"
)

// Green wraps text in green ANSI if colour is enabled.
func (o Options) Green(s string) string {
	if !o.Colour {
		return s
	}
	return ansiGreen + s + ansiReset
}

// Red wraps text in red ANSI if colour is enabled.
func (o Options) Red(s string) string {
	if !o.Colour {
		return s
	}
	return ansiRed + s + ansiReset
}

// Yellow wraps text in yellow ANSI if colour is enabled.
func (o Options) Yellow(s string) string {
	if !o.Colour {
		return s
	}
	return ansiYellow + s + ansiReset
}

// Cyan wraps text in cyan ANSI if colour is enabled.
func (o Options) Cyan(s string) string {
	if !o.Colour {
		return s
	}
	return ansiCyan + s + ansiReset
}

// Blue wraps text in blue ANSI if colour is enabled.
func (o Options) Blue(s string) string {
	if !o.Colour {
		return s
	}
	return ansiBlue + s + ansiReset
}

// Grey wraps text in grey ANSI if colour is enabled.
func (o Options) Grey(s string) string {
	if !o.Colour {
		return s
	}
	return ansiGrey + s + ansiReset
}

// Bold wraps text in bold ANSI if colour is enabled.
func (o Options) Bold(s string) string {
	if !o.Colour {
		return s
	}
	return ansiBold + s + ansiReset
}

// --- Emoji ---

const (
	EmojiValid   = "✅"
	EmojiInvalid = "❌"
	EmojiBest    = "🏆"
	EmojiRunning = "🔁"
	EmojiConfig  = "⚙️"
	EmojiSummary = "📊"
	EmojiTime    = "⏱️"
	EmojiSeed    = "🧪"
)

// Icon returns the emoji if enabled, empty string otherwise.
func (o Options) Icon(emoji string) string {
	if !o.Emoji {
		return ""
	}
	return emoji + " "
}

// --- Number Formatting ---

// FormatInt formats an integer with thousands separators (comma).
func FormatInt(n int) string {
	if n < 0 {
		return "-" + FormatInt(-n)
	}
	s := fmt.Sprintf("%d", n)
	if len(s) <= 3 {
		return s
	}

	var result strings.Builder
	remainder := len(s) % 3
	if remainder > 0 {
		result.WriteString(s[:remainder])
	}
	for i := remainder; i < len(s); i += 3 {
		if result.Len() > 0 {
			result.WriteByte(',')
		}
		result.WriteString(s[i : i+3])
	}
	return result.String()
}

// FormatMs formats milliseconds as a human-readable duration.
// e.g. 1500 -> "1.5s", 65000 -> "1m 5s", 150 -> "150ms"
func FormatMs(ms int64) string {
	if ms < 1000 {
		return fmt.Sprintf("%dms", ms)
	}
	if ms < 60000 {
		secs := float64(ms) / 1000
		if ms%1000 == 0 {
			return fmt.Sprintf("%ds", ms/1000)
		}
		return fmt.Sprintf("%.1fs", secs)
	}
	mins := ms / 60000
	secs := (ms % 60000) / 1000
	return fmt.Sprintf("%dm %ds", mins, secs)
}

// --- Table Rendering ---

// Column defines a table column with name and width.
type Column struct {
	Name  string
	Width int
	Right bool // right-align
}

// Table renders aligned tabular data.
type Table struct {
	Columns []Column
	opts    Options
}

// NewTable creates a table with the given columns and display options.
func NewTable(cols []Column, opts Options) *Table {
	return &Table{Columns: cols, opts: opts}
}

// Header returns the formatted header line.
func (t *Table) Header() string {
	var parts []string
	for _, col := range t.Columns {
		if col.Right {
			parts = append(parts, padLeft(col.Name, col.Width))
		} else {
			parts = append(parts, padRight(col.Name, col.Width))
		}
	}
	return t.opts.Cyan(strings.Join(parts, " "))
}

// Separator returns a horizontal separator line matching the table width.
func (t *Table) Separator() string {
	total := 0
	for _, col := range t.Columns {
		total += col.Width
	}
	total += len(t.Columns) - 1 // spaces between columns
	return t.opts.Grey(strings.Repeat("─", total))
}

// Row formats a row of values. Values are converted to strings.
func (t *Table) Row(values []string) string {
	var parts []string
	for i, col := range t.Columns {
		val := ""
		if i < len(values) {
			val = values[i]
		}
		if col.Right {
			parts = append(parts, padLeft(val, col.Width))
		} else {
			parts = append(parts, padRight(val, col.Width))
		}
	}
	return strings.Join(parts, " ")
}

// HighlightRow formats a row and applies green highlight (for best result).
func (t *Table) HighlightRow(values []string) string {
	return t.opts.Green(t.Row(values))
}

// ErrorRow formats a row and applies red highlight (for invalid result).
func (t *Table) ErrorRow(values []string) string {
	return t.opts.Red(t.Row(values))
}

// --- Padding Helpers ---

func padRight(s string, width int) string {
	if len(s) >= width {
		return s[:width]
	}
	return s + strings.Repeat(" ", width-len(s))
}

func padLeft(s string, width int) string {
	if len(s) >= width {
		return s[:width]
	}
	return strings.Repeat(" ", width-len(s)) + s
}

// --- Section Headers ---

// Heading prints a section heading with optional emoji.
func (o Options) Heading(emoji, text string) string {
	prefix := o.Icon(emoji)
	return o.Bold(o.Cyan(prefix + text))
}

// Warning prints a warning message.
func (o Options) Warning(text string) string {
	prefix := o.Icon(EmojiInvalid)
	return o.Red(prefix + text)
}

// Success prints a success message.
func (o Options) Success(text string) string {
	prefix := o.Icon(EmojiValid)
	return o.Green(prefix + text)
}
