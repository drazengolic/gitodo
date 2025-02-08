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

package wordwrap

import "strings"

// WrapText wraps a text in a string to the given width in characters
// with the optional "glue" between the lines, similar to strings.Join.
//
// Glue string is not calculated against the line width.
//
// Returns the text unmodified if the length is less or equal than width,
// or if zero is passed as a value of width. Panics if the width is less than zero.
func WrapText(text string, width int, glue string) string {
	if width < 0 {
		panic("the width is less than 0")
	}

	if width == 0 || len([]rune(text)) <= width {
		return text
	}

	wordBuilder := strings.Builder{}
	textBuilder := strings.Builder{}
	textBuilder.Grow(len(text))

	wordSize, lineSize, spaceSize := 0, 0, 0
	reader := strings.NewReader(text)

	for {
		r, b, err := reader.ReadRune()
		if err != nil || b == 0 {
			break
		}

		switch r {
		case '\r':
			continue
		case '\n':
			if lineSize+wordSize >= width {
				textBuilder.WriteRune('\n')
				textBuilder.WriteString(glue)
			} else {
				textBuilder.WriteString(strings.Repeat(" ", spaceSize))
			}

			textBuilder.WriteString(wordBuilder.String())
			textBuilder.WriteRune('\n')
			textBuilder.WriteString(glue)
			wordBuilder.Reset()
			wordSize, lineSize, spaceSize = 0, 0, 0
		case '\t':
			w := wordBuilder.String()
			switch {
			case lineSize+wordSize > width:
				textBuilder.WriteRune('\n')
				textBuilder.WriteString(glue)
				textBuilder.WriteString(w)
				lineSize = wordSize
				spaceSize = 4
			case lineSize+spaceSize+wordSize == width:
				textBuilder.WriteString(strings.Repeat(" ", spaceSize))
				textBuilder.WriteString(w)
				textBuilder.WriteRune('\n')
				textBuilder.WriteString(glue)
				lineSize = 0
				spaceSize = 0
			default:
				textBuilder.WriteString(strings.Repeat(" ", spaceSize))
				textBuilder.WriteString(w)
				spaceSize = 4
				lineSize += wordSize + spaceSize
			}
			wordBuilder.Reset()
			wordSize = 0
		case ' ':
			w := wordBuilder.String()
			switch {
			case lineSize+wordSize > width:
				textBuilder.WriteRune('\n')
				textBuilder.WriteString(glue)
				textBuilder.WriteString(w)
				lineSize = wordSize
				spaceSize = 1
			case lineSize+spaceSize+wordSize == width:
				textBuilder.WriteString(strings.Repeat(" ", spaceSize))
				textBuilder.WriteString(w)
				textBuilder.WriteRune('\n')
				textBuilder.WriteString(glue)
				lineSize = 0
				spaceSize = 0
			default:
				textBuilder.WriteString(strings.Repeat(" ", spaceSize))
				textBuilder.WriteString(w)
				spaceSize = 1
				lineSize += wordSize + spaceSize
			}
			wordBuilder.Reset()
			wordSize = 0
		default:
			if wordSize == width {
				textBuilder.WriteString(wordBuilder.String())
				textBuilder.WriteRune('\n')
				textBuilder.WriteString(glue)
				wordBuilder.Reset()
				wordBuilder.WriteRune(r)
				lineSize = 0
				wordSize = 1
			} else {
				wordBuilder.WriteRune(r)
				wordSize++
			}
		}
	}

	if lineSize+wordSize+spaceSize > width {
		textBuilder.WriteRune('\n')
		textBuilder.WriteString(glue)
		textBuilder.WriteString(wordBuilder.String())
	} else if wordSize > 0 {
		textBuilder.WriteString(wordBuilder.String())
	}

	return textBuilder.String()
}
