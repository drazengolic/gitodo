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
	"github.com/spf13/cobra"
)

// doneCmd represents the done command
var doneCmd = &cobra.Command{
	Use:   "done",
	Short: "Set the first available to-do item to done",
	Long: `
Set the first available to-do item to done and output the next to-do item if
found.

If there are no items to be done, the "All done!" message is shown.

If there is a timer running, it will display the session time at the moment of
the command execution.`,
	Run: func(cmd *cobra.Command, args []string) {
		env, tdb := MustInit()
		projId := tdb.FetchProjectId(env.ProjDir, env.Branch)

		te, err := tdb.CheckTimer(projId)
		HandleTimerError(err)

		item := tdb.TodoWhat(projId)

		if item == nil {
			fmt.Println(greenTextStyle.Render("All done!"))
			return
		} else {
			tdb.TodoDone(item.Id, true)
			fmt.Printf("%s\n\n", strikeTroughStyle.Render(item.Task))
		}

		nextItem := tdb.TodoWhat(projId)

		if nextItem == nil {
			fmt.Println(greenTextStyle.Render("All done!"))
			return
		} else {
			fmt.Printf("%s\n%s\n", boldText.Render("Up next:"), nextItem.Task)
		}

		if te != nil && te.ProjectId == projId && te.Action == base.TimesheetActionStart {
			fmt.Printf("\n%s\n", orangeText.Render("Timer running for "+base.FormatSeconds(te.Duration())))
		}
	},
}

func init() {
	rootCmd.AddCommand(doneCmd)
}
