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

	"github.com/spf13/cobra"
)

// utilListCmd represents the utilList command
var utilCopyItemsCmd = &cobra.Command{
	Use:   "copy-items from to",
	Short: "Copy to-do items from one branch to another",
	Long: `
Copy to-do items from one branch to another`,
	Run: func(cmd *cobra.Command, args []string) {
		env, tdb := MustInit()
		if len(args) != 2 {
			fmt.Println("Invalid number of arguments.")
			os.Exit(1)
		}
		projFrom := tdb.FetchProjectId(env.ProjDir, args[0])
		projTo := tdb.FetchProjectId(env.ProjDir, args[1])
		err := tdb.CopyProjectItems(projFrom, projTo)
		ExitOnError(err, 1)
		fmt.Printf(
			"Copied items: %d\nNew count: %d\n",
			tdb.TodoCount(projFrom),
			tdb.TodoCount(projTo),
		)
	},
}

func init() {
	utilCmd.AddCommand(utilCopyItemsCmd)
}
