package testie

import (
	"bufio"
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"regexp"
	"time"
)

type Testie struct {
	skipcount int
	passcount int
	failcount int

	updates   []*record
	lastflush int

	seen map[string]*test

	timefactor float64
	didselect  bool

	slimRegexp *regexp.Regexp

	fpStatus     io.Writer
	fpOutput     io.Writer
	lastlinesize int
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

func New(outfile string, didselection bool, tf float64) *Testie {
	re, err := regexp.Compile(`^[\s]+Test[^:]+: [^\.]+\.(go|s):\d+: `)
	if err != nil {
		panic(err)
	}

	p := Testie{
		fpStatus:   os.Stdout,
		seen:       make(map[string]*test),
		timefactor: tf,
		slimRegexp: re,
		didselect:  didselection,
		updates:    make([]*record, 0),
	}

	p.fpOutput, err = os.Create(outfile)
	if err != nil {
		panic(err)
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

	p.flush()
	p.printSummary()

	if p.failcount > 0 {
		p.printSummaryFailure()
		return 1
	} else if rc != nil {
		p.printGolangWarning(rc)
		p.printSummaryFailure()
		return 1
	} else if p.totalCount() == 0 {
		p.printNoTests()
		return 1
	} else {
		return 0
	}
}

func (p Testie) totalCount() int {
	return p.failcount + p.skipcount + p.passcount
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

	p.updates = append(p.updates, &r)

	switch r.Action {
	case runLabel:
		p.createTest(r)
	case skipLabel:
		t := p.getTest(r)
		t.skip = true
		p.skipcount++
		p.flush()
	case benchLabel:
		p.flush()
	case outputLabel:
		p.createTest(r) // needed for bench
		t := p.getTest(r)
		t.scrollback = append(t.scrollback, r.Output)
		p.printLastLine(r.Output)
	case passLabel:
		t := p.getTest(r)
		t.pass = true
		p.passcount++
		p.flush()
	case failLabel:
		t := p.getTest(r)
		t.fail = true
		p.failcount++
		p.flush()
	}
}

func (p Testie) DoPaging() bool {
	if p.failcount > 0 || // any failures
		p.didselect { // did selection
		return true
	}
	return false
}

func (p *Testie) flush() {
	for i := p.lastflush; i < len(p.updates); i++ {
		r := *p.updates[i]

		switch r.Action {
		case runLabel:
		case skipLabel:
			if p.didselect {
				p.printScrollback(r)
			}
			p.printSkipped(r)
		case benchLabel:
			p.printBench(r)
			p.printScrollback(r)
		case outputLabel:
		case passLabel:
			p.printPassed(r)
			if p.didselect {
				p.printScrollback(r)
			}
			p.printDurationWarning(r)
		case failLabel:
			p.printFailed(r)
			p.printScrollback(r)
			p.printDurationWarning(r)
		}
		p.lastflush = i + 1
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

	for loop := true; loop; {
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
