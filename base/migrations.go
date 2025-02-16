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

	"github.com/lopezator/migrator"
)

var migrations *migrator.Migrator

func init() {
	// Configure migrations

	var err error
	migrations, err = migrator.New(
		migrator.Migrations(
			&migrator.Migration{
				Name: "Initial structure",
				Func: func(tx *sql.Tx) error {
					sql := `
create table project (
	project_id integer primary key autoincrement,
	folder text not null,
	branch text not null,
	name text not null
);

create unique index idx_project on project (folder, branch);

create table todo (
	todo_id integer primary key autoincrement,
	project_id integer not null,
	task text not null,
	position integer not null default 1,
	created_at text not null 
		default (datetime(current_timestamp, 'localtime')),
	done_at text,
	commited_at text,
	foreign key (project_id) 
      references project (project_id) 
         on delete cascade 
         on update no action
);

create index idx_todo on todo (project_id, created_at);
create index idx_todo_done on todo (project_id, done_at) where done_at is not null;

create table timesheet (
	timesheet_id integer primary key autoincrement,
	project_id integer not null,
	action integer not null default 0,
	created_at text not null 
		default (datetime(current_timestamp, 'localtime')),
	foreign key (project_id) 
      references project (project_id) 
         on delete cascade 
         on update no action
);

create index idx_timesheet on timesheet (project_id, created_at);
`
					if _, err := tx.Exec(sql); err != nil {
						return err
					}
					return nil
				},
			},
		),
		// silence the migrator
		migrator.WithLogger(migrator.LoggerFunc(func(s string, i ...interface{}) {})),
	)

	// something wrong with the migrations, panic
	if err != nil {
		panic(err)
	}
}
