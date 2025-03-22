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
	"strings"

	"github.com/drazengolic/gitodo/shell"
	"github.com/spf13/cobra"
)

// addCmd represents the add command
var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add to-do items for the current branch",
	Long: `
Add to-do items for the current branch.

Invoking without arguments will open up the editor for multiple items to be 
added. If there are arguments, all of them will be joined into a single to-do
item.

If the flag -t is provided, the new item will be placed at the top of the 
list.`,
	Run: func(cmd *cobra.Command, args []string) {
		env, tdb := MustInit()
		projId := tdb.FetchProjectId(env.ProjDir, env.Branch)

		_, err := tdb.CheckTimer(projId)
		HandleTimerError(err)

		if len(args) > 0 {
			item := strings.Join(args, " ")
			id, pos := tdb.AddTodo(projId, item)
			if cmd.Flags().Changed("top") {
				tdb.ChangePosition(id, pos, 1)
			}
			fmt.Printf("Added to-do item %q to %q\n", item, env.Branch)
		} else {
			tmpfile, err := shell.NewItemsTmpFile()
			ExitOnError(err, 1)
			err = tmpfile.Edit(env.Editor, 3)
			ExitOnError(err, 1)
			items, err := tmpfile.ReadItems()

			tmpfile.Delete()

			ExitOnError(err, 1)
			err = tdb.AddTodos(projId, items)
			ExitOnError(err, 1)

			fmt.Printf("Added %d item(s) to %q\n", len(items), env.Branch)
		}
	},
}

func init() {
	RootCmd.AddCommand(addCmd)
	addCmd.Flags().BoolP("top", "t", false, "put the item at the top of the list")
}
