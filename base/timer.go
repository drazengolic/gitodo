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
	"fmt"
	"time"
)

type TimerRunningElsewhereError struct {
	Proj  *Project
	Entry *TimeEntry
}

func (e *TimerRunningElsewhereError) Error() string {
	return fmt.Sprintf("timer running in %s [%s]!", e.Proj.Folder, e.Proj.Branch)
}

type TimerError struct {
	msg string
}

func (e *TimerError) Error() string {
	return e.msg
}

// GetLatestTimeEntry returns the last recorded time entry if found
func (tdb *TodoDb) GetLatestTimeEntry() *TimeEntry {
	sql := `select timesheet_id, project_id, action, created_at from timesheet
	order by created_at desc limit 1`
	row := tdb.db.QueryRowx(sql)
	ts := TimeEntry{}
	err := row.StructScan(&ts)

	if err != nil {
		return nil
	}

	return &ts
}

// CheckTimer gets the latest time entry and checks it against the project id
func (tdb *TodoDb) CheckTimer(projId int) (*TimeEntry, error) {
	last := tdb.GetLatestTimeEntry()

	if last != nil && last.ProjectId != projId && last.Action == TimesheetActionStart {
		proj := tdb.GetProject(projId)
		return last, &TimerRunningElsewhereError{Proj: &proj, Entry: last}
	}

	return last, nil
}

// StartTimer creates a new start entry, errors if already started
func (tdb *TodoDb) StartTimer(projId int) (*TimeEntry, error) {
	last, err := tdb.CheckTimer(projId)

	if err != nil {
		return last, err
	}

	if last != nil && last.Action == TimesheetActionStart {
		t, _ := time.Parse(time.DateTime, last.CreatedAt)
		return last, &TimerError{msg: "timer running since " + t.Format(time.ANSIC)}
	}

	sql := `insert into timesheet (project_id, action) values ($1, $2) returning *`
	row := tdb.db.QueryRowx(sql, projId, TimesheetActionStart)
	ts := TimeEntry{}
	err = row.StructScan(&ts)

	if err != nil {
		return last, err
	}

	return &ts, nil
}

// StopTimer creates a new stop entry and returns both the new and
// the previous entry. Returns error if the timer is stopped.
func (tdb *TodoDb) StopTimer() (*TimeEntry, *TimeEntry, error) {
	last := tdb.GetLatestTimeEntry()

	if last == nil {
		return nil, nil, &TimerError{msg: "timer is not running."}
	}

	if last.Action == TimesheetActionStop {
		return nil, nil, &TimerError{msg: "timer is not running."}
	}

	sql := `insert into timesheet (project_id, action) values ($1, $2) returning *`
	row := tdb.db.QueryRowx(sql, last.ProjectId, TimesheetActionStop)
	ts := TimeEntry{}
	err := row.StructScan(&ts)

	if err != nil {
		return nil, nil, err
	}

	return &ts, last, nil
}
