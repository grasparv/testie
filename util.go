package testie

import (
	"os"
	"strings"
)

func getGoBinary() string {
	s := os.Getenv("PATH")
	for _, p := range strings.Split(s, ":") {
		name := p + "/go"
		finfo, err := os.Stat(name)
		if err != nil {
			continue
		}
		m := finfo.Mode().Perm()
		if m&0b001001001 != 0 {
			return name
		}
	}
	panic("no go binary found")
}

func getCommandLine(args []string) []string {
	static := []string{"test", "-json", "-v"}
	args = append(static, args...)
	return args
}
