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
  '-v' internally.

  If the environment variable TESTIE is set, those arguments will
  also be passed to 'go test'.

  testie warns if a test takes more than 1 second to run.

  testie can be given -v to also print passing tests.
`

func main() {
	verbose := false

	var extralist []string
	extras := os.Getenv("TESTIE")
	if len(extras) > 0 {
		extralist = strings.Split(extras, " ")
	}

	args := append(os.Args[1:], extralist...)

	for i, a := range args {
		if a == "-v" {
			verbose = true
			args = append(args[:i], args[i+1:]...)
		}
	}

	if len(args) == 1 && args[0] == "-h" {
		fmt.Println(helptext)
		return
	}

	t := testie.New(verbose)
	rc := t.Run(args)
	os.Exit(rc)
}
