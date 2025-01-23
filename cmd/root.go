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
	"os"

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
gitodo is a to-do list companion for git projects
that ties to-do items to git repositories and branches
without storing any files in the repositories.

A minimalist tool that helps the busy developers to:
 - keep track of what they've done and what they need
   to do per branch
 - add ideas in the queue for later
 - make stashing and popping of changes easier
 - craft commit messages based on the work done
 - prepare changelists
 - track time
 - view reports

All configuration is read from git and the environment, 
no yaml files needed.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		env, tdb := MustInit()
		projId := tdb.FetchProjectId(env.ProjDir, env.Branch)
		count := tdb.TodoCount(projId)

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

		} else {
			ui.RunTodoListUI(env, tdb)
		}
	},
}

var projectBranch string

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&projectBranch, "branch", "", "branch name to use")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// MustInit collects data and creates instances necessary for the app to function
func MustInit() (*shell.DirEnv, *base.TodoDb) {
	env, err := shell.GetDirEnv()
	ExitOnError(err, 1)
	if projectBranch != "" {
		env.Branch = projectBranch
	}
	tdb, err := base.NewTodoDb()
	ExitOnError(err, 1)
	return env, tdb
}

func ExitOnError(err error, code int) {
	if err != nil {
		fmt.Printf("error: %v", err)
		os.Exit(code)
	}
}
