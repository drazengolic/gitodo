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
	"encoding/json"
	"time"
)

type Project struct {
	Id     int    `db:"project_id"`
	Folder string `db:"folder"`
	Branch string `db:"branch"`
	Name   string `db:"name"`
}

type Todo struct {
	Id          int            `db:"todo_id"`
	ProjectId   int            `db:"project_id"`
	Task        string         `db:"task"`
	Position    int            `db:"position"`
	CreatedAt   string         `db:"created_at"`
	DoneAt      sql.NullString `db:"done_at"`
	CommittedAt sql.NullString `db:"committed_at"`
}

type TimeEntry struct {
	Id        int    `db:"timesheet_id"`
	ProjectId int    `db:"project_id"`
	Action    int    `db:"action"`
	CreatedAt string `db:"created_at"`
}

type ItemAndBranch struct {
	ItemName   string `db:"item_name"`
	BranchName string `db:"branch_name"`
}

type BranchItem struct {
	ProjectId  int    `db:"project_id"`
	BranchName string `db:"branch_name"`
	ItemCount  int    `db:"item_count"`
}

type ReportItem struct {
	Id        int    `db:"todo_id"`
	ProjectId int    `db:"project_id"`
	Task      string `db:"task"`
	TimeAt    string `db:"time_at"`
}

type ReportTimeEntry struct {
	ProjectId   int    `db:"project_id"`
	From        string `db:"started_at"`
	To          string `db:"stopped_at"`
	DurationSec int    `db:"duration"`
	Running     bool
}

type ReportProject struct {
	Proj             *Project
	CompletedItems   []ReportItem
	CreatedItems     []ReportItem
	TimeEntries      []ReportTimeEntry
	TotalTimeSeconds int
	LatestUpdate     string
	TimerRunning     bool
}

type ReportRepo struct {
	Folder           string           `json:"repo"`
	LatestUpdate     string           `json:"-"`
	Projects         []*ReportProject `json:"projects"`
	TotalTimeSeconds int              `json:"total_sec"`
}

type Report struct {
	From             string
	To               string
	Repos            []*ReportRepo
	TotalTimeSeconds int
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

func (ri ReportItem) MarshalJSON() ([]byte, error) {
	t, _ := time.ParseInLocation(time.DateTime, ri.TimeAt, time.Local)

	return json.Marshal(struct {
		Id     int    `json:"id"`
		Task   string `json:"task"`
		TimeAt string `json:"at"`
	}{Id: ri.Id, Task: ri.Task, TimeAt: t.UTC().Format(time.RFC3339)})
}

func (rte ReportTimeEntry) MarshalJSON() ([]byte, error) {
	from, _ := time.ParseInLocation(time.DateTime, rte.From, time.Local)
	to, _ := time.ParseInLocation(time.DateTime, rte.To, time.Local)

	return json.Marshal(struct {
		From        string `json:"from"`
		To          string `json:"to"`
		DurationSec int    `json:"duration_sec"`
		Running     bool   `json:"running,omitempty"`
	}{
		From:        from.UTC().Format(time.RFC3339),
		To:          to.UTC().Format(time.RFC3339),
		DurationSec: rte.DurationSec,
		Running:     rte.Running,
	})
}

func (rp *ReportProject) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Name             string            `json:"name"`
		Branch           string            `json:"branch"`
		CompletedItems   []ReportItem      `json:"completed"`
		CreatedItems     []ReportItem      `json:"created"`
		TimeEntries      []ReportTimeEntry `json:"timesheet"`
		TotalTimeSeconds int               `json:"total_sec"`
	}{
		Name:             rp.Proj.Name,
		Branch:           rp.Proj.Branch,
		CompletedItems:   rp.CompletedItems,
		CreatedItems:     rp.CreatedItems,
		TimeEntries:      rp.TimeEntries,
		TotalTimeSeconds: rp.TotalTimeSeconds,
	})
}

func (r *Report) MarshalJSON() ([]byte, error) {
	from, _ := time.ParseInLocation(time.DateTime, r.From, time.Local)
	to, _ := time.ParseInLocation(time.DateTime, r.To, time.Local)

	return json.Marshal(struct {
		From     string        `json:"from"`
		To       string        `json:"to"`
		TotalSec int           `json:"total_sec"`
		Repos    []*ReportRepo `json:"repos"`
	}{
		From:     from.UTC().Format(time.RFC3339),
		To:       to.UTC().Format(time.RFC3339),
		TotalSec: r.TotalTimeSeconds,
		Repos:    r.Repos,
	})
}
