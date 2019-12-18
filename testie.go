package testie

import (
	"bufio"
	"encoding/json"
	"io"
	"os/exec"
	"time"
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
	timefactor   float64
}

type test struct {
	name string
	pkg  string
	t0   time.Time

	scrollback []string

	pass bool
	fail bool
	skip bool
}

type record struct {
	Time    time.Time
	Action  string
	Package string
	Test    string
	Output  string
	Elapsed float64
}

const durationHigh = 1.0
const durationHanging = 10.0
const skipLabel = "skip"
const runLabel = "run"
const benchLabel = "bench"
const outputLabel = "output"
const passLabel = "pass"
const failLabel = "fail"

func New(verbose bool, extra bool, debug bool, short bool, tf float64) *Testie {
	if extra {
		verbose = true
	}
	p := Testie{
		seen:         make(map[string]*test),
		verbose:      verbose,
		extraverbose: extra,
		debug:        debug,
		short:        short,
		timefactor:   tf,
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

	rc := cmd.Wait()

	p.printSummary()

	if p.failcount > 0 {
		p.printSummaryFailure()
	}

	if rc != nil || p.failcount > 0 {
		return 1
	} else {
		return 0
	}
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
		p.printRawLine(line)
		return
	}
	if len(r.Test) == 0 {
		return
	}

	if p.debug {
		p.printDebug(r)
	}

	switch r.Action {
	case runLabel:
		p.createTest(r)
	case skipLabel:
		t := p.getTest(r)
		t.skip = true
		p.skipcount++
		if p.verbose {
			p.printSkipped(r)
		}
	case benchLabel:
		p.printBench(r)
		if p.extraverbose {
			p.printScrollback(r)
		}
	case outputLabel:
		p.createTest(r) // needed for bench
		t := p.getTest(r)
		t.scrollback = append(t.scrollback, r.Output)
	case passLabel:
		t := p.getTest(r)
		t.pass = true
		p.passcount++
		if p.verbose {
			p.printPassed(r)
			if p.extraverbose {
				p.printScrollback(r)
			}
		}
		p.printDurationWarning(r)
	case failLabel:
		t := p.getTest(r)
		t.fail = true
		p.failcount++
		p.printFailed(r)
		p.printScrollback(r)
		p.printDurationWarning(r)
	}
}

func (p Testie) makeKey(r record) string {
	return r.Package + "####" + r.Test
}

func (p *Testie) getTest(r record) *test {
	return p.seen[p.makeKey(r)]
}

func (p *Testie) createTest(r record) {
	k := p.makeKey(r)

	if _, ok := p.seen[k]; !ok {

		t := &test{
			scrollback: make([]string, 0, 100),
			pkg:        r.Package,
			name:       r.Test,
			t0:         time.Now(),
		}

		p.seen[k] = t

		go p.watchdog(t)
	}
}

func (t test) finished() bool {
	if t.pass || t.fail || t.skip {
		return true
	}
	return false
}

func (p Testie) watchdog(t *test) {
	second := int64(time.Second)
	fsecond := float64(second)
	tf := durationHanging * p.timefactor * fsecond
	dtf := time.Duration(tf)
	tick := time.NewTicker(dtf)
	loop := true
	for loop {
		select {
		case <-tick.C:
			if t.finished() {
				loop = false
			} else {
				p.printHungWarning(t)
			}
		}
	}
}
