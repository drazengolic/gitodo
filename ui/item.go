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

import (
	"fmt"
	"github.com/drazengolic/gitodo/shell"
	"github.com/drazengolic/gitodo/wordwrap"
)

// item rendered in both sections
type todoItem struct {
	id              int
	task            string
	done, committed bool
	stash           shell.StashItem
}

// Render renders a single to-do item
func (i todoItem) Render(bold, showId bool, width int, glue string) string {
	var s string
	switch {
	case bold && showId:
		s = boldText.Render(wordwrap.WrapText(fmt.Sprintf("[#%d] %s", i.id, i.task), width, glue))
	case bold:
		s = boldText.Render(wordwrap.WrapText(i.task, width, glue))
	case showId:
		s = wordwrap.WrapText(fmt.Sprintf("[#%d] %s", i.id, i.task), width, glue)
	default:
		s = wordwrap.WrapText(i.task, width, glue)
	}

	if i.committed {
		s += fmt.Sprintf("\n%s• %s", glue, committedBox)
	}
	if i.stash.Date != "" {
		s += fmt.Sprintf("\n%s• %s", glue, orangeText.Render("stashed: "+i.stash.Date))
	}
	return s
}
