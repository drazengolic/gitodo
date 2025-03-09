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
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/drazengolic/gitodo/base"
	"github.com/spf13/cobra"
)

// reportCmd represents the report command
var reportCmd = &cobra.Command{
	Use:   "report [days]",
	Short: "View activity report",
	Long: `
View the activity report for a given period of time that displays repositories,
projects/branches, completed items, added but not completed items, and
recorded time if any.

The command can be executed anywhere, it is not required to be within a git
repository.

The command takes one argument that represents the number of days to look back
for the data since the moment of requesting the report. Default value is 1.

If flags --from and --to are provided, the "days" argument is ignored and the 
given interval is used instead. Both flags must be provided.

To limit the report only to git repositories under a certain directory (child
directories included), use the --dir flag. Relative paths are supported.

To get the report in a JSON format that also contains more details than the
default screen, set the --json flag. This flag, together with --from and --to
can be used for automation scripts i.e. a cron job to feed the external systems
(like time tracking or project management software) with the recorded data.
When exporting to JSON, every timestamp will be converted to UTC.
`,
	Run: func(cmd *cobra.Command, args []string) {
		path := ""
		if cmd.Flags().Changed("path") {
			path2, _ := cmd.Flags().GetString("path")
			path2, err := filepath.Abs(path2)
			ExitOnError(err, 1)
			path = path2
		}

		days := 1
		if len(args) > 0 {
			d, err := strconv.Atoi(args[0])
			if err != nil {
				fmt.Printf("'%s' is not a valid integer number larger than 0\n", args[0])
				os.Exit(1)
			}
			days = d
		}

		if days < 1 {
			fmt.Println("The days argument can't be lower than 1")
			os.Exit(1)
		}

		toTime := time.Now()
		fromTime := toTime.Add(-24 * time.Duration(days) * time.Hour)
		builder := strings.Builder{}

		if cmd.Flags().Changed("from") != cmd.Flags().Changed("to") {
			fmt.Println("Both --from and --to must be set.")
			os.Exit(1)
		}

		if cmd.Flags().Changed("from") && cmd.Flags().Changed("to") {
			fromStr, _ := cmd.Flags().GetString("from")
			toStr, _ := cmd.Flags().GetString("to")

			ft, err := time.Parse(time.RFC3339, fromStr)

			if err != nil {
				fmt.Printf("Could not parse %q as RFC3339 date and time string.\n", fromStr)
				os.Exit(1)
			}

			tt, err := time.Parse(time.RFC3339, toStr)
			if err != nil {
				fmt.Printf("Could not parse %q as RFC3339 date and time string.\n", toStr)
				os.Exit(1)
			}

			if ft.Compare(tt) >= 0 {
				fmt.Println("'from' time can't be larger than 'to' time.")
				os.Exit(1)
			}

			toTime = tt.Local()
			fromTime = ft.Local()

			builder.WriteString(fmt.Sprintf(
				"Activity between\n%s and %s\n\n",
				fromTime.Format(time.ANSIC),
				toTime.Format(time.ANSIC),
			))
		} else {
			builder.WriteString("Activity since ")
			builder.WriteString(fromTime.Format(time.ANSIC))
			builder.WriteString("\n\n")
		}

		tdb, err := base.NewTodoDb()
		ExitOnError(err, 1)

		report, err := tdb.CreateReport(
			fromTime.Format(time.DateTime),
			toTime.Format(time.DateTime),
			path,
		)
		ExitOnError(err, 1)

		// output json
		if cmd.Flags().Changed("json") {
			b, err := json.MarshalIndent(report, "", "  ")
			ExitOnError(err, 1)
			fmt.Printf("%s\n", b)
			return
		}

		// continue console output

		if len(report.Repos) == 0 {
			fmt.Print(builder.String())
			fmt.Println("No data to display.")
			return
		}

		useColors := !cmd.Flags().Changed("pager")

		today := time.Now().Format(time.DateOnly)
		yesterday := time.Now().Add(-24 * time.Hour).Format(time.DateOnly)

		for _, repo := range report.Repos {
			builder.WriteString(txtRender(repo.Folder, &magentaText, useColors))
			builder.WriteRune('\n')

			for _, proj := range repo.Projects {
				if proj.Proj.Branch != proj.Proj.Name {
					builder.WriteString(txtRender(proj.Proj.Name, &blueText, useColors))
					builder.WriteRune('\n')
					builder.WriteString(txtRender(proj.Proj.Branch, &dimmedText, useColors))
				} else {
					builder.WriteString(txtRender(proj.Proj.Branch, &blueText, useColors))
				}

				switch {
				case strings.HasPrefix(proj.LatestUpdate, today):
					builder.WriteString(txtRender(" • updated today", &dimmedText, useColors))
				case strings.HasPrefix(proj.LatestUpdate, yesterday):
					builder.WriteString(txtRender(" • updated yesterday", &dimmedText, useColors))
				default:
					builder.WriteString(txtRender(" • updated on "+proj.LatestUpdate[0:10], &dimmedText, useColors))
				}

				builder.WriteRune('\n')

				if len(proj.CompletedItems) > 0 {
					builder.WriteString("\nCompleted:\n")
					for _, item := range proj.CompletedItems {
						builder.WriteString(fmt.Sprintf("  - %s\n", item.Task))
					}
				}

				if len(proj.CreatedItems) > 0 {
					builder.WriteString("\nAdded:\n")
					for _, item := range proj.CreatedItems {
						builder.WriteString(fmt.Sprintf("  - %s\n", item.Task))
					}
				}

				if proj.TotalTimeSeconds > 0 {
					builder.WriteString(fmt.Sprintf("\nTime: %s", base.FormatSeconds(proj.TotalTimeSeconds)))
					if proj.TimerRunning {
						builder.WriteString(txtRender(" (running)", &orangeText, useColors))
						builder.WriteRune('\n')
					} else {
						builder.WriteRune('\n')
					}
				}

				builder.WriteRune('\n')

			}

			if repo.TotalTimeSeconds > 0 && len(repo.Projects) > 1 {
				builder.WriteString(txtRender("Repo time: "+base.FormatSeconds(repo.TotalTimeSeconds), &orangeText, useColors))
				builder.WriteString("\n\n")
			}
		}

		if report.TotalTimeSeconds > 0 {
			builder.WriteString(txtRender("Total time: "+base.FormatSeconds(report.TotalTimeSeconds), &greenTextStyle, useColors))
			builder.WriteRune('\n')
		}

		output := builder.String()
		pager := os.Getenv("PAGER")

		if pager == "" || useColors {
			fmt.Print(output)
			return
		}

		run := exec.Command(pager)
		run.Stdin = strings.NewReader(output)
		run.Stdout = os.Stdout
		err = run.Run()
		if err != nil {
			fmt.Print(output)
		}
	},
}

// txtRender is a helper function to make style rendering optional
func txtRender(txt string, style *lipgloss.Style, useStyle bool) string {
	if useStyle {
		return style.Render(txt)
	} else {
		return txt
	}
}

func init() {
	rootCmd.AddCommand(reportCmd)

	reportCmd.Flags().BoolP("json", "j", false, "Print the report in JSON format")
	reportCmd.Flags().StringP("from", "f", "", "From what time (RFC3339) to read data")
	reportCmd.Flags().StringP("to", "t", "", "To what time (RFC3339) to read data")
	reportCmd.Flags().StringP("dir", "d", "", "Limit report to the repositories in this directory")
	reportCmd.Flags().BoolP("pager", "p", false, "use PAGER for output")
}
