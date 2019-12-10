package testie

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"time"

	"github.com/logrusorgru/aurora"
)

type Testie struct {
	skipcount int
	passcount int
	failcount int

	seen    map[string]*test
	notjson bool

	verbose bool
}

type test struct {
	name       string
	pass       bool
	fail       bool
	skip       bool
	scrollback []string
}

func New(verbose bool) *Testie {
	t := Testie{
		seen:    make(map[string]*test),
		verbose: verbose,
	}
	return &t
}

func reader(ch chan []byte, r io.ReadCloser) {
	rr := bufio.NewReader(r)
	for {
		line, err := rr.ReadBytes('\n')
		if err != nil {
			if err != io.EOF {
				panic(err)
			}
			if len(line) > 0 {
				ch <- line
			}
			close(ch)
			break
		}
		ch <- line
	}
}

func (t *Testie) Run(args []string) int {
	cmd := exec.Command(getGoBinary(), getCommandLine(args)...)

	stderr, err := cmd.StderrPipe()
	if err != nil {
		panic(err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		panic(err)
	}

	errch := make(chan []byte)
	outch := make(chan []byte)

	go reader(errch, stderr)
	go reader(outch, stdout)

	err = cmd.Start()
	if err != nil {
		panic(err)
	}

	errdone := false
	outdone := false

	for {
		select {
		case line, more := <-errch:
			if more {
				t.printLine(line)
			} else {
				errdone = true
			}
		case line, more := <-outch:
			if more {
				t.printLine(line)
			} else {
				outdone = true
			}
		}
		if errdone && outdone {
			break
		}
	}

	if !t.notjson {
		fmt.Printf("%d fail, %d passed, %d skipped, %d total\n",
			t.failcount,
			t.passcount,
			t.skipcount,
			t.failcount+t.passcount+t.skipcount)
		if t.failcount > 0 {
			fmt.Printf("%s\n", aurora.Red("TEST FAILED"))
		}
	}

	if t.failcount > 0 {
		return 1
	} else {
		return 0
	}
}

type record struct {
	Time    time.Time
	Action  string
	Package string
	Test    string
	Output  string
	Elapsed float64
}

/*
The Action field is one of a fixed set of action descriptions:

    run    - the test has started running
    pause  - the test has been paused
    cont   - the test has continued running
    pass   - the test passed
    bench  - the benchmark printed log output but did not fail
    fail   - the test or benchmark failed
    output - the test printed output
    skip   - the test was skipped or the package contained no tests

*/
func (t *Testie) printLine(line []byte) {
	var r record
	err := json.Unmarshal(line, &r)
	if err != nil {
		t.notjson = true
		fmt.Printf("%s", line)
		return
	}
	if len(r.Test) == 0 {
		// Only 'go test' summaries
		return
	}

	//fmt.Printf("%+v\n", r)

	switch r.Action {
	case "run":
		if _, ok := t.seen[r.Test]; !ok {
			t.seen[r.Test] = &test{
				scrollback: make([]string, 0, 100),
			}
		}
	case "skip":
		t.seen[r.Test].skip = true
		t.skipcount++
		if t.verbose {
			fmt.Printf("%s %s\n", aurora.Yellow("skip"), r.Test)
		}
	case "bench":
		fallthrough
	case "output":
		t.seen[r.Test].scrollback = append(t.seen[r.Test].scrollback, r.Output)
	case "pass":
		t.seen[r.Test].pass = true
		t.passcount++
		if t.verbose {
			fmt.Printf("%s %s\n", aurora.Green("pass"), r.Test)
		}
		if r.Elapsed > 1.0 {
			t.printDurationWarning(&r)
		}
	case "fail":
		t.seen[r.Test].fail = true
		t.failcount++
		fmt.Printf("%s %s\n", aurora.Red("fail"), r.Test)
		fmt.Printf("  in package %s\n", aurora.Bold(r.Package))
		fmt.Printf("  here follows test output:\n")
		for _, s := range t.seen[r.Test].scrollback {
			fmt.Printf("    %s", s)
		}
		if r.Elapsed > 1.0 {
			t.printDurationWarning(&r)
		}
	}
}

func (t *Testie) printDurationWarning(r *record) {
	fmt.Printf("%s test '%s' took %0.2fs\n", aurora.Blue("slow"), r.Test, r.Elapsed)
}
