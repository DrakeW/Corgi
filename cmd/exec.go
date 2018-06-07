package cmd

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var execCmd = &cobra.Command{
	Use:   "exec",
	Short: "Execute a snippet",
	Args:  cobra.MaximumNArgs(1),
	RunE:  execute,
}

var useDefaultParamValue bool
var stepRange string
var execTitle string

func execute(cmd *cobra.Command, args []string) error {
	// load config & snippets
	conf, snippets, err := loadConfigAndSnippetsMeta()
	if err != nil {
		return err
	}
	// find snippet title
	var title string
	if len(args) == 0 {
		if conf.FilterCmd != "" {
			title, err = filter(conf.FilterCmd, snippets.GetSnippetTitles())
			if err != nil || title == "" {
				return MissingSnippetTitleError
			}
		} else {
			color.Red("Install a fuzzy finder (\"fzf\" or \"peco\") to enable interactive selection")
			return MissingSnippetTitleError
		}
	} else {
		title = args[0]
	}
	// find snippet corresponds to title
	s, err := snippets.FindSnippet(title)
	if err != nil {
		return fmt.Errorf("%s, run \"corgi list\" to view all snippets", err.Error())
	}
	s.Execute(useDefaultParamValue, stepRange)
	return nil
}

func init() {
	execCmd.Flags().StringVarP(&stepRange, "step", "s", "", "Select a single step to execute with \"-s <step>\" or a range of steps to execute with \"-s <start>-<end>\", end is optional")
	execCmd.Flags().BoolVar(&useDefaultParamValue, "use-default", false, "Add this flag if you would like to use the default values for your defined template fields without being asked to enter a value")
	rootCmd.AddCommand(execCmd)
}
