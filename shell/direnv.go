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
	"errors"
	"os/exec"
	"strings"
)

type DirEnv struct {
	ProjDir, Branch, Editor string
}

func GetDirEnv() (*DirEnv, error) {
	revOutput, err := exec.Command("git", "rev-parse", "--show-toplevel", "--abbrev-ref", "HEAD").Output()
	if err != nil {
		switch e := err.(type) {
		case *exec.Error:
			return nil, errors.New("git must be installed in order to use the application.")
		case *exec.ExitError:
			return nil, errors.New(string(e.Stderr))
		default:
			return nil, err
		}
	}

	revArgs := strings.Split(strings.TrimSpace(string(revOutput)), "\n")
	editor, _ := exec.Command("git", "var", "GIT_EDITOR").Output()

	return &DirEnv{
		ProjDir: revArgs[0],
		Branch:  revArgs[1],
		Editor:  strings.TrimSpace(string(editor)),
	}, nil
}

func ListBranches() ([]string, error) {
	out, err := exec.Command("git", "--no-pager", "branch", "--format=%(refname:short)").Output()
	if err != nil {
		return nil, err
	}
	return strings.Split(strings.TrimSpace(string(out)), "\n"), nil
}
