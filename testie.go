package testie

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"time"

	"github.com/logrusorgru/aurora"
)

type Testie struct {
	skipcount int
	passcount int
	failcount int

	seen map[string]*test

	short        bool
	debug        bool
	verbose      bool
	extraverbose bool
}

type test struct {
	name string
	pkg  string

	scrollback []string

	pass bool
	fail bool
	skip bool
}

const durationHigh = 1.0

func New(verbose bool, extra bool, debug bool, short bool) *Testie {
	if extra {
		verbose = true
	}
	p := Testie{
		seen:         make(map[string]*test),
		verbose:      verbose,
		extraverbose: extra,
		debug:        debug,
		short:        short,
	}
	return &p
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

func (p *Testie) Run(args []string) int {
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
				p.printLine(line)
			} else {
				errdone = true
			}
		case line, more := <-outch:
			if more {
				p.printLine(line)
			} else {
				outdone = true
			}
		}
		if errdone && outdone {
			break
		}
	}

	fmt.Printf("%d failed, %d passed, %d skipped, %d total\n",
		p.failcount,
		p.passcount,
		p.skipcount,
		p.failcount+p.passcount+p.skipcount)

	if p.failcount > 0 {
		fmt.Printf("%s\n", aurora.Red("TEST FAILED"))
	}

	if p.failcount > 0 {
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
func (p *Testie) printLine(line []byte) {
	var r record
	err := json.Unmarshal(line, &r)
	if err != nil {
		fmt.Printf("%s", line)
		return
	}
	if len(r.Test) == 0 {
		// Only 'go test' summaries
		return
	}
	if p.isMetaTest(&r) {
		return
	}

	if p.debug {
		fmt.Printf("%+v\n", r)
	}

	switch r.Action {
	case "run":
		p.createTest(&r)
	case "skip":
		t := p.getTest(&r)
		t.skip = true
		p.skipcount++
		if p.verbose {
			p.printSkipped(&r)
		}
	case "bench":
		p.printBench(&r)
		if p.extraverbose {
			t := p.getTest(&r)
			p.printScrollback(t, &r)
		}
	case "output":
		p.createTest(&r) // needed for bench
		t := p.getTest(&r)
		t.scrollback = append(t.scrollback, r.Output)
	case "pass":
		t := p.getTest(&r)
		t.pass = true
		p.passcount++
		if p.verbose {
			if p.extraverbose {
				p.printPassed(&r)
				p.printScrollback(t, &r)
			} else {
				p.printPassed(&r)
			}
		}
		p.printDurationWarning(&r)
	case "fail":
		t := p.getTest(&r)
		t.fail = true
		p.failcount++
		p.printFailed(&r)
		p.printScrollback(t, &r)
		p.printDurationWarning(&r)
	}
}

func (p *Testie) isMetaTest(r *record) bool {
	for _, v := range p.seen {
		if strings.Contains(v.name, "/") {
			parts := strings.Split(v.name, "/")
			if len(parts) > 1 && parts[0] == r.Test && r.Package == v.pkg {
				return true
			}
		}
	}
	return false
}

func (p *Testie) makeKey(r *record) string {
	return r.Package + "####" + r.Test
}

func (p *Testie) getTest(r *record) *test {
	return p.seen[p.makeKey(r)]
}

func (p *Testie) createTest(r *record) {
	k := p.makeKey(r)

	if _, ok := p.seen[k]; !ok {
		p.seen[k] = &test{
			scrollback: make([]string, 0, 100),
			pkg:        r.Package,
			name:       r.Test,
		}
	}
}

func (p *Testie) printBench(r *record) {
	fmt.Printf("%s %s%s\n", aurora.Yellow("bnch"), p.getTimingInfo(r), r.Test)
}

func (p *Testie) printSkipped(r *record) {
	fmt.Printf("%s %s%s\n", aurora.Yellow("skip"), p.getTimingInfo(r), r.Test)
}

func (p *Testie) printPassed(r *record) {
	fmt.Printf("%s %s%s\n", aurora.Green("pass"), p.getTimingInfo(r), r.Test)
}

func (p *Testie) printFailed(r *record) {
	fmt.Printf("%s %s%s\n", aurora.Red("fail"), p.getTimingInfo(r), r.Test)
}

func (p *Testie) printRunning(r *record) {
	fmt.Printf("%s %s%s in %s\n", aurora.Bold("run "), r.Test, p.getTimingInfo(r), r.Package)
}

func (p *Testie) printDurationWarning(r *record) {
	if r.Elapsed >= durationHigh {
		fmt.Printf("%s %s took %0.2fs\n", aurora.Blue("slow"), r.Test, r.Elapsed)
	}
}

func (p *Testie) getTimingInfo(r *record) string {
	if p.extraverbose || r.Action == "bench" {
		return fmt.Sprintf("%0.2fs ", r.Elapsed)
	} else {
		return ""
	}
}

func (p *Testie) printScrollback(x *test, r *record) {
	if !p.short {
		t := p.getTest(r)
		fmt.Printf("  in package %s\n", aurora.Bold(r.Package))
		fmt.Printf("  here follows test output:\n")
		for _, s := range t.scrollback {
			fmt.Printf("    %s", s)
		}
	}
}
