// Package ui provides terminal styling for devkit's output.
//
// Colors are enabled only when stdout is a terminal, TERM is not "dumb",
// and NO_COLOR is unset (https://no-color.org). Piped output stays plain.
package ui

import (
	"fmt"
	"os"
)

var enabled = detect()

func detect() bool {
	if os.Getenv("NO_COLOR") != "" || os.Getenv("TERM") == "dumb" {
		return false
	}
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

func wrap(code, s string) string {
	if !enabled {
		return s
	}
	return "\033[" + code + "m" + s + "\033[0m"
}

func Bold(s string) string    { return wrap("1", s) }
func Faint(s string) string   { return wrap("2", s) }
func Red(s string) string     { return wrap("31", s) }
func Green(s string) string   { return wrap("32", s) }
func Yellow(s string) string  { return wrap("33", s) }
func Cyan(s string) string    { return wrap("36", s) }
func Magenta(s string) string { return wrap("35", s) }

// Success formats a green "✓" success line.
func Success(format string, a ...any) string {
	return Green("✓ " + fmt.Sprintf(format, a...))
}

// Error formats a red "✗" error line.
func Error(format string, a ...any) string {
	return Red("✗ " + fmt.Sprintf(format, a...))
}

// LangColor colors a language label with a per-language color.
func LangColor(lang string) string {
	switch lang {
	case "Go":
		return Cyan(lang)
	case "TypeScript":
		return Yellow(lang)
	case "Python":
		return Green(lang)
	}
	return lang
}
