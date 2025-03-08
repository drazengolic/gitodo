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

	"github.com/drazengolic/gitodo/shell"
	"github.com/spf13/cobra"
)

// utilListCmd represents the utilList command
var utilListCmd = &cobra.Command{
	Use:   "list",
	Short: "List branches with data",
	Long: `
List branches used with gitodo, together with a number of todo items.
If a branch does not exist within the repository, it will be printed in red.`,
	Run: func(cmd *cobra.Command, args []string) {
		env, tdb := MustInit()
		projBranches, err := tdb.GetBranches(env.ProjDir)
		ExitOnError(err, 1)
		envBranches, err := shell.ListBranches()
		ExitOnError(err, 1)

		branchMap := make(map[string]struct{})
		for _, b := range envBranches {
			branchMap[b] = struct{}{}
		}

		for _, pb := range projBranches {
			_, ok := branchMap[pb.BranchName]
			txt := fmt.Sprintf("%s (%d)", pb.BranchName, pb.ItemCount)
			if ok {
				fmt.Println(txt)
			} else {
				fmt.Println(redText.Render(txt))
			}
		}
	},
}

func init() {
	utilCmd.AddCommand(utilListCmd)
}
