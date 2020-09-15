package testie

import (
	"fmt"
	"os"
	"time"

	"golang.org/x/sys/unix"
)

var sizeX *uint16
var sizeY *uint16
var lastPrint time.Time

func getWinsize() *unix.Winsize {
	if sizeX != nil && sizeY != nil {
		return &unix.Winsize{Col: *sizeX, Row: *sizeY}
	}

	ws, err := unix.IoctlGetWinsize(int(os.Stdout.Fd()), unix.TIOCGWINSZ)
	if err != nil {
		panic(err)
	}

	sizeX = &ws.Col
	sizeY = &ws.Row

	return ws
}

func getColumns() int {
	ws := getWinsize()
	return int(ws.Col)
}

func (p *Testie) clearLastLine() {
	cols := getColumns()
	os.Stdout.Write([]byte{'\r'})
	for i := 0; i < cols; i++ {
		os.Stdout.Write([]byte{' '})
	}
	os.Stdout.Write([]byte{'\r'})
	p.lastlinesize = 0
}

func (p *Testie) printLastLine(s string) {
	if time.Since(lastPrint) <= time.Millisecond*500 {
		return
	}

	s = p.findContent(s)
	if s != "" {
		cols := getColumns()
		if len(s) > cols {
			s = s[:cols]
		}
		p.printStatus(true, "%s", s)
		lastPrint = time.Now()
	}
}

func (p Testie) findContent(s string) string {
	start := len(s) - 1
	lastidx := start

	for {
		copied := 0
		bytes := make([]byte, len(s))

		for i := start; i >= 0; i-- {
			lastidx = i
			if s[i] == 27 || s[i] >= 32 && s[i] < 127 {
				copied++
				bytes[len(s)-copied] = s[i]
			}
			if s[i] == '\r' || s[i] == '\n' {
				break
			}
		}

		if copied > 8 {
			return string(bytes[len(s)-copied:])
		}

		start = lastidx - 1

		if start <= 0 {
			break
		}
	}

	return ""
}

func (p *Testie) printStatus(status bool, format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	if status {
		p.clearLastLine()
		fmt.Fprint(p.fpStatus, msg)
		p.lastlinesize = len(msg)
	}
}
