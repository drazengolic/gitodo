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
	"maps"
	"slices"
	"strings"
	"time"
)

func (tdb *TodoDb) CreateReport(from, to, folderFilter string) (*Report, error) {
	report := &Report{From: from, To: to, Repos: []*ReportRepo{}}
	projectMap := make(map[int]*ReportProject)

	// crunch completed items
	if err := tdb.reportCompletedItems(from, to, folderFilter, func(item ReportItem) {
		reportProj := projectMap[item.ProjectId]
		// create project if missing
		if reportProj == nil {
			reportProj = &ReportProject{
				CompletedItems: []ReportItem{},
				CreatedItems:   []ReportItem{},
				TimeEntries:    []ReportTimeEntry{},
			}
			projectMap[item.ProjectId] = reportProj
		}
		reportProj.CompletedItems = append(reportProj.CompletedItems, item)
		if item.TimeAt > reportProj.LatestUpdate {
			reportProj.LatestUpdate = item.TimeAt
		}
	}); err != nil {
		return nil, err
	}

	// crunch created items
	if err := tdb.reportCreatedItems(from, to, folderFilter, func(item ReportItem) {
		reportProj := projectMap[item.ProjectId]
		// create project if missing
		if reportProj == nil {
			reportProj = &ReportProject{
				CompletedItems: []ReportItem{},
				CreatedItems:   []ReportItem{},
				TimeEntries:    []ReportTimeEntry{},
			}
			projectMap[item.ProjectId] = reportProj
		}
		reportProj.CreatedItems = append(reportProj.CreatedItems, item)
		if item.TimeAt > reportProj.LatestUpdate {
			reportProj.LatestUpdate = item.TimeAt
		}
	}); err != nil {
		return nil, err
	}

	// crunch time entries
	if err := tdb.reportTimeEntries(from, to, folderFilter, func(entry ReportTimeEntry) {
		reportProj := projectMap[entry.ProjectId]
		// create project if missing
		if reportProj == nil {
			reportProj = &ReportProject{
				CompletedItems: []ReportItem{},
				CreatedItems:   []ReportItem{},
				TimeEntries:    []ReportTimeEntry{},
			}
			projectMap[entry.ProjectId] = reportProj
		}
		reportProj.TimeEntries = append(reportProj.TimeEntries, entry)
		reportProj.TotalTimeSeconds += entry.DurationSec
		if entry.To > reportProj.LatestUpdate {
			reportProj.LatestUpdate = entry.To
		}

	}); err != nil {
		return nil, err
	}

	// check for running timer
	te := tdb.GetLatestTimeEntry()
	if te != nil && (te.Action != TimesheetActionStart || te.CreatedAt > to) {
		te = nil
	}
	// fetch projects and repos
	repoMap := make(map[string]*ReportRepo)
	for id, p := range projectMap {
		proj := tdb.GetProject(id)
		p.Proj = &proj
		repo := repoMap[proj.Folder]
		if repo == nil {
			repo = &ReportRepo{Folder: proj.Folder, Projects: []*ReportProject{p}}
			repoMap[proj.Folder] = repo
		} else {
			repo.Projects = append(repo.Projects, p)
		}
		if p.LatestUpdate > repo.LatestUpdate {
			repo.LatestUpdate = p.LatestUpdate
		}
		// patch with the running timer record
		if te != nil && te.ProjectId == id {
			tto, _ := time.ParseInLocation(time.DateTime, to, time.Local)
			tte, _ := time.ParseInLocation(time.DateTime, te.CreatedAt, time.Local)
			duration := int(tto.Sub(tte).Seconds())

			p.TimeEntries = append(p.TimeEntries, ReportTimeEntry{
				ProjectId:   id,
				From:        te.CreatedAt,
				To:          to,
				DurationSec: duration,
				Running:     true,
			})
			p.TotalTimeSeconds += duration
			p.TimerRunning = true
		}

		repo.TotalTimeSeconds += p.TotalTimeSeconds
		report.TotalTimeSeconds += p.TotalTimeSeconds
	}

	report.Repos = slices.Collect(maps.Values(repoMap))

	// sort everything in descending order
	if len(report.Repos) > 1 {
		slices.SortFunc(report.Repos, func(p1, p2 *ReportRepo) int {
			return strings.Compare(p2.LatestUpdate, p1.LatestUpdate)
		})
	}
	for _, repo := range report.Repos {
		if len(repo.Projects) > 1 {
			slices.SortFunc(repo.Projects, func(p1, p2 *ReportProject) int {
				return strings.Compare(p2.LatestUpdate, p1.LatestUpdate)
			})
		}
	}

	return report, nil
}

func (tdb *TodoDb) reportCompletedItems(from, to, folderFilter string, f func(r ReportItem)) error {
	sql := `select t.todo_id, t.project_id, t.task, t.done_at as time_at from todo t
	natural join project p
	where t.done_at >= ? and t.done_at <= ? and p.folder like ? || '%' and p.branch != '*'
	order by t.project_id desc, t.done_at desc`

	rows, err := tdb.db.Queryx(sql, from, to, folderFilter)
	if err != nil {
		return err
	}

	var item ReportItem
	for rows.Next() {
		err = rows.StructScan(&item)
		if err != nil {
			return err
		}
		f(item)
	}

	return nil
}

func (tdb *TodoDb) reportCreatedItems(from, to, folderFilter string, f func(r ReportItem)) error {
	sql := `select t.todo_id, t.project_id, t.task, t.created_at as time_at from todo t 
	natural join project p
	where t.created_at >= ? and t.created_at <= ? and t.done_at is null 
	and p.folder like ? || '%' and p.branch != '*'
	order by t.project_id desc, t.created_at desc`

	rows, err := tdb.db.Queryx(sql, from, to, folderFilter)
	if err != nil {
		return err
	}

	var item ReportItem
	for rows.Next() {
		err = rows.StructScan(&item)
		if err != nil {
			return err
		}
		f(item)
	}

	return nil
}

func (tdb *TodoDb) reportTimeEntries(from, to, folderFilter string, f func(t ReportTimeEntry)) error {
	sql := `
	with entries as (
		select t.project_id, t.action, t.created_at as stopped_at, 
		lag(t.created_at) over (order by t.created_at) as started_at from timesheet t
		natural join project p
		where t.created_at >= ? and t.created_at <= ? and p.folder like ? || '%' and p.branch != '*'
		order by t.project_id desc, t.created_at
	)
	select project_id, started_at, stopped_at, round((julianday(stopped_at)-julianday(started_at))*86400) as duration
	from entries where action = 2 and started_at is not null`

	rows, err := tdb.db.Queryx(sql, from, to, folderFilter)
	if err != nil {
		return err
	}

	var entry ReportTimeEntry
	for rows.Next() {
		err = rows.StructScan(&entry)
		if err != nil {
			return err
		}
		f(entry)
	}

	return nil
}
