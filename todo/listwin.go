// © 2012 Ethan Burns under the MIT license.

package main

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"

	"code.google.com/p/eaburns.todo/acme"
	"code.google.com/p/eaburns.todo/todotxt"
)

// A listWin is a window listing the todo.txt file, possibly with some
// filters applied to it.
type listWin struct {
	*acme.Win
	title   string
	filters []string
	less    func([]todotxt.Task, int, int) bool
}

// NewListWin creates a new list window for this set of filters.
func newListWin(filters []string) {
	title := fmt.Sprintf("%s/%s", path, strings.Join(filters, ""))
	win, err := acme.New(title)
	if err != nil {
		die(1, "Failed to create a new window %s: %s", title, err)
	}
	if err := win.Fprintf("tag", "Sort "); err != nil {
		die(1, "Failed to write the tag of %s: %s", title, err)
	}
	lw := &listWin{
		Win:     win,
		title:   title,
		filters: filters,
		less:    lessFuncs["prio"],
	}
	wg.Add(1)
	go lw.events()
	lw.refresh()
}

// lessFuncs is a map of less functions for sorting
var lessFuncs = map[string]func([]todotxt.Task, int, int) bool{
	"line": func(_ []todotxt.Task, i, j int) bool {
		return i < j
	},
	"prio": func(ts []todotxt.Task, i, j int) bool {
		switch a, b := ts[i], ts[j]; {
		case !a.IsDone() && b.IsDone():
			return true
		case a.IsDone() && !b.IsDone():
			return false
		case a.Priority() != b.Priority():
			return a.Priority() < b.Priority()
		}
		return i < j
	},
}

// Events deals with the window events, meant to be run in a
// separate go routine.
func (lw *listWin) events() {
	defer wg.Done()
	for ev := range lw.EventChan() {
		switch {
		case ev.C2 == 'l' || ev.C2 == 'L':
			if err := lw.WriteEvent(ev); err != nil {
				die(1, "Failed to write an event to %s: %s", lw.title, err)
			}

		case ev.C2 == 'x' || ev.C2 == 'X':
			fs := strings.Fields(string(ev.Text))
			if len(fs) >= 1 && fs[0] == "Sort" {
				if len(fs) > 1 {
					if less, ok := lessFuncs[fs[1]]; ok {
						lw.less = less
						lw.refresh()
						continue
					}
				}
				lst := ""
				for n := range lessFuncs {
					lst += n + " "
				}
				lst = strings.TrimSpace(lst)
				fmt.Fprintln(os.Stderr, "Valid sort functions are:", lst)
				continue
			}
			if (ev.Flag & 0x1) != 0 { // acme command
				if err := lw.WriteEvent(ev); err != nil {
					die(1, "Failed to write an event to %s: %s", lw.title, err)
				}
				if len(fs) > 0 && fs[0] == "Del" {
					return
				}
			}
			if filterOk(fs) {
				fsNew := make([]string, len(lw.filters))
				copy(fsNew, lw.filters)
				for _, f := range fs {
					found := false
					for _, f2 := range fsNew {
						if f == f2 {
							found = true
							break
						}
					}
					if !found {
						fsNew = append(fsNew, f)
					}
				}
				newListWin(fsNew)
			}
		}
	}
}

// FilterOk returns true if every element of the slice is a valid filter tag.
func filterOk(fs []string) bool {
	for _, f := range fs {
		if f[0] != todotxt.ProjectTag && f[0] != todotxt.ContextTag {
			return false
		}
		l, _ := utf8.DecodeLastRuneInString(f)
		if !unicode.IsLetter(l) && !unicode.IsDigit(l) && l != '_' {
			return false
		}
	}
	return true
}

// Refresh refreshes the window's body by re-parsing the file.
func (lw *listWin) refresh() {
	var inds []int
	for i, task := range file.Tasks {
		ok := true
		for _, filter := range lw.filters {
			if !task.HasTag(filter) {
				ok = false
				break
			}
		}
		if ok {
			inds = append(inds, i)
		}
	}

	sort.Sort(sorter{inds, file.Tasks, lw.less})

	projs := make(map[string]bool)
	ctxs := make(map[string]bool)

	if err := lw.Addr(","); err != nil {
		die(1, "Failed to set address for %s: %s", lw.title, err)
	}

	for _, i := range inds {
		task := file.Tasks[i]
		if _, err := fmt.Fprintf(lw.Data, "%5d. %s\n", i+1, task.String()); err != nil {
			die(1, "Failed to refresh window %s: %s", lw.title, err)
		}
		for _, t := range task.Tags(todotxt.ProjectTag) {
			projs[t] = true
		}
		for _, t := range task.Tags(todotxt.ContextTag) {
			ctxs[t] = true
		}
	}

	if err := lw.Addr("#0"); err != nil {
		die(1, "Failed to write address to %s: %s", lw.title, err)
	}
	if err := lw.Ctl("dot=addr"); err != nil {
		die(1, "Failed to write dot=addr to %s ctl: %s", lw.title, err)
	}
	if err := lw.Ctl("show"); err != nil {
		die(1, "Failed to write show to %s ctl: %s", lw.title, err)
	}
	if err := lw.Ctl("clean"); err != nil {
		die(1, "Failed to write clean to %s ctl: %s", lw.title, err)
	}
}

// A sorter sorts the indices using the less function from the listWin.
type sorter struct {
	inds  []int
	tasks []todotxt.Task
	less  func([]todotxt.Task, int, int) bool
}

func (s sorter) Len() int {
	return len(s.inds)
}

func (s sorter) Swap(i, j int) {
	s.inds[i], s.inds[j] = s.inds[j], s.inds[i]
}

func (s sorter) Less(i, j int) bool {
	return s.less(s.tasks, s.inds[i], s.inds[j])
}
