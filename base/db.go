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
	"os"
	"path/filepath"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

type TodoDb struct {
	db *sqlx.DB
}

func NewTodoDb() (*TodoDb, error) {
	var dbfile string
	params := "?_fk=true&cache=shared&_loc=auto"
	if p := os.Getenv("GITODO_DB"); p != "" {
		path, err := filepath.Abs(p)
		if err != nil {
			return nil, err
		}
		err = os.MkdirAll(filepath.Dir(path), os.ModePerm)
		if err != nil {
			return nil, err
		}
		dbfile = "file:" + path + params
	} else {
		homedir, err := os.UserHomeDir()
		if err != nil {
			homedir = "."
		}
		dbfile = "file:" + homedir + "/.gitodo.db" + params
	}

	return NewTodoDbSrc(dbfile)
}

func NewTodoDbSrc(path string) (*TodoDb, error) {
	db, err := sqlx.Connect("sqlite3", path)

	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(1)

	err = migrations.Migrate(db.DB)

	if err != nil {
		return nil, err
	}

	return &TodoDb{db: db}, nil
}

func (db *TodoDb) Close() error {
	if db.db == nil {
		return nil
	}

	return db.db.Close()
}
