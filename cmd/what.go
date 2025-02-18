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

	"github.com/drazengolic/gitodo/base"
	"github.com/drazengolic/gitodo/shell"
	"github.com/spf13/cobra"
)

// whatCmd represents the what command
var whatCmd = &cobra.Command{
	Use:     "what",
	Aliases: []string{"status"},
	Short:   "Display what's next to do",
	Long: `
Display the first to-do item that isn't completed yet, starting from the top
of the list.

If there is no such item, "All done!" message will be shown.

If there is a timer running, it will display the session time at the moment of
the command execution. Also, if there are stashed changes assigned to any of
the to-do items in the repository, the full list will be printed, organized by
the branch name.`,
	Run: func(cmd *cobra.Command, args []string) {
		env, tdb := MustInit()
		proj := tdb.GetProject(tdb.FetchProjectId(env.ProjDir, env.Branch))

		te, err := tdb.CheckTimer(proj.Id)
		HandleTimerError(err)

		fmt.Printf("%s\n", blueText.Render(proj.Name))

		item := tdb.TodoWhat(proj.Id)

		if item == nil {
			fmt.Println(greenTextStyle.Render("All done!"))
		} else {
			fmt.Printf("%s\n\n%s\n\n", boldText.Render("To do:"), item.Task)
		}

		if te != nil && te.ProjectId == proj.Id && te.Action == base.TimesheetActionStart {
			fmt.Printf("%s\n", orangeText.Render("Timer running for "+base.FormatSeconds(te.Duration())))
		}

		stash, err := shell.GetStashItems()
		ExitOnError(err, 1)

		if len(stash) == 0 {
			return
		}

		ids := make([]int, 0, len(stash))
		for k := range stash {
			ids = append(ids, k)
		}

		ibs, err := tdb.GetItemsAndBranch(ids)
		ExitOnError(err, 1)

		fmt.Printf("\n%s\n\n", dimmedText.Render("Items with stash:"))
		currBranch := ""
		for _, ib := range ibs {
			if ib.BranchName != currBranch {
				fmt.Println(dimmedText.Render(ib.BranchName))
				currBranch = ib.BranchName
			}
			fmt.Printf("  %s\n", dimmedText.Render("- "+ib.ItemName))
		}
	},
}

func init() {
	rootCmd.AddCommand(whatCmd)
}
