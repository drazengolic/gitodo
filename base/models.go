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

func (ts *TimeEntry) Duration() string {
	if ts == nil || ts.Action == TimesheetActionStop {
		return ""
	}

	since, err := time.ParseInLocation(time.DateTime, ts.CreatedAt, time.Local)
	if err != nil {
		return ""
	}

	d := int(time.Since(since).Seconds())
	return fmt.Sprintf("%02d:%02d:%02d", d/3600, (d%3600)/60, d%60)
}
