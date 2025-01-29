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

	"github.com/spf13/cobra"
)

// whatCmd represents the what command
var whatCmd = &cobra.Command{
	Use:   "what",
	Short: "Displays what's next to do",
	Long: `Displays the first to-do item that isn't completed yet,
starting from the top of the list.

If there is no such item, "All done!" message will be shown.`,
	Run: func(cmd *cobra.Command, args []string) {
		env, tdb := MustInit()
		proj := tdb.GetProject(tdb.FetchProjectId(env.ProjDir, env.Branch))

		fmt.Printf("On: %s\n\n", proj.Name)

		item := tdb.TodoWhat(proj.Id)

		if item == nil {
			fmt.Println("All done!")
		} else {
			fmt.Printf("What to do: %s\n", item.Task)
		}
	},
}

func init() {
	rootCmd.AddCommand(whatCmd)
}
