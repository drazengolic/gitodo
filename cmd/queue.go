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
	"strings"

	"github.com/drazengolic/gitodo/shell"
	"github.com/spf13/cobra"
)

// addCmd represents the add command
var queueCmd = &cobra.Command{
	Use:   "queue",
	Short: "Add to-do items to the repository queue",
	Long: `
Add to-do items to the repository queue.

Queue is not an active to-do list, but a list of things you'd like to take on
later, perhaps in another branch.

Invoking without arguments will open up the editor for multiple items to be 
added. If there are arguments, all of them will be joined into a single to-do
item.`,
	Run: func(cmd *cobra.Command, args []string) {
		env, tdb := MustInit()
		projId := tdb.FetchProjectId(env.ProjDir, "*")

		if len(args) > 0 {
			tdb.AddTodo(projId, strings.Join(args, " "))
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
			err = tdb.AddTodos(projId, items)
			ExitOnError(err, 1)
		}
	},
}

func init() {
	rootCmd.AddCommand(queueCmd)
}
