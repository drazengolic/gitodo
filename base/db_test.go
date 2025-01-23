package base

import "testing"

// TestNewTodoDbSrc verifies that a new db instance is
// created successfuly.
func TestNewTodoDbSrc(t *testing.T) {
	db, err := NewTodoDbSrc("file:test.db?mode=memory&_fk=true")

	if err != nil {
		t.Fatalf("Got error: %v", err)
	}

	db.Close()
}
