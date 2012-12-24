// © 2012 Ethan Burns under the MIT license.

// Todotxt is a package for using todo.txt formatted files.
// See http://todotxt.com/ for more details.
package todotxt

import (
	"bufio"
	"io"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

const (
	// ProjectTag is the rune that begins a project name.
	ProjectTag = '+'

	// ContextTag is the rune that begins a context name.
	ContextTag = '@'

	// KeywordSep is the rune separating a keyword=value binding.
	KeywordSep = ':'

	// PriorityRunes is a string of all valid priority runes.
	PriorityRunes = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"

	// DateFormat is the format string for dates.
	DateFormat = "2006-01-02"
)

// File is a todo.txt file.
type File struct {
	Tasks []Task
}

// ReadFile reads a todo.txt file and returns it or an error.
// Read uses its own internal buffering.
func ReadFile(in io.Reader) (File, error) {
	var f File
	bufIn := bufio.NewReader(in)
	for {
		line, err := bufIn.ReadString('\n')
		if err != nil && err != io.EOF {
			return File{}, err
		}

		line = strings.TrimRight(line, "\r\n")
		f.Tasks = append(f.Tasks, MakeTask(line))

		if err == io.EOF {
			break
		}
	}
	return f, nil
}

// WriteTo writes the File to the given Writer, implementing
// the io.WriterTo interface.
func (f *File) WriteTo(out io.Writer) (int64, error) {
	var tot int64
	for _, t := range f.Tasks {
		n, err := io.WriteString(out, t.String()+"\n")
		tot += int64(n)
		if err != nil {
			return tot, err
		}
	}
	return tot, nil
}

// A Task is a single line of a todo.txt file.
type Task struct {
	text                 string
	fields               []string
	done                 bool
	prio                 rune
	createDate, doneDate time.Time
}

// MakeTask returns a task for the given text.  If the text contains
// newlines then they are interpreted as space characters (' ').
func MakeTask(text string) Task {
	text = strings.Replace(text, "\r\n", " ", -1)
	text = strings.Replace(text, "\n", " ", -1)

	t := Task{text: text, fields: strings.Fields(text)}

	t.done, text = parseDone(text)
	if t.done {
		t.doneDate, text = parseDate(text)
	}
	t.prio, text = parsePriority(text)
	t.createDate, _ = parseDate(text)

	return t
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
func parsePriority(s string) (rune, string) {
	if len(s) < 3 || s[0] != '(' || !strings.ContainsRune(PriorityRunes, rune(s[1])) || s[2] != ')' {
		return rune(0), s
	}
	prio := rune(s[1])
	s = s[3:]
	if len(s) >= 1 && s[0] == ' ' {
		s = s[1:]
	}
	return prio, s
}

// String returns the single-line string representation of this task.
func (t *Task) String() string {
	return t.text
}

// Priority returns the task's priority value rune or the zero rune if
// the task does not have a priority.
func (t *Task) Priority() rune {
	return t.prio
}

// Complete marks the task as complete.
func (t *Task) Complete() {
	if t.IsDone() {
		return
	}
	fmt := DateFormat
	if t.text != "" {
		fmt += " "
	}
	prefix := "x " + time.Now().Format(fmt)
	*t = MakeTask(prefix + t.text)
}

// IsDone returns true if this task is completed.
func (t *Task) IsDone() bool {
	return t.done
}

// CompletionDate returns the completion date if this task is done and
// included such a date.  Otherwise, the zero time is returned.
func (t *Task) CompletionDate() time.Time {
	return t.doneDate
}

// CreationDate returns the creation date if specified.  Otherwise, the
// zero time is returned.
func (t *Task) CreationDate() time.Time {
	return t.createDate
}

// Tags returns all tag with the given marker rune.
// A tag is a white-space delienated field that begins with a marker
// rune and ends with an alphanumeric or '_' rune.
// Projects are tags that begin with '+'.
// Contexts are tags that begin with '@'.
func (t *Task) Tags(marker rune) []string {
	var tags []string
	for _, f := range t.fields {
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
	for _, f := range t.fields {
		i := strings.IndexRune(f, KeywordSep)
		if i < 0 {
			continue
		}
		kwds[f[:i]] = f[i+1:]
	}
	return kwds
}
