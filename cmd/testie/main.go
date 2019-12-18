package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/grasparv/testie"
)

var helptext = `
  usage: testie ['go test' flags]

  testie is a wrapper utility that executes 'go test' and formats
  the result in a more readable manner. The arguments to testie
  are the same as 'go test', but testie always adds '-json' and
  '-v' internally, so those are not necessary to specify.

  If the environment variable TESTIE is set, those arguments will
  also be passed to 'go test'.

  testie warns if a test takes more than 1 second to run.

  Without arguments, testie only prints:
  
  1) failed tests and their scrollback
  2) warnings about slow tests
  3) a minimal summary at the end

    -s dont print any scrollback even on failures

    -v print passing tests

    -vv print test output

`

func main() {
	short := false
	verbose := false
	extra := false
	debug := false

	var extralist []string
	extras := os.Getenv("TESTIE")
	if len(extras) > 0 {
		extralist = strings.Split(extras, " ")
	}

	args := append(os.Args[1:], extralist...)

	for i := 0; i < len(args); i++ {
		if args[i] == "-v" {
			verbose = true
		} else if args[i] == "-vv" {
			extra = true
		} else if args[i] == "-json" {
		} else if args[i] == "-s" {
			short = true
		} else if args[i] == "-debug" || args[i] == "-d" {
			debug = true
		} else {
			continue
		}
		args = append(args[:i], args[i+1:]...)
		i--
	}

	if len(args) == 1 && args[0] == "-h" {
		fmt.Print(helptext)
		return
	}

	t := testie.New(verbose, extra, debug, short)
	rc := t.Run(args)
	os.Exit(rc)
}
