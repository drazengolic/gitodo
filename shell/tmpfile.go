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
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"unicode"
)

const tmpFilePref = "gitodo_"

type TmpFile struct {
	path string
}

func NewTmpFile(content io.WriterTo) (*TmpFile, error) {
	f, err := os.CreateTemp("", tmpFilePref)

	if err != nil {
		return nil, err
	}

	content.WriteTo(f)
	f.Close()

	return &TmpFile{path: filepath.Join(os.TempDir(), f.Name())}, nil
}

func NewTmpFileString(content string) (*TmpFile, error) {
	f, err := os.CreateTemp("", tmpFilePref)

	if err != nil {
		return nil, err
	}

	_, err = f.WriteString(content)
	if err != nil {
		return nil, err
	}

	f.Close()

	return &TmpFile{path: f.Name()}, nil
}

func (tf *TmpFile) Delete() error {
	return os.Remove(tf.path)
}

func (tf *TmpFile) Edit(editor string, startAt int) error {
	editorSplit := strings.Split(editor, " ")
	editor = editorSplit[0]
	var args []string

	if len(editorSplit) > 1 {
		if editor == "subl" && startAt > 0 {
			args = append(editorSplit[1:], fmt.Sprint(tf.path, ":", startAt))
		} else {
			args = append(editorSplit[1:], tf.path)
		}
	} else {
		switch editor {
		case "vi", "vim", "nvim":
			if startAt > 0 {
				args = []string{"+normal Ga", tf.path}
			} else {
				args = []string{tf.path}
			}
		case "nano", "emacs":
			if startAt > 0 {
				args = []string{fmt.Sprint("+", startAt), tf.path}
			} else {
				args = []string{tf.path}
			}
		case "subl":
			if startAt > 0 {
				args = []string{"-n", "-w", fmt.Sprint(tf.path, ":", startAt)}
			} else {
				args = []string{"-n", "-w", tf.path}
			}
		default:
			args = []string{tf.path}
		}
	}
	edit := exec.Command(editor, args...)
	edit.Env = os.Environ()
	edit.Stdin = os.Stdin
	edit.Stdout = os.Stdout
	edit.Stderr = os.Stderr
	return edit.Run()
}

func (tf *TmpFile) ReadAll() string {
	bytes, err := os.ReadFile(tf.path)
	if err != nil {
		return ""
	} else {
		return string(bytes)
	}
}

func (tf *TmpFile) ReadItems() ([]string, error) {
	items := []string{}
	file, err := os.Open(tf.path)

	if err != nil {
		return nil, err
	}
	defer file.Close()

	builder := strings.Builder{}
	scanner := bufio.NewScanner(file)

	var txt, item string
	for scanner.Scan() {
		txt = scanner.Text()
		switch {
		case txt != "" && txt[0] == '#':
			continue
		case (txt != "" && txt[0] == '-') && builder.Len() > 0:
			item = strings.TrimSpace(builder.String())
			if item != "" {
				items = append(items, item)
			}
			builder.Reset()
			fallthrough
		case txt != "" && txt[0] == '-':
			builder.WriteString(strings.TrimLeftFunc(
				txt, func(r rune) bool { return r == '-' || unicode.IsSpace(r) },
			))
			builder.WriteRune('\n')
		default:
			builder.WriteString(txt)
			builder.WriteRune('\n')
		}
	}

	if builder.Len() > 0 {
		item = strings.TrimSpace(builder.String())
	} else {
		item = ""
	}

	if item != "" {
		items = append(items, item)
	}

	return items, nil
}

func (tf *TmpFile) Path() string {
	return tf.path
}
