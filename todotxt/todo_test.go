// © 2012 Ethan Burns under the MIT license.

package todotxt

import (
	"reflect"
	"sort"
	"testing"
	"time"
)

func TestMakeTask(t *testing.T) {
	tests := []struct {
		text                 string
		done                 bool
		prio                 rune
		doneDate, createDate time.Time
	}{
		{"", false, rune(0), time.Time{}, time.Time{}},
		{"x ", true, rune(0), time.Time{}, time.Time{}},
		{"x 2012-12-23", true, rune(0), d(2012, time.December, 23), time.Time{}},
		{"x 2012-12-23 (A)", true, 'A', d(2012, time.December, 23), time.Time{}},
		{"x 2012-12-23 (A) 2012-12-20", true, 'A', d(2012, time.December, 23), d(2012, time.December, 20)},
		{"2012-12-23 (A) 2012-12-20", false, rune(0), time.Time{}, d(2012, time.December, 23)},
		{"x (A) 2012-12-20", true, 'A', time.Time{}, d(2012, time.December, 20)},
		{"x 2012-12-23 2012-12-20", true, rune(0), d(2012, time.December, 23), d(2012, time.December, 20)},
		{"x\n2012-12-23", true, rune(0), d(2012, time.December, 23), time.Time{}},
		{"x\r\n2012-12-23", true, rune(0), d(2012, time.December, 23), time.Time{}},
		{"x\r\n2012-12-23\n2012-12-20", true, rune(0), d(2012, time.December, 23), d(2012, time.December, 20)},
	}
	for _, test := range tests {
		task := MakeTask(test.text)
		if task.done != test.done {
			t.Errorf("Text [%s] expected done %t, got %t", test.text, test.done, task.done)
		}
		if task.prio != test.prio {
			t.Errorf("Text [%s] expected prio %s, got %s", test.text, test.prio, task.prio)
		}
		if !task.doneDate.Equal(test.doneDate) {
			t.Errorf("Text [%s] expected doneDate %s, got %s", test.text, test.doneDate, task.doneDate)
		}
		if !task.createDate.Equal(test.createDate) {
			t.Errorf("Text [%s] expected createDate %s, got %s", test.text, test.createDate, task.createDate)
		}
	}
}

func d(year int, month time.Month, day int) time.Time {
	return time.Date(year, month, day, 0, 0, 0, 0, time.FixedZone("UTC", 0))
}

func TestTags(t *testing.T) {
	tests := []struct {
		text   string
		marker rune
		tags   []string
	}{
		{"", '+', nil},
		{"+foo +bar", '+', []string{"+foo", "+bar"}},
		{"@foo @bar", '@', []string{"@foo", "@bar"}},
		{"hello +foo there +bar", '+', []string{"+foo", "+bar"}},
		{"hello @foo there @bar", '@', []string{"@foo", "@bar"}},
		{"+foo+ +bar", '+', []string{"+bar"}},
	}
	for _, test := range tests {
		task := MakeTask(test.text)
		tags := task.Tags(test.marker)
		sort.Strings(tags)
		sort.Strings(test.tags)
		if !reflect.DeepEqual(tags, test.tags) {
			t.Errorf("Text [%s], marker %c expected %v, got %v", test.text, test.marker, test.tags, tags)
		}
	}
}

func TestKeywords(t *testing.T) {
	tests := []struct {
		text string
		kwds map[string]string
	}{
		{"", map[string]string{}},
		{"due:2012-12-23", map[string]string{
			"due": "2012-12-23",
		}},
		{"due:2012-12-23 due:2012-12-24", map[string]string{
			"due": "2012-12-24",
		}},
		{"foo:bar baz:zap", map[string]string{
			"foo": "bar",
			"baz": "zap",
		}},
	}
	for _, test := range tests {
		task := MakeTask(test.text)
		kwds := task.Keywords()
		if len(kwds) != len(test.kwds) {
			t.Errorf("Text [%s] expected %d keywords, got %d", test.text, len(test.kwds), len(kwds))
		}
		for key, val := range kwds {
			if v, ok := test.kwds[key]; !ok {
				t.Errorf("Text [%s] unexpected keyword %s", test.text, key)
			} else if v != val {
				t.Errorf("Text [%s] expected %s:%s, got %s:%s", test.text, key, v, key, val)
			}
		}
	}
}

func TestComplete(t *testing.T) {
	today := time.Now().Format(DateFormat)
	tests := []struct {
		text, doneText string
	}{
		{"", "x " + today},
		{"x ", "x "},
		{"x", "x " + today + " x"}, // No space after initial x: not initially done.
		{"+foo +bar @baz", "x " + today + " +foo +bar @baz"},
	}
	for _, test := range tests {
		task := MakeTask(test.text)
		task.Complete()
		doneText := task.String()
		if doneText != test.doneText {
			t.Errorf("Text [%s], expected completed version to be [%s], got [%s]", test.text, test.doneText, doneText)
		}
	}
}
