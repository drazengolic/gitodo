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
	"time"

	"github.com/gen2brain/beeep"
	"github.com/spf13/cobra"
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Starts a timer for the given repository+branch",
	Long: `gitodo can also do time tracking for the given branch.

Executing "start" will start the timer. Where possible,
an OS notification will be displayed.

If the timer is already running, an error will be displayed.`,
	Run: func(cmd *cobra.Command, args []string) {
		env, tdb := MustInit()
		projId := tdb.FetchProjectId(env.ProjDir, env.Branch)
		_, err := tdb.StartTimer(projId)
		HandleTimerError(err)

		msg := fmt.Sprintf("timer started on %s", time.Now().Format(time.ANSIC))
		fmt.Println(msg)
		beeep.Alert("gitodo", msg, "")
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
}
