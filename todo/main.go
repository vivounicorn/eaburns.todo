// © 2012 Ethan Burns under the MIT license.

package main

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"code.google.com/p/eaburns.todo/todotxt"
)

var (
	path string
	file todotxt.File
	wg   sync.WaitGroup
)

func main() {
	if len(os.Args) != 2 {
		die(2, "Usage: todo <todo.txt path>\n")
	}
	path = os.Args[1]

	file = readFile()

	win := newListWin(nil)
	if wd, err := os.Getwd(); err != nil {
		panic("Failed to set dump working directory: " + err.Error())
	} else {
		win.Ctl("dumpdir %s", wd)
		win.Ctl("dump %s", strings.Join(os.Args, " "))
	}

	wg.Wait()
}

// Die prints a message to standard error and exits with the given status.
func die(status int, f string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, f, args...)
	os.Exit(status)
}

// ReadFile returns the todotxt.File.
func readFile() todotxt.File {
	in, err := os.Open(path)
	if err != nil {
		die(1, "Failed to open %s: %s\n", path, err)
	}
	defer in.Close()
	file, err := todotxt.ReadFile(in)
	if err != nil {
		die(1, "Failed to read %s: %s\n", path, err)
	}
	return file
}
