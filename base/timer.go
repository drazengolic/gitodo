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
	return fmt.Sprintf("Timer running in %s [%s]!", e.Proj.Folder, e.Proj.Branch)
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
		return last, &TimerError{msg: "Timer running since " + t.Format(time.ANSIC)}
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
		return nil, nil, &TimerError{msg: "Timer is not running."}
	}

	if last.Action == TimesheetActionStop {
		return nil, nil, &TimerError{msg: "Timer is not running."}
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

// GetProjectTime calculates total amount of seconds recorded for the project,
// including the ongoing time if the timer is active
func (tdb *TodoDb) GetProjectTime(projId int) (int, error) {
	// The first line sums the timestamps according to the action type.
	// The second one adds the current time if the timer is running.
	// Everything is multiplied with 86400 since the calculation is in
	// the amount of days.
	sql := `select round(coalesce((
		(select sum(case when action=1 then -julianday(created_at) else julianday(created_at) end) from timesheet where project_id=$1) +
		(select case when action=1 then julianday('now', 'localtime') else 0 end from timesheet where project_id=$2 order by created_at desc limit 1)
	),0)*86400)`

	row := tdb.db.QueryRowx(sql, projId, projId)
	var seconds int
	err := row.Scan(&seconds)
	return seconds, err
}

func FormatSeconds(s int) string {
	return fmt.Sprintf("%02d:%02d:%02d", s/3600, (s%3600)/60, s%60)
}
