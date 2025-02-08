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
	"os/exec"
	"strings"

	"github.com/drazengolic/gitodo/base"
	"github.com/spf13/cobra"
)

// changelistCmd represents the changelist command
var changelistCmd = &cobra.Command{
	Use:     "changelist",
	Aliases: []string{"cl"},
	Short:   "Display to-do items as a changelist",
	Long: `
Display to-do items as a changelist usable in markdown documents.

By default, it displays only the completed items. If --all flag is set, all
items will be displayed in the form of a GitHub task list.

A pager will be used where available. If using a pager is not desirable, set
the --no-pager flag.`,
	Run: func(cmd *cobra.Command, args []string) {
		env, tdb := MustInit()
		all := cmd.Flags().Changed("all")
		noPager := cmd.Flags().Changed("no-pager")
		builder := strings.Builder{}
		count := 0

		tdb.TodoItems(tdb.FetchProjectId(env.ProjDir, env.Branch), func(t base.Todo) {
			count++
			switch {
			case all && t.DoneAt.Valid:
				builder.WriteString("- [x] ")
				builder.WriteString(t.Task)
			case all && !t.DoneAt.Valid:
				builder.WriteString("- [ ] ")
				builder.WriteString(t.Task)
			case !all && t.DoneAt.Valid:
				builder.WriteString("- ")
				builder.WriteString(t.Task)
			default:
				return
			}
			builder.WriteRune('\n')
		})

		if count == 0 {
			return
		}

		if noPager {
			fmt.Print(builder.String())
			return
		}

		pager := os.Getenv("PAGER")

		if pager == "" {
			fmt.Print(builder.String())
			return
		}

		run := exec.Command(pager)
		run.Stdin = strings.NewReader(builder.String())
		run.Stdout = os.Stdout
		err := run.Run()
		if err != nil {
			fmt.Print(builder.String())
		}
	},
}

func init() {
	rootCmd.AddCommand(changelistCmd)

	changelistCmd.Flags().BoolP("all", "a", false, "show all")
	changelistCmd.Flags().BoolP("no-pager", "n", false, "do not page output")
}
