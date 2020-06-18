package testie

import (
	"fmt"
	"time"

	"github.com/logrusorgru/aurora"
)

func (p Testie) getTimingInfo(r record) string {
	if p.extraverbose || r.Action == benchLabel {
		return fmt.Sprintf("%0.2fs ", r.Elapsed)
	} else {
		return ""
	}
}

func (p Testie) printBench(r record) {
	fmt.Fprintf(p.Fp, "%s %s%s\n", aurora.Yellow("bnch"), p.getTimingInfo(r), r.Test)
}

func (p Testie) printSkipped(r record) {
	fmt.Fprintf(p.Fp, "%s %s%s\n", aurora.Yellow("skip"), p.getTimingInfo(r), r.Test)
}

func (p Testie) printPassed(r record) {
	fmt.Fprintf(p.Fp, "%s %s%s\n", aurora.Green("pass"), p.getTimingInfo(r), r.Test)
}

func (p Testie) printFailed(r record) {
	fmt.Fprintf(p.Fp, "%s %s%s\n", aurora.Red("fail"), p.getTimingInfo(r), r.Test)
}

func (p Testie) printRunning(r record) {
	fmt.Fprintf(p.Fp, "%s %s%s in %s\n", aurora.Bold("run "), r.Test, p.getTimingInfo(r), r.Package)
}

func (p Testie) printDurationWarning(r record) {
	if r.Elapsed >= durationHigh*p.timefactor {
		fmt.Fprintf(p.Fp, "%s %s took %0.2fs\n", aurora.Blue("slow"), r.Test, r.Elapsed)
	}
}

func (p Testie) printHungWarning(t *test) {
	fmt.Fprintf(p.Fp, "%s %s, ran for %v\n", aurora.Blue("hung"), t.name, time.Since(t.t0))
}

func (p Testie) printScrollback(r record) {
	if !p.short {
		t := p.getTest(r)
		if !p.slim {
			fmt.Fprintf(p.Fp, "  in package %s\n", aurora.Bold(r.Package))
			fmt.Fprintf(p.Fp, "  here follows test output:\n")
		} else {
			fmt.Fprintf(p.Fp, "in package %s\n", aurora.Bold(r.Package))
			fmt.Fprintf(p.Fp, "here follows test output:\n")
		}
		for _, s := range t.scrollback {
			if !p.slim {
				fmt.Fprintf(p.Fp, "    %s", s)
			} else {
				if tmp := p.slimRegexp.FindStringIndex(s); tmp != nil {
					fmt.Print(s[tmp[1]:])
				} else {
					fmt.Print(s)
				}
			}
		}
	}
}

func (p Testie) printGolangWarning(err error) {
	fmt.Fprintf(p.Fp, "%s\n", aurora.Red(fmt.Sprintf("go test %s", err)))
}

func (p Testie) printSummaryFailure() {
	fmt.Fprintf(p.Fp, "%s\n", aurora.Red("TEST FAILED"))
}

func (p Testie) printSummary() {
	fmt.Fprintf(p.Fp, "%d failed, %d passed, %d skipped, %d total\n",
		p.failcount,
		p.passcount,
		p.skipcount,
		p.failcount+p.passcount+p.skipcount)
}

func (p Testie) printDebug(r record) {
	fmt.Fprintf(p.Fp, "%+v\n", r)
}

func (p Testie) printRawLine(line []byte) {
	fmt.Fprintf(p.Fp, "%s", line)
}
