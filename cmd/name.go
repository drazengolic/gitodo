/*
Copyright © 2024 Drazen Golic

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

// nameCmd represents the name command
var nameCmd = &cobra.Command{
	Use:   "name",
	Short: "Displays or sets the name for your to-do project.",
	Long: `Displays or sets the name for your to-do project.

The name defaults to the active branch. 

When no argument is given, the command will output the current name. 
If there are arguments provided, the first one will be used to set
the project name.`,
	Run: func(cmd *cobra.Command, args []string) {
		env, tdb := MustInit()
		proj := tdb.GetProject(tdb.FetchProjectId(env.ProjDir, env.Branch))

		if len(args) == 0 {
			fmt.Println(proj.Name)
		} else {
			tdb.UpdateProjectName(proj.Id, args[0])
			fmt.Printf("name set to \"%s\"\n", args[0])
		}
	},
}

func init() {
	rootCmd.AddCommand(nameCmd)
}
