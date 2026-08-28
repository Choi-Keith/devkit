package cli

import (
	"fmt"

	"github.com/Choi-Keith/devkit/cmd/devkit/internal/templates"
	"github.com/Choi-Keith/devkit/cmd/devkit/internal/ui"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List available project templates",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()
		// Pad with plain text before coloring: ANSI escapes would break
		// tabwriter's width calculation.
		nameWidth, langWidth := 0, 0
		for _, t := range templates.All {
			nameWidth = max(nameWidth, len(t.Name))
			langWidth = max(langWidth, len(t.Lang))
		}
		fmt.Fprintf(out, "%s  %s  %s\n",
			ui.Faint(pad("NAME", nameWidth)), ui.Faint(pad("LANG", langWidth)), ui.Faint("DESCRIPTION"))
		for _, t := range templates.All {
			fmt.Fprintf(out, "%s  %s  %s\n",
				ui.Cyan(pad(t.Name, nameWidth)),
				ui.LangColor(t.Lang)+pad("", langWidth-len(t.Lang)),
				ui.Faint(t.Desc))
		}
		return nil
	},
}

// pad right-pads s with spaces to width (negative width returns s unchanged).
func pad(s string, width int) string {
	for len(s) < width {
		s += " "
	}
	return s
}

func init() {
	rootCmd.AddCommand(listCmd)
}
