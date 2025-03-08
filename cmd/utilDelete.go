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
	"golang.org/x/term"
)

// utilDeleteCmd represents the utilDelete command
var utilDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete branch data",
	Long: `
Delete to-do items and related data for all branch names provided as arguments.`,
	Run: func(cmd *cobra.Command, args []string) {
		env, tdb := MustInit()
		deleteMap := map[string]struct{}{}
		for _, b := range args {
			deleteMap[b] = struct{}{}
		}

		yes := cmd.Flags().Changed("yes")
		branches, err := tdb.GetBranches(env.ProjDir)
		ExitOnError(err, 1)

		key := make([]byte, 1)

		for _, b := range branches {
			_, del := deleteMap[b.BranchName]

			if !yes && del {
				fmt.Printf("Delete %q? (y/n) ", b.BranchName)
				// set up single key reading
				oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
				if err != nil {
					fmt.Println(err)
					return
				}
				os.Stdin.Read(key)
				term.Restore(int(os.Stdin.Fd()), oldState)

				r := key[0]
				del = r == 'y' || r == 'Y'
				fmt.Println("")
			}

			if del {
				fmt.Printf("Deleting %q...", b.BranchName)
				err := tdb.DeleteProject(b.ProjectId)
				if err != nil {
					fmt.Println(err.Error())
				} else {
					fmt.Println("ok")
				}
			}

			delete(deleteMap, b.BranchName)
		}

		for notFound, _ := range deleteMap {
			fmt.Printf("Not found: %q\n", notFound)
		}

		fmt.Print("Optimizing...")
		err = tdb.Vacuum()
		if err != nil {
			fmt.Println(err.Error())
		} else {
			fmt.Println("ok")
		}
	},
}

func init() {
	utilCmd.AddCommand(utilDeleteCmd)
	utilDeleteCmd.Flags().BoolP("yes", "y", false, "Delete without asking")
}
