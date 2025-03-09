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

	"github.com/charmbracelet/lipgloss"
	"github.com/drazengolic/gitodo/base"
	"github.com/drazengolic/gitodo/shell"
	"github.com/drazengolic/gitodo/ui"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "gitodo",
	Short: "The stupid to-do list application for git projects",
	Long: `
gitodo is a to-do list companion for git projects that ties to-do items to git 
repositories and branches without storing any files in the actual repositories.

A minimalist tool that helps the busy developers to:
 - keep track of what they've done and what they need
   to do per branch
 - add ideas in the queue for later
 - make stashing and popping of changes easier
 - craft commit messages based on the work done
 - prepare changelists
 - track time
 - view reports

All configuration is read from git and the environment, no yaml files needed.

Running the application without arguments will either:
  - open up the editor to add items if none are found
  - open a TUI screen where to-do items can be managed

The invoked editor will be the same one that git invokes.

To-do items do not have a priority. The top-most item should be always the one 
with the top priority, and commands like "what" and "done" read items from top
to bottom. Use the TUI screen to change the order of the items.

When stashing changes for an item, the "--include-untracked" flag will be 
passed to git, so if you don't want to have some untracked files to be stashed,
make sure to add them to .gitignore file or move them somewhere else. 

By default, gitodo will store the database file into the current user's home
directory. To override the path to the database file, set GITODO_DB environment
variable to a desired path to the file.

  `,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		env, tdb := MustInit()
		projId := tdb.FetchProjectId(env.ProjDir, env.Branch)
		count := tdb.TodoCount(projId)

		_, err := tdb.CheckTimer(projId)
		HandleTimerError(err)

		if count == 0 {
			tmpfile, err := shell.NewTmpFileString(`# Start a line with a hyphen (-) to indicate a new item.
# Comments like this are ignored.
- `)
			ExitOnError(err, 1)
			err = tmpfile.Edit(env.Editor, 3)
			ExitOnError(err, 1)
			items, err := tmpfile.ReadItems()

			tmpfile.Delete()

			ExitOnError(err, 1)
			tdb.AddTodos(projId, items)

			fmt.Printf("Added %d item(s) to %q\n", len(items), env.Branch)
		} else {
			ui.RunTodoListUI(env, tdb)
		}
	},
}

var (
	redText    = lipgloss.NewStyle().Foreground(lipgloss.Color("#ff0000"))
	orangeText = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#ff7500", Dark: "#ffa500"})
	dimmedText = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#777777"))
	blueText = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "12", Dark: "86"})
	strikeTroughStyle = lipgloss.NewStyle().Strikethrough(true)
	greenTextStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#00ff00"))
	boldText          = lipgloss.NewStyle().Bold(true)
	magentaText       = lipgloss.NewStyle().Foreground(lipgloss.Color("5"))
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
}

// MustInit collects data and creates instances necessary for the app to function
func MustInit() (*shell.DirEnv, *base.TodoDb) {
	env, err := shell.GetDirEnv()
	ExitOnError(err, 1)
	tdb, err := base.NewTodoDb()
	ExitOnError(err, 1)
	return env, tdb
}

func ExitOnError(err error, code int) {
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(code)
	}
}

func HandleTimerError(err error) {
	if err != nil {
		switch e := err.(type) {
		case *base.TimerError:
			fmt.Println(err.Error())
		case *base.TimerRunningElsewhereError:
			format := `There is a timer running for another project!

Repository: %q
Branch: %s
Duration: %s

To continue, please cd/checkout to the given repository/branch,
or type "gitodo stop" (in any directory) to stop the active timer.`
			fmt.Println(redText.Render(fmt.Sprintf(format, e.Proj.Folder, e.Proj.Branch, base.FormatSeconds(e.Entry.Duration()))))
		default:
			fmt.Println(redText.Render(e.Error()))
		}
		os.Exit(1)
	}
}
