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

package wordwrap

import "testing"

func TestWrapText(t *testing.T) {
	text := `This is a potentially long text
that needs to be fit into a specific length. It should also support padding on the left after the line breaks only.`

	exp1 := `This is a potentially long text
that needs to be fit into a specific length. It should
also support padding on the left after the line breaks
only.`

	exp2 := `This is a
potentially long
text
that needs to be fit
into a specific
length. It should
also support padding
on the left after
the line breaks
only.`

	exp3 := `This is a
	potentially long
	text
	that needs to be fit
	into a specific
	length. It should
	also support padding
	on the left after
	the line breaks
	only.`

	if s := WrapText(text, 54, ""); s != exp1 {
		t.Errorf("not equal.\nwant: %#v\n got: %#v", exp1, s)
	}

	if s := WrapText(text, 20, ""); s != exp2 {
		t.Errorf("not equal.\nwant: %#v\n got: %#v", exp2, s)
	}

	if s := WrapText(text, 20, "\t"); s != exp3 {
		t.Errorf("not equal.\nwant: %#v\n got: %#v", exp3, s)
	}

	if s := WrapText("", 20, ""); s != "" {
		t.Errorf("not empty")
	}

	exp4 := "text that:\n    has an important\npoint\n"

	if s := WrapText("text that:\n\thas an important point\n", 20, ""); s != exp4 {
		t.Errorf("not equal.\nwant: %#v\n got: %#v", exp4, s)
	}

	exp5 := "abcde\nfg"

	if s := WrapText("abcdefg", 5, ""); s != exp5 {
		t.Errorf("not equal.\nwant: %#v\n got: %#v", exp5, s)
	}
}

func TestWrapTextPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("no panic detected")
		}
	}()

	WrapText("text", -1, "")
}
