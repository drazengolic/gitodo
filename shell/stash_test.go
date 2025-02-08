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

package shell

import (
	"maps"
	"testing"
)

func TestParseStashList(t *testing.T) {
	sample := `stash@{Tue Jan 09 10:11:12 2025}: WIP on master: 04fd51c update docs
stash@{Tue Jan 14 19:13:06 2025}: On master: gitodo_7
stash@{Tue Jan 18 10:11:12 2025}: WIP on master: 04fd51c update docs`

	expected := map[int]string{7: "stash@{Tue Jan 14 19:13:06 2025}"}
	got := ParseStashList(sample)
	cmp := maps.Equal(expected, got)

	if !cmp {
		t.Errorf("maps not equal. got: %v", got)
	}
}
