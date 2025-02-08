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

package ui

import "testing"

func TestGetStashTimestamp(t *testing.T) {
	expected := "Tue Jan 14 19:13:06 2025"

	item := todoItem{stash: "stash@{Tue Jan 14 19:13:06 2025}: On master: gitodo_7"}

	if s := item.getStashTimestamp(); s != expected {
		t.Errorf("not equal, got %s", s)
	}

	item.stash = "stash@{Tue Jan 14 19:13:06 2025}"
	if s := item.getStashTimestamp(); s != expected {
		t.Errorf("not equal, got %s", s)
	}

	item.stash = ""
	if s := item.getStashTimestamp(); s != "" {
		t.Errorf("not equal, got %s", s)
	}
}
