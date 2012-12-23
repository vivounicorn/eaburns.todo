package todo

import (
	"reflect"
	"sort"
	"testing"
	"time"
)

func TestParseTime(t *testing.T) {
	tests := []struct{
		text string
		time time.Time
		rest string
	}{
		{ "", time.Time{}, "" },
		{ "Foo bar", time.Time{}, "Foo bar" },
		{ "2012-12-23 ", d(2012, time.December, 23), "" },
		{ "2012-12-23 Hello", d(2012, time.December, 23), "Hello" },
	}
	for _, test := range tests {
		time, rest := parseDate(test.text)
		if !time.Equal(test.time) {
			t.Errorf("Text [%s], expected time [%s], got [%s]", test.text, test.time, time)
		}
		if rest != test.rest {
			t.Errorf("Text [%s], expected rest [%s], got [%s]", test.text, test.rest, rest)
		}
	}
}

func TestParsePriority(t *testing.T) {
	tests := []struct{
		text, prio, rest string
	}{
		{ "(A) ", "(A)", "" },
		{ "(a) ", "", "(a) " },
		{ "(B) Hello", "(B)", "Hello" },
		{ "(B", "", "(B" },
		{ "", "", "" },
		{ "Really gotta call Mom (A) @phone @someday", "", "Really gotta call Mom (A) @phone @someday" },
		{ "(b) Get back to the boss", "", "(b) Get back to the boss" },
		{ "(B)->Submit TPS report", "", "(B)->Submit TPS report" },
	}
	for _, test := range tests {
		prio, rest := parsePriority(test.text)
		if string(prio) != test.prio {
			t.Errorf("Text [%s], expected priority [%s], got [%s]", test.text, test.prio, prio)
		}
		if rest != test.rest {
			t.Errorf("Text [%s], expected rest [%s], got [%s]", test.text, test.rest, rest)
		}
	}
}

func TestFindTags(t *testing.T) {
	tests := []struct{
		text string
		marker rune
		tags []string
	}{
		{ "", '+', nil },
		{ "(A) Call Mom +Family +PeaceLoveAndHappiness @iphone @phone", '+', []string{"+Family", "+PeaceLoveAndHappiness"} },
		{ "(A) Call Mom +Family +PeaceLoveAndHappiness @iphone @phone", '@', []string{"@iphone", "@phone"} },
		{ "+foo+bar", '+', []string{ "+foo+bar", "+bar" } },
		{ "+foo+bar+", '+', nil },
		{ "++foo+bar", '+', []string{ "++foo+bar", "+foo+bar", "+bar" } },
	}

	for _, test := range tests {
		tags := findTags(test.marker, test.text)
		sort.Strings(test.tags)
		sort.Strings(tags)
		if !reflect.DeepEqual(test.tags, tags) {
			t.Errorf("Text [%s], expected tags %v, got %v", test.text, test.tags, tags)
		}
	}
}

func TestParseItem(t *testing.T) {
	tests := []struct{
		text string
		// The item's text is filled in automatically by the test harness,
		// so just leave it empty.
		item Item
	}{
		{ "xhello", Item{} },
		{ "x hello", Item{Done: true} },

		{ "x 2012-12-23 hello", Item{
			Done: true,
			FinishDate: d(2012, time.December, 23),
		} },

		{ "x 2012-12-23 2012-12-20 hello", Item{
			Done: true,
			AddedDate: d(2012, time.December, 20),
			FinishDate: d(2012, time.December, 23),
		} },

		{ "2012-12-23 2012-12-20 hello", Item{
			AddedDate: d(2012, time.December, 23),
		} },

		{ "(A) 2012-12-23 hello", Item{
			Priority: PriorityA,
			AddedDate: d(2012, time.December, 23),
		} },

		{ "(A) 2012-12-23 +hello", Item{
			Priority: PriorityA,
			Projects: []string{"+hello"},
			AddedDate: d(2012, time.December, 23),
		} },

		{ "(A) +hello", Item{
			Priority: PriorityA,
			Projects: []string{"+hello"},
		} },

		{ "(A) +hello @goodbye Hi there", Item{
			Priority: PriorityA,
			Contexts: []string{"@goodbye"},
			Projects: []string{"+hello"},
		} },

		{ "(A) +hello +goodbye Hi there", Item{
			Priority: PriorityA,
			Projects: []string{"+hello", "+goodbye"},
		} },

		{ "x 2011-03-02 2011-03-01 Review Tim's pull request +TodoTxtTouch @github", Item {
			Done: true,
			FinishDate: d(2011, time.March, 2),
			AddedDate: d(2011, time.March, 1),
			Projects: []string{"+TodoTxtTouch"},
			Contexts: []string{ "@github" },
		} },

		{ "(A) Call Mom +Family +PeaceLoveAndHappiness @iphone @phone", Item{
			Priority: PriorityA,
			Projects: []string{"+Family", "+PeaceLoveAndHappiness"},
			Contexts: []string{ "@iphone", "@phone"},
		} },
	}
	for _, test := range tests {
		test.item.Text = test.text
		item := ParseItem(test.text)
		if !itemEq(item, &test.item) {
			t.Errorf("Text [%s], expected\n%#v,\ngot\n%#v", test.text, test.item, *item)
		}
	}
}

// ItemEq returns true if the two items are equal.
func itemEq(a, b *Item) bool {
	sort.Strings(a.Contexts)
	sort.Strings(b.Contexts)
	sort.Strings(a.Projects)
	sort.Strings(b.Projects)
	return a.Text == b.Text &&
		a.Priority == b.Priority &&
		reflect.DeepEqual(a.Projects, b.Projects) &&
		reflect.DeepEqual(a.Contexts, b.Contexts) &&
		a.Done == b.Done &&
		a.AddedDate.Equal(b.AddedDate) &&
		a.FinishDate.Equal(b.FinishDate)
}

// D returns a date.
func d(year int, month time.Month, day int) time.Time {
	return time.Date(year, month, day, 0, 0, 0, 0, time.FixedZone("UTC", 0))
}