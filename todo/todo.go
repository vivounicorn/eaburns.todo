// © 2012 Ethan Burns under the MIT license.

// Todo is a package for reading, modifying, and writing todo.txt files.
package todo

import (
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

// A File is a sequence of todo items.
type File []Item

// An Item represents a single line of a todo file.
type Item struct {
	Text string
	Priority priority
	Projects []string
	Contexts []string
	Done bool
	AddedDate, FinishDate time.Time
}

// timeFormat is the time format required by todo.txt.
const timeFormat = "2006-01-02 "

// ParseItem returns an item parsed from a line, or an error
// if the line is not a valid item.
func ParseItem(line string) *Item {
	i := &Item{ Text: line }

	if len(line) >= 2 && line[0] == 'x' && line[1] == ' ' {
		i.Done = true
		line = line[2:]
		i.FinishDate, line = parseDate(line)
	}

	i.Priority, line = parsePriority(line)
	i.AddedDate, line = parseDate(line)
	i.Projects = findTags('+', line)
	i.Contexts = findTags('@', line)
	return i
}

// ParseDate returns a date parsed from the beginning of the string
// and the remainder of the string.  If the string does not begin with
// a valid date then the zero time is returned along with the input string.
func parseDate(s string) (time.Time, string) {
	if len(s) < len(timeFormat) {
		return time.Time{}, s
	}
	t, err := time.Parse(timeFormat, s[:len(timeFormat)])
	if err != nil {
		return time.Time{}, s
	}
	return t, s[len(timeFormat):]
}

// ParsePriority returns a priority value from the beginning of the string
// and the remainder of the string.  If there is no priority value at the
// beginning of the string, then PriorityNone is returned along with the
// input string.
func parsePriority(s string) (priority, string) {
	if len(s) < 4 || s[0] != '(' || !strings.ContainsRune(prioRunes, rune(s[1])) || s[2] != ')' || s[3] != ' ' {
		return PriorityNone, s
	}
	return priority(s[:3]), s[4:]	// discard the space
}

// FindTags returns all tags in the string that begin with the given marker.
func findTags(marker rune, s string) (tags []string) {
	for {
		start := strings.IndexRune(s, marker)
		if start < 0 {
			break
		}
		s = s[start:]

		end := strings.IndexFunc(s, unicode.IsSpace)
		if end < 0 {
			end = len(s)
		}

		tag := s[:end]
		if len(tag) > 1 {
			l, _ := utf8.DecodeLastRuneInString(tag)
			if unicode.IsDigit(l) || unicode.IsLetter(l) || l == '_' {
				tags = append(tags, tag)
			}
		}

		s = s[1:]
	}
	return
}

// String returns the string representation of an Item.
func (i *Item) String() string {
	return i.Text
}

// Complete marks a task as completed.
func (i *Item) Complete() {
	i.Done = true
	i.FinishDate = time.Now()
}

// A priority is a letter A-Z in parens, representing the priority of a task.
type priority string

// PrioRunes are the valid priority runes.
const prioRunes = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"

// Priority values.
const (
	PriorityNone priority = ""
	PriorityA priority = "(A)"
	PriorityB priority = "(B)"
	PriorityC priority = "(C)"
	PriorityD priority = "(D)"
	PriorityE priority = "(E)"
	PriorityF priority = "(F)"
	PriorityG priority = "(G)"
	PriorityH priority = "(H)"
	PriorityI priority = "(I)"
	PriorityJ priority = "(J)"
	PriorityK priority = "(K)"
	PriorityL priority = "(L)"
	PriorityM priority = "(M)"
	PriorityN priority = "(N)"
	PriorityO priority = "(O)"
	PriorityP priority = "(P)"
	PriorityQ priority = "(Q)"
	PriorityR priority = "(R)"
	PriorityS priority = "(S)"
	PriorityT priority = "(T)"
	PriorityU priority = "(U)"
	PriorityV priority = "(V)"
	PriorityW priority = "(W)"
	PriorityX priority = "(X)"
	PriorityY priority = "(Y)"
	PriorityZ priority = "(Z)"
)

// String returns the string representation of the priority.
func (p priority) String() string {
	return string(p)
}
