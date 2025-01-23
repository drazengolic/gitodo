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

import (
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

func ParseStashList(content string) map[int]string {
	res := make(map[int]string)
	r := regexp.MustCompile("gitodo_([0-9]+)")

	for _, row := range strings.Split(content, "\n") {
		m := r.FindStringSubmatch(row)
		if len(m) == 2 {
			id, _ := strconv.Atoi(m[1])
			val, _, _ := strings.Cut(row, "}:")
			res[id] = val + "}"
		}
	}

	return res
}

func GetStashItems() (map[int]string, error) {
	revOutput, err := exec.Command("git", "--no-pager", "stash", "list", "--date=local").Output()
	if err != nil {
		return nil, err
	}
	return ParseStashList(string(revOutput)), nil
}

func PushStash(todoId int) error {
	_, err := exec.Command("git", "stash", "push", "-m", "gitodo_"+strconv.Itoa(todoId), "--include-untracked").Output()
	return err
}

func PopStash(stash string) error {
	_, err := exec.Command("git", "stash", "pop", stash).Output()
	return err
}
