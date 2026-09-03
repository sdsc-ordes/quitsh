// Package style renders the CLI's help/usage output with the same color
// palette the logger uses (see the `Color*` constants in `pkg/log`).
package style

import (
	"os"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/sdsc-ordes/quitsh/pkg/ci"
	"github.com/sdsc-ordes/quitsh/pkg/log"
)

// NOTE: We use our own renderer here (and not the global one `lipgloss.NewStyle`
// which the logger implicitly uses):
// the logger renders to `os.Stderr`, whereas cobra renders the help to
// `os.Stdout`. Different sinks -> different color profiles
// (e.g. `quitsh --help | less` must stay uncolored while the log stays colored).
//
// WARNING: Only use fixed colors (`lipgloss.Color`) below, never
// `lipgloss.AdaptiveColor`: adaptive colors make the renderer query the
// terminal's background color, which is the synchronous terminal operation
// `pkg/log`'s `init` guards with a file lock. We must not trigger that here
// again on another sink.
var renderer = newRenderer() //nolint:gochecknoglobals // Intended, same as the logger.

func newRenderer() *lipgloss.Renderer {
	r := lipgloss.NewRenderer(os.Stdout)

	if ci.IsRunning() && log.ForceColorInCI {
		// Same forcing as in `pkg/log`, otherwise CI logs have no colors.
		r.SetColorProfile(termenv.ANSI256)
	}

	return r
}

const ColorHeading = "#1166ab"
const ColorFlagName = "#00ae04"
const ColorCommand = "#00e6ff"
const ColorFlagType = "#bdbdbd"
const ColorMuted = "#00a4b6"
const ColorExample = "#bdbdbd"

type styles struct {
	heading     lipgloss.Style
	command     lipgloss.Style
	description lipgloss.Style
	example     lipgloss.Style
	flagName    lipgloss.Style
	flagType    lipgloss.Style
}

// Apply installs the colored help/usage templates on `rootCmd`.
// All subcommands inherit them: cobra looks the templates up the command tree
// (see `Command.getHelpTemplateFunc`), so this only needs to happen on the root.
func Apply(rootCmd *cobra.Command) {
	// Section headings, e.g. `Usage:`, `Flags:`.
	style := styles{
		heading: renderer.NewStyle().Foreground(lipgloss.Color(ColorHeading)).Bold(true),
		// Command names and command paths.

		command: renderer.NewStyle().Foreground(lipgloss.Color(ColorCommand)).Bold(true),
		// Descriptions.
		description: renderer.NewStyle().Foreground(lipgloss.Color(ColorMuted)),

		// Default values and Examples.
		example: renderer.NewStyle().Foreground(lipgloss.Color(ColorExample)).Italic(true),

		// Flag names, e.g. `-C, --cwd`.
		flagName: renderer.NewStyle().Foreground(lipgloss.Color(ColorFlagName)),

		// Flag value types, e.g. the `string` in `--cwd string`.
		flagType: renderer.NewStyle().Foreground(lipgloss.Color(ColorFlagType)).Italic(true),
	}

	// NOTE: Template funcs are global in cobra and the templates are parsed
	// lazily on first render, hence registering them here is enough.
	cobra.AddTemplateFuncs(map[string]any{
		"styleHeading": style.heading.Render,
		"styleCmd":     style.command.Render,
		"styleDesc":    style.description.Render,
		"styleExample": style.example.Render,

		"styleCmdPad": func(n string, p int) string { return styleCmdPad(n, p, &style) },
		"styleFlags":  func(f *pflag.FlagSet) string { return styleFlags(f, &style) },
	})

	rootCmd.SetHelpTemplate(helpTemplate)
	rootCmd.SetUsageTemplate(usageTemplate)
}

// styleCmdPad renders a command name padded to `padding` visible columns.
// NOTE: We cannot use cobra's `rpad` and style the result, because the padding
// would end up inside the ANSI escape sequence. Pad manually instead.
func styleCmdPad(name string, padding int, style *styles) string {
	pad := max(padding-utf8.RuneCountInString(name), 0)

	return style.command.Render(name) + strings.Repeat(" ", pad)
}

// Matches the alignment gap `pflag` inserts between the flag and its usage.
// The flag part itself never contains two consecutive spaces, the gap always
// does, which is what separates the two.
var reFlagGap = regexp.MustCompile(`\s{2,}`)

// Matches the flag part of a flag line, e.g. `-C, --cwd string`,
// `--parallel` or `-K, --config-val "a.b.c: {\"a\": 3}"`.
// Groups: the (optional shorthand and) name, the optional value type.
// NOTE: The value type may contain spaces (`pflag` takes it from a backquoted
// section of the usage string), hence it is matched greedily to the end.
var reFlagName = regexp.MustCompile(
	`^((?:-\w, )?--[^\s]+)(?:\s+(.+))?$`,
)

// Matches a trailing `(default "x")` which `pflag` appends to the usage.
var reFlagDefault = regexp.MustCompile(
	`\s*\(default .*\)$`,
)

// styleFlags renders a flag set the way `pflag` does, but colored.
//
// NOTE: We deliberately colorize `pflag`'s already formatted output instead of
// re-implementing `FlagUsages`: the alignment stays exactly `pflag`'s (ANSI
// sequences are zero-width) and we keep its handling of hidden/deprecated flags
// and of zero-valued defaults (`defaultIsZeroValue` is not exported).
func styleFlags(f *pflag.FlagSet, style *styles) string {
	// Same as cobra's default template which uses `FlagUsages` == wrapped(0).
	lines := strings.Split(f.FlagUsagesWrapped(0), "\n")

	for i, line := range lines {
		styled, ok := styleFlagLine(line, style)
		if !ok {
			// Not a flag line (e.g. a wrapped usage continuation): mute it.
			if strings.TrimSpace(line) != "" {
				styled = style.description.Render(line)
			} else {
				styled = line
			}
		}

		lines[i] = styled
	}

	return strings.Join(lines, "\n")
}

func styleFlagLine(line string, style *styles) (string, bool) {
	trimmed := strings.TrimLeft(line, " ")
	indent := line[:len(line)-len(trimmed)]

	// Split the line into the flag part and its usage at the alignment gap.
	// Only search behind the flag's name, a shorthand is followed by a
	// single space only (`-C, --cwd`).
	nameIdx := strings.Index(trimmed, "--")
	if nameIdx < 0 {
		return "", false
	}

	gap := reFlagGap.FindStringIndex(trimmed[nameIdx:])
	if gap == nil {
		return "", false
	}

	flg := trimmed[:nameIdx+gap[0]]
	spacing := trimmed[nameIdx+gap[0] : nameIdx+gap[1]]
	usage := trimmed[nameIdx+gap[1]:]

	m := reFlagName.FindStringSubmatch(flg)
	if m == nil {
		return "", false
	}

	styled := indent + style.flagName.Render(m[1])
	if m[2] != "" {
		styled += " " + style.flagType.Render(m[2])
	}

	// Split off a trailing `(default ...)` and render it separately.
	def := reFlagDefault.FindString(usage)
	usage = strings.TrimSuffix(usage, def)

	styled += spacing + style.description.Render(usage)
	if def != "" {
		styled += style.example.Render(def)
	}

	return styled, true
}
