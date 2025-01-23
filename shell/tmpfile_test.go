/*
Copyright © 2025 Drazen Golic

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

package shell

import "testing"

func TestReadItems(t *testing.T) {
	tmp, err := NewTmpFileString(`#ignored
- Not ignored.
 -Still here.


- Another one.
# comment
-Test
- Test2
`)
	if err != nil {
		t.Fatal(err)
	}
	defer tmp.Delete()

	expected := []string{
		"Not ignored.\n -Still here.",
		"Another one.",
		"Test",
		"Test2",
	}

	result, err := tmp.ReadItems()

	if err != nil {
		t.Fatal(err)
	}

	if len(expected) != len(result) {
		t.Fatalf("size not equal. expected %d, got %d", len(expected), len(result))
	}

	for i := 0; i < len(expected); i++ {
		if expected[i] != result[i] {
			t.Errorf("not equal. expected %s, got %s", expected[i], result[i])
		}
	}
}

func TestReadItemsEmpty(t *testing.T) {
	tmp, err := NewTmpFileString(`#ignored
# also ignored
-
-
`)
	if err != nil {
		t.Fatal(err)
	}
	defer tmp.Delete()

	result, err := tmp.ReadItems()

	if err != nil {
		t.Fatal(err)
	}

	if len(result) != 0 {
		t.Fatalf("expected empty, got %v", result)
	}
}
