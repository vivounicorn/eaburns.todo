// © 2012 Ethan Burns under the MIT license.

// Todotxt is a package for using todo.txt formatted files.
// See http://todotxt.com/ for more details.
package todotxt

import (
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

const (
	// ProjectMarker is the rune that begins a project name.
	ProjectMarker = '+'

	// ContextMarker is the rune that begins a context name.
	PontextMarker = '@'

	// KeywordSep is the rune separating a keyword/value binding.
	KeywordSep = ':'

	// PrioRunes is a string of all valid priority runes.
	PrioRunes = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"

	// DateFormat is the format string for dates.
	DateFormat = "2006-01-02"
)

// A Task is a single line of a todo.txt file.
type Task struct {
	text string
}

// MakeTask returns a task for the given text.  If the text contains
// newlines then they are interpreted as space characters (' ').
func MakeTask(text string) Task {
	text = strings.Replace(text, "\r\n", " ", -1)
	text = strings.Replace(text, "\n", " ", -1)
	return Task{text}
}

// Done returns true if the task is marked as done, otherwise false.
// If the task is done and has a completion date than that is returned
// as the second argument, otherwise the second argument is the
// zero time.
func (t *Task) Done() (bool, time.Time) {
	d, doneDate, _, _ := t.header()
	return d, doneDate
}

// Priority returns the priority string for this task.
func (t *Task) Priority() string {
	_, _, prio, _ := t.header()
	return prio
}

// CreationDate returns the creation date for this task, if it does not
// have a creation date than the zero time is returned.
func (t *Task) CreationDate() time.Time {
	_, _, _, addDate := t.header()
	return addDate
}

// Heaheaderder returns the header information from the task.
func (t *Task) header() (d bool, dDate time.Time, p string, cDate time.Time) {
	txt := t.text
	d, txt = parseDone(txt)
	if d {
		dDate, txt = parseDate(txt)
	}
	p, txt = parsePriority(txt)
	cDate, _ = parseDate(txt)
	return
}

// ParseDone returns the completed status from the string and the
// rest of the string after it.
func parseDone(s string) (bool, string) {
	if len(s) >= 2 && s[0] == 'x' && s[1] == ' ' {
		return true, s[2:]
	}
	return false, s
}

// ParseDate returns the time from the beginning of the string, or the
// zero time if the string doesn't begin with a time, and the rest of the
// string.
func parseDate(s string) (time.Time, string) {
	if len(s) < len(DateFormat) {
		return time.Time{}, s
	}
	t, err := time.Parse(DateFormat, s[:len(DateFormat)])
	if err != nil {
		return time.Time{}, s
	}
	s = s[len(DateFormat):]
	if len(s) >= 1 && s[0] == ' ' {
		s = s[1:]
	}
	return t, s
}

// ParsePriority parses a priority value from the string and returns it and
// the rest of the string. If the string doesn't begin with a priority then
// an empty string.
func parsePriority(s string) (string, string) {
	if len(s) < 3 || s[0] != '(' || !strings.ContainsRune(PrioRunes, rune(s[1])) || s[2] != ')' {
		return "", s
	}
	prio := s[1:2]
	s = s[3:]
	if len(s) >= 1 && s[0] == ' ' {
		s = s[1:]
	}
	return prio, s
}

// Tags returns all tag with the given marker rune.
// A tag is a white-space delienated field that begins with a marker
// rune and ends with an alphanumeric or '_' rune.
// Projects are tags that begin with '+'.
// Contexts are tags that begin with '@'.
func (t *Task) Tags(marker rune) []string {
	var tags []string
	for _, f := range strings.Fields(t.text) {
		if first, _ := utf8.DecodeRuneInString(f); first != marker {
			continue
		}
		if last, _ := utf8.DecodeLastRuneInString(f); !tagEnd(last) {
			continue
		}
		tags = append(tags, f)
	}
	return tags
}

// TagEnd returns true for runes that are valid tag ends.
func tagEnd(r rune) bool {
	return unicode.IsDigit(r) || unicode.IsLetter(r) || r == '_'
}

// Keywords returns a mapping of <keyword>:<value> pairs in this task.
// If there are multiple assignments to the same keyword then only the
// last one is returned.
func (t *Task) Keywords() map[string]string {
	kwds := make(map[string]string)
	for _, f := range strings.Fields(t.text) {
		i := strings.IndexRune(f, KeywordSep)
		if i < 0 {
			continue
		}
		kwds[f[:i]] = f[i+1:]
	}
	return kwds
}
