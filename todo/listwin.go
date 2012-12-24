// © 2012 Ethan Burns under the MIT license.

package main

import (
	"fmt"
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
}

// NewListWin creates a new list window for this set of filters.
func newListWin(filters []string) {
	title := fmt.Sprintf("%s/%s", path, strings.Join(filters, ""))
	win, err := acme.New(title)
	if err != nil {
		die(1, "Failed to create a new window %s: %s", title, err)
	}
	lw := &listWin{win, title, filters}
	wg.Add(1)
	go lw.events()
	lw.refresh()
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
			if (ev.Flag & 0x1) != 0 { // acme command
				if err := lw.WriteEvent(ev); err != nil {
					die(1, "Failed to write an event to %s: %s", lw.title, err)
				}
				if len(fs) > 0 && fs[0] == "Del" {
					return
				}
			}
			if filterOk(fs) {
				newListWin(fs)
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
	for i, task := range file.Tasks {
		ok := true
		for _, filter := range lw.filters {
			if !task.HasTag(filter) {
				ok = false
				break
			}
		}
		if !ok {
			continue
		}
		if _, err := fmt.Fprintf(lw.Data, "%5d. %s\n", i, task.String()); err != nil {
			die(1, "Failed to refresh window %s: %s", lw.title, err)
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
