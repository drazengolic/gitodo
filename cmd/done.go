/*
Copyright © 2024 Drazen Golic <drazen@fastmail.com>

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

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var strikeTroughStyle = lipgloss.NewStyle().Strikethrough(true)
var greenTextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#00ff00"))

// doneCmd represents the done command
var doneCmd = &cobra.Command{
	Use:   "done",
	Short: "Sets the first available to-do item to done",
	Long: `Sets the first available to-do item to done and outputs the next
to-do item if found.

If there are no items to be done, the "All done!" message is shown.`,
	Run: func(cmd *cobra.Command, args []string) {
		env, tdb := MustInit()
		projId := tdb.FetchProjectId(env.ProjDir, env.Branch)

		_, err := tdb.CheckTimer(projId)
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
			fmt.Printf("Up next: %s\n", nextItem.Task)
		}
	},
}

func init() {
	rootCmd.AddCommand(doneCmd)
}
