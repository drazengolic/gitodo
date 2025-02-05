/*
Copyright © 2024 Dražen Golić

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
	"time"

	"github.com/drazengolic/gitodo/base"
	"github.com/gen2brain/beeep"
	"github.com/spf13/cobra"
)

// stopCmd represents the stop command
var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stops the timer from anywhere",
	Long: `If there is a timer running, it will be stopped with the "stop" command.

The command can be run from anywhere, it is not required
to be in the same repository or at the same branch
where the timer has started.`,
	Run: func(cmd *cobra.Command, args []string) {
		tdb, err := base.NewTodoDb()
		ExitOnError(err, 1)

		_, prev, err := tdb.StopTimer()

		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		proj := tdb.GetProject(prev.ProjectId)
		fmt.Printf(
			"timer stopped on %s.\n\nrepository: %q\nbranch: %s\nduration: %s\n",
			time.Now().Format(time.ANSIC),
			proj.Folder,
			proj.Branch,
			prev.Duration(),
		)
		beeep.Alert("gitodo", "Timer stopped.", "")
	},
}

func init() {
	rootCmd.AddCommand(stopCmd)
}
