package cli_test

import (
	"strings"
	"testing"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/cli"
)

func TestPlainMode_NoANSI(t *testing.T) {
	opts := cli.PlainOptions()

	results := []string{
		opts.Green("hello"),
		opts.Red("error"),
		opts.Yellow("warn"),
		opts.Cyan("heading"),
		opts.Blue("info"),
		opts.Grey("detail"),
		opts.Bold("important"),
	}

	for _, s := range results {
		if strings.Contains(s, "\033[") {
			t.Errorf("plain mode should not contain ANSI codes, got: %q", s)
		}
	}
}

func TestPlainMode_NoEmoji(t *testing.T) {
	opts := cli.PlainOptions()

	icon := opts.Icon(cli.EmojiValid)
	if icon != "" {
		t.Errorf("plain mode icon should be empty, got: %q", icon)
	}
}

func TestColourMode_HasANSI(t *testing.T) {
	opts := cli.DefaultOptions()

	green := opts.Green("test")
	if !strings.Contains(green, "\033[32m") {
		t.Errorf("expected green ANSI code, got: %q", green)
	}
	if !strings.Contains(green, "\033[0m") {
		t.Errorf("expected reset ANSI code, got: %q", green)
	}
}

func TestEmojiMode_HasEmoji(t *testing.T) {
	opts := cli.DefaultOptions()

	icon := opts.Icon(cli.EmojiValid)
	if !strings.Contains(icon, "✅") {
		t.Errorf("expected emoji, got: %q", icon)
	}
}

func TestFormatInt_ThousandsSeparators(t *testing.T) {
	tests := []struct {
		input    int
		expected string
	}{
		{0, "0"},
		{42, "42"},
		{999, "999"},
		{1000, "1,000"},
		{12345, "12,345"},
		{1000000, "1,000,000"},
		{1234567890, "1,234,567,890"},
		{-500, "-500"},
		{-1234, "-1,234"},
	}

	for _, tt := range tests {
		got := cli.FormatInt(tt.input)
		if got != tt.expected {
			t.Errorf("FormatInt(%d) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestFormatMs_HumanReadable(t *testing.T) {
	tests := []struct {
		input    int64
		expected string
	}{
		{0, "0ms"},
		{150, "150ms"},
		{999, "999ms"},
		{1000, "1s"},
		{1500, "1.5s"},
		{5000, "5s"},
		{60000, "1m 0s"},
		{65000, "1m 5s"},
		{125000, "2m 5s"},
	}

	for _, tt := range tests {
		got := cli.FormatMs(tt.input)
		if got != tt.expected {
			t.Errorf("FormatMs(%d) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestTable_DeterministicOutput(t *testing.T) {
	opts := cli.PlainOptions()
	cols := []cli.Column{
		{Name: "Rank", Width: 6},
		{Name: "Penalty", Width: 10, Right: true},
		{Name: "Runtime", Width: 10, Right: true},
	}

	tbl := cli.NewTable(cols, opts)

	header := tbl.Header()
	row := tbl.Row([]string{"1", "2,450", "150ms"})

	// Run twice — must be identical.
	header2 := tbl.Header()
	row2 := tbl.Row([]string{"1", "2,450", "150ms"})

	if header != header2 {
		t.Error("table header not deterministic")
	}
	if row != row2 {
		t.Error("table row not deterministic")
	}

	// No ANSI in plain mode.
	if strings.Contains(row, "\033[") {
		t.Error("plain mode row should not contain ANSI")
	}
}

func TestTable_ColumnAlignment(t *testing.T) {
	opts := cli.PlainOptions()
	cols := []cli.Column{
		{Name: "Name", Width: 10},
		{Name: "Value", Width: 8, Right: true},
	}

	tbl := cli.NewTable(cols, opts)
	row := tbl.Row([]string{"hello", "42"})

	// "hello" left-padded to 10, "42" right-padded to 8.
	if !strings.HasPrefix(row, "hello     ") {
		t.Errorf("expected left-aligned 'hello', got: %q", row)
	}
	if !strings.HasSuffix(row, "      42") {
		t.Errorf("expected right-aligned '42', got: %q", row)
	}
}

func TestHeading_PlainMode(t *testing.T) {
	opts := cli.PlainOptions()
	heading := opts.Heading(cli.EmojiSummary, "Results")

	if strings.Contains(heading, "\033[") {
		t.Error("plain heading should not contain ANSI")
	}
	if strings.Contains(heading, "📊") {
		t.Error("plain heading should not contain emoji")
	}
	if !strings.Contains(heading, "Results") {
		t.Error("heading should contain text")
	}
}
