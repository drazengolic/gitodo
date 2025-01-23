package base

import (
	"database/sql"
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

type Timesheet struct {
	Id        int    `db:"timesheet_id"`
	ProjectId int    `db:"project_id"`
	Action    int    `db:"action"`
	CreatedAt string `db:"created_at"`
}

const (
	TimesheetActionStart int = iota + 1
	TimesheetActionStop
)
