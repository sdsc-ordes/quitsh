package listcmd

import (
	"encoding/json"
	"io"
	"os"

	"github.com/sdsc-ordes/quitsh/pkg/cli"
	"github.com/sdsc-ordes/quitsh/pkg/cli/general"
	"github.com/sdsc-ordes/quitsh/pkg/component"
	"github.com/sdsc-ordes/quitsh/pkg/errors"
	"github.com/sdsc-ordes/quitsh/pkg/log"

	"github.com/spf13/cobra"
)

type listArgs struct {
	compArgs general.ComponentArgs
	jsonFile string
}

const longDesc = `
List all components found in the current working directory.
`

func AddCmd(cl cli.ICLI, parent *cobra.Command) {
	var args listArgs

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List components",
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
		StringVar(&args.jsonFile,
			"json", "", "Output the found components in JSON format to this file.")

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

	if c.jsonFile != "" {
		var writer io.WriteCloser

		if c.jsonFile == "-" {
			writer = os.Stdout
		} else {
			writer, err = os.Create(c.jsonFile)
			if err != nil {
				return err
			}
			defer writer.Close()
		}

		err := outputJSON(comps, writer)
		if err != nil {
			return errors.AddContext(err, "Could not marshal output to JSON.")
		}
	}

	return nil
}

func outputJSON(comps []*component.Component, w io.Writer) error {
	type D struct {
		Root     string
		OutDir   string
		Name     string
		Language string
	}

	var configs []D
	for i := range comps {
		c := comps[i]
		configs = append(
			configs,
			D{Root: c.Root(), OutDir: c.OutDir(), Name: c.Name(), Language: c.Language()},
		)
	}

	js, err := json.Marshal(configs)
	if err != nil {
		return err
	}

	_, err = w.Write(js)

	return err
}
