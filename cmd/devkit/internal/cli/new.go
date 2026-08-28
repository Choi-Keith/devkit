package cli

import (
	"bufio"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Choi-Keith/devkit/cmd/devkit/internal/fetch"
	"github.com/Choi-Keith/devkit/cmd/devkit/internal/generator"
	"github.com/Choi-Keith/devkit/cmd/devkit/internal/templates"
	"github.com/Choi-Keith/devkit/cmd/devkit/internal/ui"
	"github.com/spf13/cobra"
)

var (
	newDir    string
	newModule string
	newForce  bool
	newRepo   string
	newRef    string
	newLocal  string
)

var newCmd = &cobra.Command{
	Use:   "new [template] [project-name]",
	Short: "Create a new project from a template",
	Long: `Create a new project from a template.

The template is downloaded from the template repository at generation time.
Run without arguments to pick a template interactively:
  devkit new

Or non-interactively:
  devkit new golang-ddd my-service --module github.com/acme/my-service

Use --local to generate from a local template directory instead (no network).`,
	Args: cobra.MaximumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			tmpl templates.Template
			name string
			err  error
		)
		// A single buffered reader is shared by all prompts: creating a new
		// bufio.Reader per prompt would discard input it already buffered.
		reader := bufio.NewReader(cmd.InOrStdin())

		if len(args) > 0 {
			tmpl, err = templates.Find(args[0])
			if err != nil {
				return err
			}
		} else {
			tmpl, err = promptTemplate(cmd, reader)
			if err != nil {
				return err
			}
		}

		if len(args) > 1 {
			name = args[1]
			if err := validateProjectName(name); err != nil {
				return err
			}
		} else {
			name, err = promptProjectName(cmd, reader)
			if err != nil {
				return err
			}
		}

		srcDir, cleanup, err := resolveSource(cmd, tmpl)
		if err != nil {
			return err
		}
		defer cleanup()

		destDir := filepath.Join(newDir, name)
		if err := generator.Generate(srcDir, destDir, tmpl.Lang == "Go", generator.Options{
			Module: newModule,
			Force:  newForce,
		}); err != nil {
			return err
		}

		out := cmd.OutOrStdout()
		fmt.Fprintln(out, "\n"+ui.Success("Project %s created at %s", ui.Bold(name), destDir))
		fmt.Fprintln(out, "\n"+ui.Bold("Next steps:"))
		fmt.Fprintf(out, "  %s\n%s", ui.Yellow("cd "+destDir), nextSteps(tmpl))
		return nil
	},
}

func init() {
	newCmd.Flags().StringVarP(&newDir, "dir", "d", ".", "parent directory where the project is created")
	newCmd.Flags().StringVar(&newModule, "module", "", "rewrite the Go module path (Go templates only)")
	newCmd.Flags().BoolVar(&newForce, "force", false, "write into a non-empty directory")
	newCmd.Flags().StringVar(&newRepo, "repo", templates.DefaultRepo, "git repository to download templates from")
	newCmd.Flags().StringVar(&newRef, "ref", "", "git branch or tag to download templates from (default: default branch)")
	newCmd.Flags().StringVar(&newLocal, "local", "", "use a local template directory instead of downloading")
	rootCmd.AddCommand(newCmd)
}

// resolveSource returns the template source directory, either the --local
// path as-is or a freshly cloned subdirectory of the template repository.
func resolveSource(cmd *cobra.Command, tmpl templates.Template) (dir string, cleanup func(), err error) {
	if newLocal != "" {
		return newLocal, func() {}, nil
	}
	fmt.Fprintf(cmd.OutOrStdout(), "%s\n", ui.Faint(fmt.Sprintf("Downloading template %q from %s ...", tmpl.Name, newRepo)))
	return fetch.Template(newRepo, newRef, tmpl.Subdir)
}

func promptTemplate(cmd *cobra.Command, reader *bufio.Reader) (templates.Template, error) {
	out := cmd.OutOrStdout()
	fmt.Fprintln(out, ui.Bold("Select a template:"))
	nameWidth := 0
	for _, t := range templates.All {
		nameWidth = max(nameWidth, len(t.Name))
	}
	for i, t := range templates.All {
		fmt.Fprintf(out, "  %s %s  %s\n",
			ui.Yellow(fmt.Sprintf("%d)", i+1)), ui.Cyan(pad(t.Name, nameWidth)), ui.Faint(t.Desc))
	}
	for {
		answer, err := promptString(cmd, reader, "Enter a number")
		if err != nil {
			return templates.Template{}, err
		}
		if n, err := strconv.Atoi(answer); err == nil && n >= 1 && n <= len(templates.All) {
			return templates.All[n-1], nil
		}
		// Also accept the template name directly.
		if t, err := templates.Find(answer); err == nil {
			return t, nil
		}
		fmt.Fprintf(out, "%s\n", ui.Red(fmt.Sprintf("Invalid choice %q, please enter 1-%d.", answer, len(templates.All))))
	}
}

// promptProjectName asks for a project name, re-prompting until the input
// is non-empty and valid.
func promptProjectName(cmd *cobra.Command, reader *bufio.Reader) (string, error) {
	out := cmd.OutOrStdout()
	for {
		name, err := promptString(cmd, reader, "Project name")
		if err != nil {
			return "", err
		}
		if name == "" {
			fmt.Fprintf(out, "%s\n", ui.Red("Project name must not be empty."))
			continue
		}
		if err := validateProjectName(name); err == nil {
			return name, nil
		} else {
			fmt.Fprintf(out, "%s\n", ui.Red(err.Error()))
		}
	}
}

func promptString(cmd *cobra.Command, reader *bufio.Reader, label string) (string, error) {
	fmt.Fprintf(cmd.OutOrStdout(), "%s: ", ui.Magenta(label))
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(line), nil
}

func validateProjectName(name string) error {
	if strings.ContainsAny(name, `/\`) || name == "." || name == ".." {
		return fmt.Errorf("invalid project name %q", name)
	}
	return nil
}

func nextSteps(t templates.Template) string {
	switch t.Lang {
	case "Go":
		return "  go mod tidy\n  go run ./cmd/...\n"
	case "TypeScript":
		return "  pnpm install   # or npm install\n  pnpm start:dev\n"
	case "Python":
		return "  uv sync\n  uv run python -m src.main  # see README.md\n"
	}
	return ""
}
