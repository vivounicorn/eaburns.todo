// © 2012 Ethan Burns under the MIT license.

package todotxt

import (
	"reflect"
	"sort"
	"testing"
	"time"
)

func TestHeader(t *testing.T) {
	tests := []struct {
		text              string
		done              bool
		prio              string
		doneDate, addDate time.Time
	}{
		{"", false, "", time.Time{}, time.Time{}},
		{"x ", true, "", time.Time{}, time.Time{}},
		{"x 2012-12-23", true, "", d(2012, time.December, 23), time.Time{}},
		{"x 2012-12-23 (A)", true, "A", d(2012, time.December, 23), time.Time{}},
		{"x 2012-12-23 (A) 2012-12-20", true, "A", d(2012, time.December, 23), d(2012, time.December, 20)},
		{"2012-12-23 (A) 2012-12-20", false, "", time.Time{}, d(2012, time.December, 23)},
		{"x (A) 2012-12-20", true, "A", time.Time{}, d(2012, time.December, 20)},
		{"x 2012-12-23 2012-12-20", true, "", d(2012, time.December, 23), d(2012, time.December, 20)},
	}
	for _, test := range tests {
		task := &Task{text: test.text}
		done, doneDate, prio, addDate := task.header()
		if done != test.done {
			t.Errorf("Text [%s] expected done %t, got %t", test.text, test.done, done)
		}
		if prio != test.prio {
			t.Errorf("Text [%s] expected prio %s, got %s", test.text, test.prio, prio)
		}
		if !doneDate.Equal(test.doneDate) {
			t.Errorf("Text [%s] expected doneDate %s, got %s", test.text, test.doneDate, doneDate)
		}
		if !addDate.Equal(test.addDate) {
			t.Errorf("Text [%s] expected addDate %s, got %s", test.text, test.addDate, addDate)
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
		task := &Task{text: test.text}
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
