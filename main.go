package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/grasparv/testie/testie"
)

const outfile = "/tmp/testie.log"
const helptext = `
  usage: testie ['go test' flags]

  testie is a wrapper utility that executes 'go test' and formats the
  result in a more readable manner. The arguments to testie are the same
  as 'go test'. If the environment variable TESTIE is set, those
  arguments will also be passed to testie and 'go test'.

  testie warns if a test takes more than 1 second to complete
  ("slow"). testie also warns while a test is running if the test
  seems stuck ("hung"), which happens after 10s. Adjust these
  thresholds with the timefactor switch, -tf=XX, for example -tf=0.1
  to make 0.1s be considered slow and 1s be considered "stuck".

`

func main() {
	timefactor := 1.0
	didselection := false

	var extralist []string
	extras := os.Getenv("TESTIE")
	if len(extras) > 0 {
		extralist = strings.Split(extras, " ")
	}

	args := append(os.Args[1:], extralist...)

	for i := 0; i < len(args); i++ {
		if args[i] == "-h" ||
			args[i] == "-json" ||
			args[i] == "-v" {
			fmt.Print(helptext)
			return
		} else if strings.HasPrefix(args[i], "-run=") {
			didselection = true
			continue
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

	t := testie.New(outfile, didselection, timefactor)
	rc := t.Run(args)

	if t.DoPaging() {
		cmd := exec.Command("/usr/bin/less", "-SRn", outfile)
		cmd.Stdout = os.Stdout
		cmd.Run()
	}

	os.Exit(rc)
}
