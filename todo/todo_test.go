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
		if prio.String() != test.prio {
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
		{ "(A) Call Mom +Family +PeaceLoveAndHappiness @iphone @phone", '+', []string{"Family", "PeaceLoveAndHappiness"} },
		{ "(A) Call Mom +Family +PeaceLoveAndHappiness @iphone @phone", '@', []string{"iphone", "phone"} },
		{ "+foo+bar", '+', []string{ "foo+bar", "bar" } },
		{ "+foo+bar+", '+', nil },
		{ "++foo+bar", '+', []string{ "+foo+bar", "foo+bar", "bar" } },
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
		item Item
	}{
		{ "xhello", Item{
			text: "xhello",
		} },

		{ "x hello", Item{
			text: "hello",
			Done: true,
		} },

		{ "x 2012-12-23 hello", Item{
			text: "hello",
			Done: true,
			FinishDate: d(2012, time.December, 23),
		} },

		{ "x 2012-12-23 2012-12-20 hello", Item{
			text: "hello",
			Done: true,
			AddedDate: d(2012, time.December, 20),
			FinishDate: d(2012, time.December, 23),
		} },

		{ "2012-12-23 2012-12-20 hello", Item{
			text: "2012-12-20 hello",
			AddedDate: d(2012, time.December, 23),
		} },

		{ "(A) 2012-12-23 hello", Item{
			text: "hello",
			Priority: Priority('A'),
			AddedDate: d(2012, time.December, 23),
		} },

		{ "(A) 2012-12-23 +hello", Item{
			text: "+hello",
			Priority: Priority('A'),
			Projects: []string{"hello"},
			AddedDate: d(2012, time.December, 23),
		} },

		{ "(A) +hello", Item{
			text: "+hello",
			Priority: Priority('A'),
			Projects: []string{"hello"},
		} },

		{ "(A) +hello @goodbye Hi there", Item{
			text: "+hello @goodbye Hi there",
			Priority: Priority('A'),
			Contexts: []string{"goodbye"},
			Projects: []string{"hello"},
		} },

		{ "(A) +hello +goodbye Hi there", Item{
			text: "+hello +goodbye Hi there",
			Priority: Priority('A'),
			Projects: []string{"hello", "goodbye"},
		} },

		{ "x 2011-03-02 2011-03-01 Review Tim's pull request +TodoTxtTouch @github", Item {
			text: "Review Tim's pull request +TodoTxtTouch @github",
			Done: true,
			FinishDate: d(2011, time.March, 2),
			AddedDate: d(2011, time.March, 1),
			Projects: []string{"TodoTxtTouch"},
			Contexts: []string{ "github" },
		} },

		{ "(A) Call Mom +Family +PeaceLoveAndHappiness @iphone @phone", Item{
			text: "Call Mom +Family +PeaceLoveAndHappiness @iphone @phone",
			Priority: Priority('A'),
			Projects: []string{"Family", "PeaceLoveAndHappiness"},
			Contexts: []string{ "iphone", "phone"},
		} },
	}
	for _, test := range tests {
		item := ParseItem(test.text)
		if !itemEq(item, &test.item) {
			t.Errorf("Text [%s], expected\n%#v,\ngot\n%#v", test.text, test.item, *item)
		}
	}
}

func TestRmTags(t *testing.T) {
	tests := []struct{
		text string
		tags []string
		result string
	}{
		{ "+foo bar", []string{"+foo"}, "bar"},
		{ "+foo +bar", []string{"+foo"}, "+bar"},
		{ "+foo +bar", []string{"+bar"}, "+foo "},
		{ "+foo+bar", []string{"+foo"}, "+foo+bar"},
		{ "foo+bar", []string{"+bar"}, "foo"},
	}
	for _, test := range tests {
		text := test.text
		for _, tag := range test.tags {
			text = rmTag(text, tag)
		}
		if text != test.result {
			t.Errorf("Text [%s], removing %v, expected [%s], got [%s]",
				test.text, test.tags, test.result, text)
		}
	}
}

func TestItem_TextAndString(t *testing.T) {
	tests := []struct{
		item Item
		str, txt string
	}{
		{
			Item{
				text: "hello",
				Priority: Priority('A'),
			},
			"(A) hello",
			"hello",
		},
		{
			Item{
				text: "hello +there",
				Priority: Priority('A'),
				Projects: []string{"there"},
			},
			"(A) hello +there",
			"hello +there",
		},
		{
			Item{
				text: "hello +there",
				Priority: Priority('A'),
				Projects: []string{"there", "foo"},
			},
			"(A) hello +there +foo",
			"hello +there +foo",
		},
		{
			Item{
				text: "hello +there ",
				Priority: Priority('A'),
				Projects: []string{"there", "foo"},
			},
			"(A) hello +there +foo",
			"hello +there +foo",
		},
		{
			Item{
				text: "hello @there",
				Priority: Priority('A'),
				Contexts: []string{"there"},
			},
			"(A) hello @there",
			"hello @there",
		},
		{
			Item{
				text: "hello @there ",
				Priority: Priority('A'),
				Contexts: []string{"there", "foo"},
			},
			"(A) hello @there @foo",
			"hello @there @foo",
		},
		{
			Item{
				Done: true,
				text: "hello @there ",
				Priority: Priority('A'),
				Contexts: []string{"there", "foo"},
			},
			"x (A) hello @there @foo",
			"hello @there @foo",
		},
		{
			Item{
				Done: true,
				FinishDate: d(2012, time.December, 23),
				text: "hello @there ",
				Priority: Priority('A'),
				Contexts: []string{"there", "foo"},
			},
			"x 2012-12-23 (A) hello @there @foo",
			"hello @there @foo",
		},
		{
			Item{
				Done: true,
				FinishDate: d(2012, time.December, 23),
				AddedDate: d(2012, time.December, 20),
				text: "hello @there ",
				Priority: Priority('A'),
				Contexts: []string{"there", "foo"},
			},
			"x 2012-12-23 (A) 2012-12-20 hello @there @foo",
			"hello @there @foo",
		},
		{
			Item{
				text: "+foo @bar +baz @bazm",
				Priority: Priority('A'),
				Projects: []string{"foo"},
				Contexts: []string{"bar"},
			},
			"(A) +foo @bar",
			"+foo @bar",
		},
		{
			Item{
				text: "+foo+bar +baz",
				Priority: Priority('A'),
				Projects: []string{"foo"},
			},
			"(A) +foo",
			"+foo",
		},
		{
			Item{
				text: "+foo+bar",
				Projects: []string{"bar"},
			},
			"+bar",
			"+bar",
		},
		{
			Item{
				text: "a +foo+bar	b	",
				Projects: []string{"bar"},
			},
			"a b	+bar",
			"a b	+bar",
		},
	}
	for _, test := range tests {
		if txt := test.item.Text(); txt != test.txt {
			t.Errorf("Item %v\nexpected text [%s],\ngot [%s]", test.item, test.txt, txt)
		}
		if str := test.item.String(); str != test.str {
			t.Errorf("Item %v\nexpected text [%s],\ngot [%s]", test.item, test.str, str)
		}
		
	}
}

// ItemEq returns true if the two items are equal.
func itemEq(a, b *Item) bool {
	sort.Strings(a.Contexts)
	sort.Strings(b.Contexts)
	sort.Strings(a.Projects)
	sort.Strings(b.Projects)
	return a.text == b.text &&
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