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

package base

import (
	"database/sql"
	"fmt"
	"time"
)

type Project struct {
	Id     int    `db:"project_id"`
	Folder string `db:"folder"`
	Branch string `db:"branch"`
	Name   string `db:"name"`
}

type Todo struct {
	Id         int            `db:"todo_id"`
	ProjectId  int            `db:"project_id"`
	Task       string         `db:"task"`
	Position   int            `db:"position"`
	CreatedAt  string         `db:"created_at"`
	DoneAt     sql.NullString `db:"done_at"`
	CommitedAt sql.NullString `db:"commited_at"`
}

type TimeEntry struct {
	Id        int    `db:"timesheet_id"`
	ProjectId int    `db:"project_id"`
	Action    int    `db:"action"`
	CreatedAt string `db:"created_at"`
}

const (
	TimesheetActionStart int = iota + 1
	TimesheetActionStop
)

// Duration returns a duration in seconds of an active timer.
// If the timer entry is a stop entry, method returns zero.
func (ts *TimeEntry) Duration() int {
	if ts == nil || ts.Action == TimesheetActionStop {
		return 0
	}

	since, err := time.ParseInLocation(time.DateTime, ts.CreatedAt, time.Local)
	if err != nil {
		return 0
	}

	return int(time.Since(since).Seconds())
}

func FormatSeconds(s int) string {
	return fmt.Sprintf("%02d:%02d:%02d", s/3600, (s%3600)/60, s%60)
}
