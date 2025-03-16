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
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

type StashItem struct {
	Date, Ref string
}

func parseStashList(content string) map[int]StashItem {
	result := make(map[int]StashItem)
	regName := regexp.MustCompile("gitodo_([0-9]+)")
	regDate := regexp.MustCompile(`\{([^}]+)\}`)

	for i, row := range strings.Split(content, "\n") {
		m := regName.FindStringSubmatch(row)
		if len(m) == 2 {
			todoId, _ := strconv.Atoi(m[1])
			date := regDate.FindStringSubmatch(row)
			item := StashItem{Ref: "stash@{" + strconv.Itoa(i) + "}"}
			if len(date) == 2 {
				item.Date = date[1]
			}
			result[todoId] = item
		}
	}

	return result
}

func GetStashItems() (map[int]StashItem, error) {
	revOutput, err := exec.Command("git", "--no-pager", "stash", "list", "--date=local").Output()
	if err != nil {
		return nil, err
	}
	return parseStashList(string(revOutput)), nil
}

func PushStash(todoId int) error {
	_, err := exec.Command("git", "stash", "push", "-m", "gitodo_"+strconv.Itoa(todoId), "--include-untracked").Output()
	return err
}

func PushStashNoItem() (string, error) {
	out, err := exec.Command("git", "stash", "--include-untracked").Output()
	return string(out), err
}

func PopStash(stash string) error {
	_, err := exec.Command("git", "stash", "pop", stash).Output()
	return err
}
