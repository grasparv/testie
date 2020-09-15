package testie

import (
	"fmt"
	"time"

	"github.com/logrusorgru/aurora"
)

func (p Testie) getTimingInfo(r record) string {
	if r.Action == benchLabel {
		return fmt.Sprintf("%0.2fs ", r.Elapsed)
	} else {
		return ""
	}
}

func (p *Testie) print(status bool, format string, a ...interface{}) {
	if status {
		fmt.Fprintf(p.fpStatus, format, a...)
	}
	if p.fpOutput != nil {
		fmt.Fprintf(p.fpOutput, format, a...)
	}
}

func (p *Testie) printBench(r record) {
	p.print(true, "%s %s%s\n", aurora.Yellow("bnch"), p.getTimingInfo(r), r.Test)
}

func (p *Testie) printSkipped(r record) {
	p.print(true, "%s %s%s\n", aurora.Yellow("skip"), p.getTimingInfo(r), r.Test)
}

func (p *Testie) printPassed(r record) {
	p.print(true, "%s %s%s\n", aurora.Green("pass"), p.getTimingInfo(r), r.Test)
}

func (p *Testie) printFailed(r record) {
	p.print(true, "%s %s%s\n", aurora.Red("fail"), p.getTimingInfo(r), r.Test)
}

func (p *Testie) printDurationWarning(r record) {
	if r.Elapsed >= durationHigh*p.timefactor {
		p.print(true, "%s %s took %0.2fs\n", aurora.Blue("slow"), r.Test, r.Elapsed)
	}
}

func (p *Testie) printHungWarning(t *test) {
	p.print(true, "%s %s, ran for %v\n", aurora.Blue("hung"), t.name, time.Since(t.t0))
}

func (p *Testie) printScrollback(r record) {
	t := p.getTest(r)
	p.print(false, "in package %s\n", aurora.Bold(r.Package))
	for _, s := range t.scrollback {
		if tmp := p.slimRegexp.FindStringIndex(s); tmp != nil {
			p.print(false, s[tmp[1]:])
		} else {
			p.print(false, s)
		}
	}
}

func (p *Testie) printNoTests() {
	p.print(true, "%s\n", aurora.Red("no tests found, report as error"))
}

func (p *Testie) printGolangWarning(err error) {
	p.print(true, "%s\n", aurora.Red(fmt.Sprintf("go test %s", err)))
}

func (p *Testie) printSummaryFailure() {
	p.print(true, "%s\n", aurora.Red("TEST FAILED"))
}

func (p *Testie) printSummary() {
	p.print(true, "%d failed, %d passed, %d skipped, %d total\n", p.failcount, p.passcount, p.skipcount, p.failcount+p.passcount+p.skipcount)
}

func (p *Testie) printRawLine(line []byte) {
	p.print(false, "%s", line)
}
