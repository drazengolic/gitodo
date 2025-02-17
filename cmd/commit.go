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
	"os"
	"os/exec"
	"strings"

	"github.com/drazengolic/gitodo/base"
	"github.com/drazengolic/gitodo/shell"
	"github.com/spf13/cobra"
)

// commitCmd represents the commit command
var commitCmd = &cobra.Command{
	Use:                "commit [git flags]",
	DisableFlagParsing: true,
	Short:              "Run a git commit with a prepared message",
	Long: `
Run "git commit" with a prepared message based on the completed to-do items
that will also be marked as committed if the commit was successful.

Items that are previously marked as committed will not be included in the 
message unless "--amend" flag is provided.

By default, the command will execute "git commit -eF msgfile", and any 
additional arguments or flags passed to this command will be appended to the
base command.

Special flag handling:

 - if "--amend" flag is passed to commit, the msgfile will contain all of the
   completed to-do items that either aren't flagged as committed, or they were
   flagged as committed in the previously executed commit.

 - if "--no-edit" is passed together with "--amend", no message will be 
   generated and "-eF" will be left out
`,
	Run: func(cmd *cobra.Command, args []string) {
		env, tdb := MustInit()
		amend, noEdit := false, false

		for _, arg := range args {
			switch {
			case arg == "--amend":
				amend = true
			case arg == "--no-edit":
				noEdit = true
			}
		}

		proj := tdb.GetProject(tdb.FetchProjectId(env.ProjDir, env.Branch))

		_, err := tdb.CheckTimer(proj.Id)
		HandleTimerError(err)

		if amend && noEdit {
			err := runCommit(args)
			ExitOnError(err, 1)
			err = tdb.SetItemsCommitted(proj.Id, true)
			ExitOnError(err, 1)
			return
		}

		builder := strings.Builder{}

		if proj.Name != proj.Branch {
			builder.WriteRune('#')
			builder.WriteString(proj.Name)
			builder.WriteString("\n\n")
		}

		tdb.TodoItemsForCommit(proj.Id, amend, func(t base.Todo) {
			builder.WriteString("- ")
			builder.WriteString(t.Task)
			builder.WriteRune('\n')
		})

		file, err := shell.NewTmpFileString(builder.String())
		ExitOnError(err, 1)

		err = runCommit(append([]string{"-eF", file.Path()}, args...))
		if err == nil {
			err = tdb.SetItemsCommitted(proj.Id, amend)
		}
		file.Delete()
		ExitOnError(err, 1)
	},
}

func init() {
	rootCmd.AddCommand(commitCmd)
	// all arguments are passed trough to git, so remove the flags from the doc
	commitCmd.PersistentFlags().BoolP("help", "h", false, "Print usage")
	commitCmd.PersistentFlags().Lookup("help").Hidden = true
	commitCmd.PersistentFlags().StringP("branch", "", "", "Unused branch")
	commitCmd.PersistentFlags().Lookup("branch").Hidden = true
}

func runCommit(args []string) error {
	args = append([]string{"commit"}, args...)
	cmd := exec.Command("git", args...)
	cmd.Env = os.Environ()
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
