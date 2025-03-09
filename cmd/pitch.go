/*
Copyright © 2025 Dražen Golić

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cmd

import (
	"fmt"
	"os"

	"github.com/drazengolic/gitodo/shell"
	"github.com/spf13/cobra"
)

// pitchCmd represents the pitch command
var pitchCmd = &cobra.Command{
	Use:     "pitch branch_name [args...]",
	Aliases: []string{"p"},
	Short:   "Checkout branch and add items at one go",
	Long: `
Quickly checkout to a branch (or create a new one if it doesn't exist) and add
to-do items via arguments, or with an editor if no arguments are provided.

If --base flag is not provided, the current branch will be used as a starting
point for the new branch.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println("Branch name not provided in the arguments.")
			os.Exit(1)
		}

		env, tdb := MustInit()

		// checkout the (new) branch
		envBranches, err := shell.ListBranches()
		ExitOnError(err, 1)
		createBranch := true
		for _, b := range envBranches {
			if b == args[0] {
				createBranch = false
				break
			}
		}
		base, _ := cmd.Flags().GetString("base")
		err = shell.CheckoutBranch(args[0], base, createBranch)
		ExitOnError(err, 1)

		env.Branch = args[0]
		projId := tdb.FetchProjectId(env.ProjDir, env.Branch)

		// update name if provided
		if cmd.Flags().Changed("name") {
			name, _ := cmd.Flags().GetString("name")
			if name == "" {
				name = env.Branch
			}
			tdb.UpdateProjectName(projId, name)
		}

		itemCount := 0

		// add items
		if len(args) > 1 {
			tdb.AddTodos(projId, args[1:])
			itemCount = len(args) - 1
		} else {
			tmpfile, err := shell.NewTmpFileString(`# Start a line with a hyphen (-) to indicate a new item.
# Comments like this are ignored.
- `)
			ExitOnError(err, 1)
			err = tmpfile.Edit(env.Editor, 3)
			ExitOnError(err, 1)
			items, err := tmpfile.ReadItems()

			tmpfile.Delete()

			ExitOnError(err, 1)
			tdb.AddTodos(projId, items)
			itemCount = len(items)
		}

		// print summary or git status
		if itemCount > 0 {
			fmt.Printf("Added %d new to-do item(s) for %q.\n", itemCount, env.Branch)
		} else {
			shell.GitStatus()
		}
	},
}

func init() {
	rootCmd.AddCommand(pitchCmd)

	pitchCmd.Flags().StringP("base", "b", "", "Starting point (base) for the new branch")
	pitchCmd.Flags().StringP("name", "n", "", "Project name")
}
