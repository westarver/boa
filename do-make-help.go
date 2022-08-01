package boa

import (
	"fmt"
	"os"
	"strings"

	"github.com/westarver/helper"
)

//─────────────┤ doMakeHelp ├─────────────

func doMakeHelp(helpdoc, program string, commands, flags []string) error {
	var (
		hlp    strings.Builder
		cmdstr = "command"
		flstr  = "flag"
	)

	lc := len(commands)
	if lc > 1 {
		cmdstr += "s..."
	}

	lf := len(flags)
	if lf > 1 {
		flstr += "s..."
	}

	hlp.WriteString(fmt.Sprintf("Usage: %s [%s] [%s]\n\n", program, cmdstr, flstr))

	maxlen := maxLen(commands, flags) + 3
	if lc > 0 {
		hlp.WriteString("Commands:\n")
		for _, c := range commands {
			s := c + strings.Repeat(" ", maxlen-len(c)) + ": short help string for " + c
			hlp.WriteString(s + "\n")
		}
	}

	if lf > 0 {
		hlp.WriteString("\n")
		hlp.WriteString("Flags:\n")
		for _, f := range flags {
			s := f + strings.Repeat(" ", maxlen-len(f)) + ": short help string for " + f
			hlp.WriteString(s + "\n")
		}
	}

	hlp.WriteString("\nMore:\ngeneral help text as needed here\n")
	var err error
	var out *os.File
	if len(helpdoc) == 0 {
		fmt.Fprintln(os.Stdout, hlp.String())
	} else {
		out, err = helper.OpenTrunc(helpdoc)
		if err != nil {
			err = fmt.Errorf("writing to stdout due to %s", err.Error())
			fmt.Fprintln(os.Stdout, hlp.String())
		} else {
			fmt.Fprintln(out, hlp.String())
			out.Close()
		}
	}

	return err
}

func maxLen(sl ...[]string) int {
	maxlen := 0
	for _, s := range sl {
		for _, it := range s {
			if len(it) > maxlen {
				maxlen = len(it)
			}
		}
	}
	return maxlen
}
