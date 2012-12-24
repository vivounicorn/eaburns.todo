// © 2012 Ethan Burns under the MIT license.

package todo

import (
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

const (
	// ProjectMarker is the rune that begins a project name.
	projectMarker = '+'

	// ContextMarker is the rune that begins a context name.
	contextMarker = '@'

	// KeywordSep is the rune separating a keyword/value binding.
	keywordSep = ':'

	// PrioRunes is a string of all valid priority runes.
	prioRunes = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"

	// DateFormat is the format string for dates.
	dateFormat = "2006-01-02"
)

// An Item is a single line of a todo.txt file.
type Item struct {
	LineNo int
	Text   string
}

// Done returns true if the item is marked as done.
func (item *Item) Done() bool {
	d, _ := parseDone(item.Text)
	return d
}

// Priority returns the priority string for this item.
func (item *Item) Priority() string {
	_, prio, _, _ := item.Header()
	return prio
}

// Header returns the header information from the item.
func (item *Item) Header() (done bool, prio string, doneDate, addDate time.Time) {
	text := item.Text
	done, text = parseDone(text)
	if done {
		doneDate, text = parseDate(text)
	}
	prio, text = parsePriority(text)
	addDate, _ = parseDate(text)
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
	if len(s) < len(dateFormat) {
		return time.Time{}, s
	}
	t, err := time.Parse(dateFormat, s[:len(dateFormat)])
	if err != nil {
		return time.Time{}, s
	}
	s = s[len(dateFormat):]
	if len(s) >= 1 && s[0] == ' ' {
		s = s[1:]
	}
	return t, s
}

// ParsePriority parses a priority value from the string and returns it and
// the rest of the string. If the string doesn't begin with a priority then
// an empty string.
func parsePriority(s string) (string, string) {
	if len(s) < 3 || s[0] != '(' || !strings.ContainsRune(prioRunes, rune(s[1])) || s[2] != ')' {
		return "", s
	}
	prio := s[1:2]
	s = s[3:]
	if len(s) >= 1 && s[0] == ' ' {
		s = s[1:]
	}
	return prio, s
}

// Projects returns a slice of all projects for this item.
func (item *Item) Projects() []string {
	return item.tags(projectMarker)
}

// Contexts returns a slice of all contexts for this item.
func (item *Item) Contexts() []string {
	return item.tags(contextMarker)
}

// Tags returns all tags in the text of the item that begin with the
// given marker.
func (item *Item) tags(marker rune) []string {
	var tags []string
	for _, f := range strings.Fields(item.Text) {
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

// Keywords returns a mapping of all <keyword>:<value> pairs in this item.
func (item *Item) Keywords() map[string]string {
	kwds := make(map[string]string)
	for _, f := range strings.Fields(item.Text) {
		i := strings.IndexRune(f, keywordSep)
		if i < 0 {
			continue
		}
		kwds[f[:i]] = f[i+1:]
	}
	return kwds
}
