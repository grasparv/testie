package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/grasparv/testie/testie"
)

const outfile = "/tmp/testie.log"

const pagelines = 30

const helptext = `
  usage: testie ['go test' flags]

  testie is a wrapper utility that executes 'go test' and formats
  the result in a more readable manner. The arguments to testie
  are the same as 'go test', but testie always adds '-json' and
  '-v' internally, so those are not necessary to specify.

  If the environment variable TESTIE is set, those arguments will
  also be passed to testie and 'go test'.

  testie warns if a test takes more than 1 second to complete
  ("slow"). testie also warns while a test is running if the test
  seems stuck ("hung"), which happens after 10s. Adjust these
  thresholds with the timefactor switch, -tf=XX, for example -tf=0.1
  to make 0.1s be considered slow and 1s be considered "stuck".

  Without arguments, testie only prints:
  
  1) failed tests and their scrollback
  2) warnings about slow tests
  3) a minimal summary at the end

    -s dont print any scrollback even on failures

    -v print passing tests

    -vv print test output

    -tf=0.1 change slow/hung warnings threshold

    -no-slim do indent and keep t.Logf() annotations

    -no-page avoid automatic pagination

`

func main() {
	short := false
	verbose := false
	extra := false
	debug := false
	timefactor := 1.0
	slim := true
	paginate := true

	var extralist []string
	extras := os.Getenv("TESTIE")
	if len(extras) > 0 {
		extralist = strings.Split(extras, " ")
	}

	args := append(os.Args[1:], extralist...)

	for i := 0; i < len(args); i++ {
		if args[i] == "-h" {
			fmt.Print(helptext)
			return
		} else if args[i] == "-v" {
			verbose = true
		} else if args[i] == "-vv" {
			extra = true
		} else if args[i] == "-no-slim" {
			slim = false
		} else if args[i] == "-no-page" {
			paginate = false
		} else if args[i] == "-json" {
		} else if args[i] == "-s" {
			short = true
		} else if args[i] == "-debug" || args[i] == "-d" {
			debug = true
		} else if strings.HasPrefix(args[i], "-tf=") {
			var f float64
			n, err := fmt.Sscanf(args[i], "-tf=%f", &f)
			if n != 1 || err != nil || f <= 0.0 {
				fmt.Print(helptext)
				return
			}
			timefactor = f
		} else {
			continue
		}
		args = append(args[:i], args[i+1:]...)
		i--
	}

	stdout := os.Stdout
	fp := stdout
	if paginate {
		outfp, err := os.Create(outfile)
		if err == nil {
			defer outfp.Close()
			fp = outfp
			os.Stdout = fp
			os.Stderr = fp
		}
	}

	t := testie.New(fp, verbose, extra, debug, short, timefactor, slim)
	rc := t.Run(args)

	if paginate && fp != stdout {
		fp.Close()
		if t.Lines() < pagelines {
			data, err := ioutil.ReadFile(outfile)
			if err == nil {
				fmt.Fprint(stdout, string(data))
			}
		} else {
			cmd := exec.Command("/usr/bin/less", "-SRn", outfile)
			cmd.Stdout = stdout
			cmd.Run()
		}
	}

	os.Exit(rc)
}
