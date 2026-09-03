package style

// The templates below are cobra's `defaultHelpTemplate`/`defaultUsageTemplate`
// (cobra v1.10.2, `command.go`) with the `style*` template functions threaded
// through. Keep them in sync when bumping cobra.

const helpTemplate = usageTemplate

//nolint:lll // Template whitespace is significant, lines cannot be wrapped.
const usageTemplate = `{{styleHeading "Usage:"}}{{if .Runnable}}
  {{styleCmd .UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{styleCmd .CommandPath}} {{styleCmd "[command]"}}{{end}}{{if gt (len .Aliases) 0}}

{{styleHeading "Aliases:"}}
  {{styleCmd .NameAndAliases}}{{end}}{{if .HasExample}}

{{styleHeading "Examples:"}}
{{styleExample .Example}}{{end}}{{if .HasAvailableSubCommands}}{{$cmds := .Commands}}{{if eq (len .Groups) 0}}

{{styleHeading "Available Commands:"}}{{range $cmds}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{styleCmdPad .Name .NamePadding}} {{styleDesc .Short}}{{end}}{{end}}{{else}}{{range $group := .Groups}}

{{styleHeading .Title}}{{range $cmds}}{{if (and (eq .GroupID $group.ID) (or .IsAvailableCommand (eq .Name "help")))}}
  {{styleCmdPad .Name .NamePadding}} {{styleDesc .Short}}{{end}}{{end}}{{end}}{{if not .AllChildCommandsHaveGroup}}

{{styleHeading "Additional Commands:"}}{{range $cmds}}{{if (and (eq .GroupID "") (or .IsAvailableCommand (eq .Name "help")))}}
  {{styleCmdPad .Name .NamePadding}} {{styleDesc .Short}}{{end}}{{end}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

{{styleHeading "Flags:"}}
{{styleFlags .LocalFlags | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

{{styleHeading "Global Flags:"}}
{{styleFlags .InheritedFlags | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

{{styleHeading "Additional help topics:"}}{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{styleCmdPad .CommandPath .CommandPathPadding}} {{styleDesc .Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{styleCmd .CommandPath}} {{styleCmd "[command]"}} {{styleCmd "--help"}}" for more information about a command.{{end}}
`
