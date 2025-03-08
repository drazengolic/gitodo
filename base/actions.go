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
	"time"

	"github.com/jmoiron/sqlx"
)

// FetchProjectId gets or creates a project for the given folder and branch
// and returns the project id
func (tdb *TodoDb) FetchProjectId(folder, branch string) int {
	row := tdb.db.QueryRow(
		`select project_id from project where folder = $1 and branch = $2`,
		folder, branch,
	)

	var id int
	err := row.Scan(&id)

	if err == nil {
		return id
	}

	row = tdb.db.QueryRow(
		`insert into project (folder, branch, name) values ($1, $2, $3) returning project_id`,
		folder, branch, branch,
	)

	row.Scan(&id)
	return id
}

func (tdb *TodoDb) TodoCount(projId int) int {
	row := tdb.db.QueryRow(`select count(*) from todo where project_id = $1`, projId)
	var count int
	err := row.Scan(&count)

	if err == nil {
		return count
	}

	return 0
}

func (tdb *TodoDb) AddTodo(projId int, task string) (int, int) {
	count := tdb.TodoCount(projId)
	row := tdb.db.QueryRow(
		`insert into todo (project_id, task, position) values ($1, $2, $3) returning todo_id`,
		projId, task, count+1,
	)

	var id int
	row.Scan(&id)
	return id, count + 1
}

func (tdb *TodoDb) AddTodos(projId int, tasks []string) error {
	count := tdb.TodoCount(projId)
	insert := make([]map[string]any, len(tasks))
	for i, t := range tasks {
		insert[i] = map[string]any{"project_id": projId, "task": t, "position": count + i + 1}
	}
	_, err := tdb.db.NamedExec(`insert into todo(project_id, task, position)
        values (:project_id, :task, :position)`, insert)
	return err
}

func (tdb *TodoDb) GetProject(projId int) Project {
	var proj Project
	tdb.db.Get(&proj, "select project_id, folder, branch, name from project where project_id = $1", projId)
	return proj
}

func (tdb *TodoDb) TodoItems(projId int, f func(t Todo)) error {
	todo := Todo{}
	rows, err := tdb.db.Queryx(`select todo_id, project_id, task, position, created_at, done_at, committed_at 
		from todo where project_id = $1 order by position`, projId)

	if err != nil {
		return err
	}

	for rows.Next() {
		err = rows.StructScan(&todo)
		if err != nil {
			return err
		}
		f(todo)
	}

	return nil
}

func (tdb *TodoDb) TodoItemsDone(projId int, f func(t Todo)) error {
	todo := Todo{}
	rows, err := tdb.db.Queryx(`select todo_id, project_id, task, position, created_at, done_at, committed_at 
		from todo where project_id = $1 and done_at is not null order by done_at`, projId)

	if err != nil {
		return err
	}

	for rows.Next() {
		err = rows.StructScan(&todo)
		if err != nil {
			return err
		}
		f(todo)
	}

	return nil
}

func (tdb *TodoDb) TodoItemsForCommit(projId int, previous bool, f func(t Todo)) error {
	todo := Todo{}
	sql := `select todo_id, project_id, task, position, created_at, done_at, committed_at 
		from todo where project_id = :projId and done_at is not null`

	if previous {
		sql += ` and (committed_at = (select max(committed_at) from todo where project_id=:projId)
		or committed_at is null)`
	} else {
		sql += " and committed_at is null"
	}

	sql += " order by done_at"

	rows, err := tdb.db.NamedQuery(sql, map[string]any{
		"projId": projId,
	})

	if err != nil {
		return err
	}

	for rows.Next() {
		err = rows.StructScan(&todo)
		if err != nil {
			return err
		}
		f(todo)
	}

	return nil
}

func (tdb *TodoDb) TodoWhat(projId int) *Todo {
	sql := `select todo_id, project_id, task, position, created_at, done_at, committed_at 
		from todo where project_id = $1 and done_at is null order by position limit 1`
	row := tdb.db.QueryRowx(sql, projId)
	todo := Todo{}
	err := row.StructScan(&todo)

	if err != nil {
		return nil
	}

	return &todo
}

func (tdb *TodoDb) ChangePosition(todoId, from, to int) error {
	sql := `update todo set position = case 
		when todo_id = :todoId then :to
		when position >= :to and position < :from then position + 1
		when position > :from and position <= :to then position - 1
		else position
	end where project_id=(select project_id from todo where todo_id=:todoId)`

	_, err := tdb.db.NamedExec(sql, map[string]any{
		"todoId": todoId,
		"from":   from,
		"to":     to,
	})
	return err
}

// MoveTodo moves an item to another project and updates positions
func (tdb *TodoDb) MoveTodo(todoId, projId int) error {
	count := tdb.TodoCount(projId)

	tx, err := tdb.db.Beginx()

	if err != nil {
		return err
	}
	defer tx.Rollback()

	// update positions first
	_, err = tx.Exec(`with my_item as (
		select project_id, position from todo
		where todo_id=$1
	)
	update todo set position=position-1
	where project_id = (select project_id from my_item limit 1)
	and position > (select position from my_item limit 1)`, todoId)

	if err != nil {
		return err
	}

	// move
	_, err = tx.Exec(`update todo set project_id=$1, position=$2, 
	created_at=datetime(current_timestamp, 'localtime') where todo_id=$3`, projId, count+1, todoId)

	if err != nil {
		return err
	}

	return tx.Commit()
}

func (tdb *TodoDb) TodoDone(todoId int, done bool) error {
	var sql string
	if done {
		sql = "update todo set done_at=datetime(current_timestamp, 'localtime') where todo_id=$1"
	} else {
		sql = "update todo set done_at=null where todo_id=$1"
	}
	_, err := tdb.db.Exec(sql, todoId)
	return err
}

func (tdb *TodoDb) Delete(todoId int) error {
	_, err := tdb.db.Exec("delete from todo where todo_id=$1", todoId)
	return err
}

func (tdb *TodoDb) UpdateTask(todoId int, task string) error {
	_, err := tdb.db.Exec("update todo set task=$1 where todo_id=$2", task, todoId)
	return err
}

func (tdb *TodoDb) UpdateProjectName(projId int, name string) error {
	_, err := tdb.db.Exec("update project set name=$1 where project_id=$2", name, projId)
	return err
}

func (tdb *TodoDb) SetItemsCommitted(projId int, previous bool) error {
	sql := `update todo set committed_at=:ts where project_id=:projId and done_at is not null`

	if previous {
		sql += ` and (committed_at = (select max(committed_at) from todo where project_id=:projId)
		or committed_at is null)`
	} else {
		sql += " and committed_at is null"
	}

	_, err := tdb.db.NamedExec(sql, map[string]any{
		"projId": projId,
		"ts":     time.Now().Format(time.DateTime),
	})
	return err
}

func (tdb *TodoDb) GetItemsAndBranch(ids []int) ([]*ItemAndBranch, error) {
	var resultSet []*ItemAndBranch
	q, args, err := sqlx.In(
		`select t.task as item_name, p.branch as branch_name from todo t natural join project p 
		where t.todo_id in (?) order by p.project_id desc, t.created_at desc`, ids)

	if err != nil {
		return nil, err
	}

	err = tdb.db.Select(&resultSet, tdb.db.Rebind(q), args...)

	if err != nil {
		return nil, err
	}

	return resultSet, nil
}

func (tdb *TodoDb) GetBranches(repo string) ([]BranchItem, error) {
	var resultSet []BranchItem
	sql := `select p.branch as branch_name, count(*) as item_count, p.project_id
	from project p
	natural join todo t
	where p.folder = ? and p.branch != '*'
	group by p.branch`

	if err := tdb.db.Select(&resultSet, sql, repo); err != nil {
		return nil, err
	}

	return resultSet, nil
}

func (tdb *TodoDb) DeleteProject(projId int) error {
	_, err := tdb.db.Exec("delete from project where project_id = ?", projId)
	return err
}

func (tdb *TodoDb) Vacuum() error {
	_, err := tdb.db.Exec("vacuum")
	return err
}

func (tdb *TodoDb) CopyProjectItems(projFrom, projTo int) error {
	sql := `insert into todo (project_id, task, position) 
	select ?, task, position from todo where project_id = ?`

	_, err := tdb.db.Exec(sql, projTo, projFrom)
	return err
}
