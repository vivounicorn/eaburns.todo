// © 2012 Ethan Burns under the MIT license.

// Todo is a package for reading, modifying, and writing todo.txt files.
package todo

import (
	"sort"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

// PriorityRunes are the valid priority runes.
const PriorityRunes = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"

type priority rune

// NoPriority is the priority constant for no priority
const NoPriority priority = priority(0)

// Priority returns a priority value for the given rune.  If the rune is
// not a valid priority then it panics.
func Priority(r rune) priority {
	if !strings.ContainsRune(PriorityRunes, r) {
		panic("Invalid priority: " + string(r))
	}
	return priority(r)
}

// String returns the string representation of the priority.
func (p priority) String() string {
	if p == NoPriority {
		return ""
	}
	return "(" + string(p) + ")"
}

// A File is a sequence of todo items.
type File []Item

// An Item represents a single line of a todo file.
type Item struct {
	// Priority is the priority string for this item.
	Priority priority

	// Projects is a list of projects for this item without
	// the leading + marker.
	Projects []string

	// Contexts is a list of the contexts for this item without
	// the leading @ marker.
	Contexts []string

	// Done is true if this item is completed.
	Done bool

	// AddedDate and FinishedDate are the added and finished
	// dates of this item or they are the zero time if the item
	// does not have the corresponding date.
	AddedDate, FinishDate time.Time

	// Text is the text of the item (excluding completion info,
	// priority, and added date prefix) at the time that it was
	// parsed, or at the last call to Text().
	text string
}

// ParseItem returns an item parsed from a line, or an error
// if the line is not a valid item.
func ParseItem(line string) *Item {
	i := new(Item)
	i.Done, i.FinishDate, line = parseCompleted(line)
	i.Priority, line = parsePriority(line)
	i.AddedDate, line = parseDate(line)
	i.text = line
	i.Projects = findTags('+', line)
	i.Contexts = findTags('@', line)
	return i
}

// timeFormat is the time format required by todo.txt.
const timeFormat = "2006-01-02 "

// ParseCompleted returns the done status, finish time and the rest
// of the string.
func parseCompleted(s string) (bool, time.Time, string) {
	if len(s) >= 2 && s[0] == 'x' && s[1] == ' ' {
		t, s := parseDate(s[2:])
		return true, t, s
	}
	return false, time.Time{}, s
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
	if len(s) < 4 || s[0] != '(' || !strings.ContainsRune(PriorityRunes, rune(s[1])) || s[2] != ')' || s[3] != ' ' {
		return priority('\000'), s
	}
	return priority(s[1]), s[4:]	// discard the space
}

// FindTags returns all tags in the string that begin with the given marker.
func findTags(marker rune, s string) (tags []string) {
	for {
		start := strings.IndexRune(s, marker)
		if start < 0 {
			break
		}
		s = s[start+1:]

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
	}
	return
}

// Text returns the item text, which is the string representation
// of the item excluding any completed, priority, and date prefix.
func (i *Item) Text() string {
	i.text = reconsileTags(i.text, '+', i.Projects)
	i.text = reconsileTags(i.text, '@', i.Contexts)
	i.text = strings.TrimSpace(i.text)
	return i.text
}

// ReconsileTags returns a new copy of the given text with
// tags added if they are in the slice but not the text and tags
// removed if they are in the text but no the slice.
func reconsileTags(text string, marker rune, tags []string) string {
	inSlice := make(map[string]bool)
	for _, t := range tags {
		inSlice[t] = true
	}

	// Remove tags not in the slice, longest first.
	inText := make(map[string]bool)
	for _, t := range findTags(marker, text) {
		inText[t] = true
	}
	sort.Sort(longFirstSorter(tags))
	for t := range inText {
		if !inSlice[t] {
			text = rmTag(text, string(marker)+t)
		}
	}

	// See what tags are still in the text, and add all remaining ones
	inText = make(map[string]bool)
	for _, t := range findTags(marker, text) {
		inText[t] = true
	}
	for _, t := range tags {
		if !inText[t] {
			text = addTag(text, string(marker)+t)
		}
	}

	return text
}

type longFirstSorter []string

func (l longFirstSorter) Len() int {
	return len(l)
}

func (l longFirstSorter) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}

func (l longFirstSorter) Less(i, j int) bool {
	return len(l[i]) > len(l[j])
}

// AddTag adds the tag to the string, inserting a space if necessary.
func addTag(s, tag string) string {
	l, _ := utf8.DecodeLastRuneInString(s)
	if !unicode.IsSpace(l) {
		return s + " " + tag
	}
	return s + tag
}

// RmTag removes the tag from the string.  This is pretty inefficient.
func rmTag(s, tag string) string {
	rest := s
	s = ""
	space := true
	for len(rest) > 0 {
		if ok, sz := tagPrefix(rest, tag); ok {
			if !space {	// Not preceeded by a space, don't cut the space
				sz = len(tag)
			}
			rest = rest[sz:]
			continue
		}
		r, w := utf8.DecodeRuneInString(rest)
		s += string(r)
		space = unicode.IsSpace(r)
		rest = rest[w:]
	}
	return s
}

// TagPrefix returns true if the prefix is the tag.
func tagPrefix(s, tag string) (bool, int) {
	if !strings.HasPrefix(s, tag) {
		return false, 0
	}
	if len(s) == len(tag) {
		return true, len(s)
	}
	if r, w := utf8.DecodeRuneInString(s[len(tag):]); unicode.IsSpace(r) {
		return true, len(tag)+w
	}
	return false, 0
}

// String returns the string representation of an Item.
func (i *Item) String() string {
	s := ""
	if i.Done {
		s += "x "
		if !i.FinishDate.IsZero() {
			s += i.FinishDate.Format(timeFormat)
		}
	}
	if i.Priority != NoPriority {
		s += i.Priority.String() + " "
	}
	if !i.AddedDate.IsZero() {
		s += i.AddedDate.Format(timeFormat)
	}
	return strings.TrimSpace(s + i.Text())
}

