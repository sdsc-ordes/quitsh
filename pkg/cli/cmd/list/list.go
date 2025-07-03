package listcmd

import (
	"fmt"
	"io"
	"os"
	"text/template"

	"github.com/Masterminds/sprig"
	"github.com/sdsc-ordes/quitsh/pkg/cli"
	"github.com/sdsc-ordes/quitsh/pkg/cli/general"
	"github.com/sdsc-ordes/quitsh/pkg/component"
	"github.com/sdsc-ordes/quitsh/pkg/errors"
	"github.com/sdsc-ordes/quitsh/pkg/log"

	"github.com/spf13/cobra"
)

const longDesc = `
List all components found in the current working directory.
`

const defaultOutputFormat = "{{ . | toJson }}"

type listArgs struct {
	compArgs   general.ComponentArgs
	outputFile string
	format     string
}

func AddCmd(cl cli.ICLI, parent *cobra.Command) {
	var args listArgs

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List components.",
		Long:  longDesc,
		PreRunE: func(_cmd *cobra.Command, _args []string) error {
			return nil
		},
		RunE: func(_cmd *cobra.Command, _args []string) error {
			return listComponents(cl, &args)
		},
	}

	listCmd.Flags().
		StringArrayVarP(&args.compArgs.ComponentPatterns,
			"components", "c", []string{"*"}, "Components matched by these patterns are listed.")

	listCmd.Flags().
		StringVar(&args.outputFile,
			"output", "", "Output the found components to this file (if `-` = `stdout`, see `format`).")

	listCmd.Flags().
		StringVar(&args.format,
			"format", "",
			fmt.Sprintf("Template format (Go) string to use for output (defaults to '%s'.", defaultOutputFormat),
		)

	parent.AddCommand(listCmd)
}

func listComponents(cl cli.ICLI, c *listArgs) error {
	comps, _, _, err := cl.FindComponents(&c.compArgs)

	if err != nil {
		return err
	}

	for i := range comps {
		log.Info(
			"Component:",
			"root",
			comps[i].Root(),
			"name",
			comps[i].Config().Name,
			"version",
			comps[i].Config().Version.String(),
		)
	}

	if c.outputFile != "" || c.format != "" {
		err := outputToFile(comps, c.outputFile, c.format) //nolint:govet //intentional

		if err != nil {
			return errors.AddContext(err, "Could not marshal output to JSON.")
		}
	}

	return nil
}

func outputToFile(comps []*component.Component, outputFile, format string) error {
	var w io.WriteCloser

	if outputFile == "-" || outputFile == "" {
		w = os.Stdout
	} else {
		writer, err := os.Create(outputFile)
		if err != nil {
			return err
		}
		defer writer.Close()
	}

	type D struct {
		Root     string `json:"root"`
		OutDir   string `json:"outDir"`
		Name     string `json:"name"`
		Language string `json:"language"`
	}

	if format == "" {
		format = defaultOutputFormat
	}

	tmpl, err := template.New("output").Funcs(sprig.FuncMap()).Parse(format)
	if err != nil {
		return errors.AddContext(err, "failed to parse template")
	}

	var configs []D
	for i := range comps {
		c := comps[i]
		configs = append(
			configs,
			D{Root: c.Root(), OutDir: c.OutDir(), Name: c.Name(), Language: c.Language()},
		)
	}

	if err = tmpl.Execute(w, configs); err != nil {
		return errors.AddContext(err, "failed to execute template")
	}

	return err
}
