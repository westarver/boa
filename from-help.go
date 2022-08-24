package boa

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/bitfield/script"
)

const (
	eol               = -1
	eof               = -1
	Usagepat          = `^[Uu]{1}sage: `
	CommandSectionPat = `^Commands[\t ]*:`
	FlagSectionPat    = `^Flags[\t ]*:`
	MoreSectionPat    = `^More[\t ]*:`
	LongSectionPat    = `^Long Description[\t ]*:`
)

type scanfunc func(*helpParser) scanfunc

type helpParser struct {
	usagepat          *regexp.Regexp
	commandsectionpat *regexp.Regexp
	flagsectionpat    *regexp.Regexp
	moresectionpat    *regexp.Regexp
	longsectionpat    *regexp.Regexp
	itemMap           map[string]CmdLineItem
	text              string
	lines             []string
	errs              []parseError
	pos               int
	line              int
	cmd               int
	flg               int
	lng               int
	more              int
}

func (h *helpParser) appendArg(a CmdLineItem) {
	h.itemMap[a.name] = a
}

func (h *helpParser) nextLine() string {
	h.line++
	if h.line >= len(h.lines) {
		h.setError(newError(eof, "reached end of text"))
		return ""
	}

	h.text = h.lines[h.line]
	h.setPos(0)
	return h.text
}

func (h *helpParser) Line() string {
	if h.Pos() > len(h.text) {
		return h.nextLine()
	}
	if n := h.text[h.Pos()]; n == '\n' {
		return h.nextLine()
	}
	return h.text[h.Pos():]
}

func (h *helpParser) Pos() int {
	return h.pos
}

func (h *helpParser) setPos(n int) {
	if n >= len(h.Line()) {
		h.nextLine()
	} else {
		h.pos = n
	}
}

func (h *helpParser) setError(err parseError) {
	h.errs = append(h.errs, err)
}

func (h *helpParser) Errors() string {
	var errs []string
	for _, e := range h.errs {
		errs = append(errs, e.Error())
	}
	return strings.Join(errs, "\n")
}

func FromHelp(helpstring string) *CLI {
	lines, _ := script.Echo(helpstring).Slice()
	parser := helpParser{
		usagepat:          regexp.MustCompile(Usagepat),
		commandsectionpat: regexp.MustCompile(CommandSectionPat),
		flagsectionpat:    regexp.MustCompile(FlagSectionPat),
		moresectionpat:    regexp.MustCompile(MoreSectionPat),
		longsectionpat:    regexp.MustCompile(LongSectionPat),
		itemMap:           map[string]CmdLineItem{},
		text:              "",
		lines:             lines,
		errs:              []parseError{},
		pos:               0,
		line:              0,
		cmd:               0,
		flg:               0,
		lng:               0,
		more:              0,
	}

	parser.cmd, parser.flg, parser.lng, parser.more = getLimits(&parser)
	parser.text = parser.lines[0]
	scanf := scanUsage(&parser)
	for {
		if scanf == nil {
			break
		}
		scanf = scanf(&parser)
	}

	fmt.Fprintln(os.Stderr, parser.Errors())

	cli := ParseCommandLineArgs(parser.itemMap, os.Args[1:])
	cli.AllHelp = make(map[string]string)
	for _, item := range parser.itemMap {
		cli.AllHelp[item.name] = formatHelp(item.name, item.alias, item.shortHelp, item.longHelp)
	}
	return cli
}

//─────────────┤ formatHelp ├─────────────

func formatHelp(name, alias, short, long string) string {
	name = strings.Trim(name, " \t")
	s := strings.Trim(short, "\t\n ")
	s = strings.TrimPrefix(s, name)
	s = strings.TrimPrefix(s, ":")
	s = strings.Trim(s, "\t\n ") + "\n"
	var comb string
	if len(alias) > 0 {
		comb = name + " | " + alias
	} else {
		comb = name
	}

	var spc int
	if len(comb) > 12 {
		spc = 4
	} else {
		spc = 16 - len(comb)
	}

	s = comb + strings.Repeat(" ", spc) + s //+ "\n"
	if len(long) == 0 {
		return s
	}

	lng := strings.Split(long, "\n")
	var ret []string
	for _, l := range lng {
		l = strings.Trim(l, "\t ")
		l = strings.TrimPrefix(l, name)
		l = strings.TrimPrefix(l, ":")
		l = strings.Trim(l, "\t ")

		ret = append(ret, l)
	}
	long = strings.Join(ret, "\n")
	return s + long
}

//─────────────┤ getLimits ├─────────────

func getLimits(h *helpParser) (int, int, int, int) {
	var cmd = -1
	var flg = -1
	var lng = -1
	var more = -1
	for i, ln := range h.lines {
		loc := h.commandsectionpat.FindStringIndex(ln)
		if loc != nil {
			cmd = i
		}
		loc = h.flagsectionpat.FindStringIndex(ln)
		if loc != nil {
			flg = i
		}
		loc = h.longsectionpat.FindStringIndex(ln)
		if loc != nil {
			lng = i
		}
		loc = h.moresectionpat.FindStringIndex(ln)
		if loc != nil {
			more = i
		}
	}
	return cmd, flg, lng, more
}

//─────────────┤ scanUsage ├─────────────

func scanUsage(h *helpParser) scanfunc {
	loc := h.usagepat.FindStringIndex(h.lines[0])
	if loc == nil {
		return nil
	}
	if loc[0] > 0 {
		return nil
	}
	h.setPos(loc[1])
	return scanCommand
}

//─────────────┤ scanCommand ├─────────────

func scanCommand(h *helpParser) scanfunc {

	if 2 >= len(h.lines) {
		h.setError(newError(WrongFileFormat, "First line in help text must be Usage: etc. followed by blank line. See example"))
		return nil
	}

	if h.cmd == -1 {
		return scanFlag
	}

	var limit int
	if h.flg != -1 && h.flg > h.cmd {
		limit = h.flg
	} else if h.lng != -1 && h.lng > h.cmd {
		limit = h.lng
	} else if h.more != -1 {
		limit = h.more
	} else {
		limit = len(h.lines)
	}

	for i := h.cmd + 1; i < limit; i++ {
		pos := 0
		line := strings.Trim(h.lines[i], "\t ")
		if len(line) == 0 {
			continue
		}

		item := CmdLineItem{}
		item.Type = StringType // initial type is string

		// scan for meta characters before command name, interpret
		// * means exclusive, + means default, # means int, . means float
		meta := "*+#."
		if len(line) > pos+3 {
			if !strings.ContainsAny(line[pos:pos+3], meta) {
				item.exclusive = false
				item.isDefault = false
			} else {
				if strings.Contains(line[pos:pos+3], "*") {
					item.exclusive = true
					pos += 1
				}
				if strings.Contains(line[pos:pos+3], "+") {
					item.isDefault = true
					pos += 1
				}
				if strings.Contains(line[pos:pos+3], "#") {
					item.Type = IntType
					pos += 1
				}
				if strings.Contains(line[pos:pos+3], ".") {
					item.Type = FloatType
					pos += 1
				}
			}
		}

		var cmd, remnant string
		if strings.HasPrefix(line[pos:], "[") {
			item.required = false
			pos += 1
			if n := strings.Index(line, "]"); n != -1 {
				cmd = line[pos:n]
				pos += n - 1
				remnant = line[pos:]
			}
		} else {
			if n := strings.Index(line, " "); n != -1 {
				cmd = line[pos:n]
				pos += n - 1
				remnant = line[pos:]
			} else {
				cmd = line[pos:]
			}
			item.required = true
		}
		sl := strings.Split(cmd, " | ")
		if len(sl) > 1 {
			cmd = strings.Trim(sl[0], "\t ")
			item.alias = strings.Trim(sl[1], "\t ")
		}

		r := regexp.MustCompile(`<[a-zA-Z0-9-_.]*>`)
		loc := r.FindStringIndex(remnant)
		if loc != nil {
			if len(remnant) >= loc[1]+3 {
				if strings.Contains(remnant[loc[1]:loc[1]+3], "...") {
					switch item.Type {
					case StringType:
						item.Type = StringSliceType
					case IntType:
						item.Type = IntSliceType
					case FloatType:
						item.Type = FloatSliceType
					default:
						h.setError(newError(UnsupportedType, "unsupported type at line %d", h.line))
					}
				}
			}
		} else {
			item.Type = BoolType
		}

		n := strings.Index(remnant, ":")
		if n != -1 {
			item.shortHelp = cmd + "\t-" + strings.Trim(remnant[n+1:], "\t \n")
		}

		if len(cmd) > 0 {
			item.name = cmd
			item.longHelp = scanLong(h, cmd)
			h.appendArg(item)
		}
	}
	return scanFlag
}

//─────────────┤ scanFlag ├─────────────

func scanFlag(h *helpParser) scanfunc {

	if h.flg == -1 {
		return nil
	}

	var limit int
	if h.lng != -1 && h.lng > h.flg {
		limit = h.lng
	} else if h.more != -1 {
		limit = h.more
	} else {
		limit = len(h.lines)
	}

	for i := h.flg + 1; i < limit; i++ {
		pos := 0
		line := strings.Trim(h.lines[i], "\t ")
		if len(line) == 0 {
			continue
		}

		item := CmdLineItem{}
		item.Type = StringType // initial type is string

		// scan for meta characters before command name, interpret
		// * means exclusive, + means default, # means int, . means float
		meta := "*+#."
		if len(line) > pos+3 {
			if !strings.ContainsAny(line[pos:pos+3], meta) {
				item.exclusive = false
				item.isDefault = false
			} else {
				if strings.Contains(line[pos:pos+3], "*") {
					item.exclusive = true
					pos += 1
				}
				if strings.Contains(line[pos:pos+3], "+") {
					item.isDefault = true
					pos += 1
				}
				if strings.Contains(line[pos:pos+3], "#") {
					item.Type = IntType
					pos += 1
				}
				if strings.Contains(line[pos:pos+3], ".") {
					item.Type = FloatType
					pos += 1
				}
			}
		}
		var flag, remnant string
		if strings.HasPrefix(line[pos:], "[") {
			item.required = false
			pos += 1
			if n := strings.Index(line, "]"); n != -1 {
				flag = line[pos:n]
				pos += n - 1
				remnant = line[pos:]
			}
		} else {
			if n := strings.Index(line, " "); n != -1 {
				flag = line[pos:n]
				pos += n - 1
				remnant = line[pos:]
			} else {
				flag = line[pos:]
			}
			item.required = true
		}

		sl := strings.Split(flag, " | ")
		if len(sl) > 1 {
			flag = strings.Trim(sl[0], "\t ")
			item.alias = strings.Trim(sl[1], "\t ")
		}

		r := regexp.MustCompile(`<[a-zA-Z0-9-_.]*>`)
		loc := r.FindStringIndex(remnant)
		if loc != nil {
			if len(remnant) >= loc[1]+3 {
				if strings.Contains(remnant[loc[1]:loc[1]+3], "...") {
					switch item.Type {
					case StringType:
						item.Type = StringSliceType
					case IntType:
						item.Type = IntSliceType
					case FloatType:
						item.Type = FloatSliceType
					default:
						h.setError(newError(UnsupportedType, "unsupported type at line %d", h.line))
					}
				}
			}
		} else {
			item.Type = BoolType //no args to flag means its true if present
		}
		n := strings.Index(remnant, ":")
		if n != -1 {
			item.shortHelp = flag + "\t-" + strings.Trim(remnant[n+1:], "\t \n")
		}

		item.isFlag = true
		if len(flag) > 0 {
			item.name = flag
			item.longHelp = scanLong(h, flag)
			h.appendArg(item)
		}
	}
	return nil
}

//─────────────┤ scanLong ├─────────────

func scanLong(h *helpParser, name string) string {

	if len(name) == 0 {
		return ""
	}

	if h.lng == -1 {
		return ""
	}

	var limit int
	if h.more != -1 {
		limit = h.more
	} else {
		limit = len(h.lines)
	}
	var long []string
	start, end := getLimitsForName(h.lines[h.lng+1:limit], name, h.lng)
	if start == -1 {
		return ""
	}
	if start == h.lng { //dirty hack to fix off-by-one error
		start++
	}

	long = h.lines[start:end]
	return strings.Join(long, "\n")
}

func getLimitsForName(lines []string, name string, i int) (int, int) {
	r := regexp.MustCompile(`^[[:print:]]+[[:blank:]]*:[[:blank:]]+`)
	var start, end int
	var found bool
	for j, ln := range lines {
		loc := r.FindStringIndex(ln)
		if loc != nil {
			str := ln[loc[0]:loc[1]]
			str = strings.Trim(str, "\t ")
			str = strings.TrimSuffix(str, ":")
			if str == name {
				start = j
				found = true
				break
			}
		}
	}
	if start < len(lines)-1 {
		lines = lines[start+1:]
		var j int
		for _, ln := range lines {
			j++
			loc := r.FindStringIndex(ln)
			if loc != nil {
				break
			}
		}
		end = start + j + 1
	} else {
		end = start
	}
	if start == 0 && !found {
		return -1, -1
	}
	return start + i, end + i
}
