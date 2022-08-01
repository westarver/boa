package boa

import (
	"fmt"
	"os"
	"reflect"
	"regexp"
	"strings"

	"github.com/bitfield/script"
)

const (
	eol               = -1
	eof               = -1
	newline           = 10
	tab               = 9
	Usagepat          = `[Uu]{1}sage: `
	CommandSectionPat = `Commands[\t ]*:`
	FlagSectionPat    = `Flags[\t ]*:`
	MoreSectionPat    = `More[\t ]*:`
	StringDef         = "var DesignData = `"
)

type scanfunc func(*helpParser) scanfunc

type helpParser struct {
	usagepat          *regexp.Regexp
	commandsectionpat *regexp.Regexp
	flagsectionpat    *regexp.Regexp
	moresectionpat    *regexp.Regexp
	itemMap           map[string]cmdLineArg
	text              string
	lines             []string
	errs              []parseError
	pos               int
	i                 int
	width             int
	line              int
}

func (h *helpParser) appendArg(a cmdLineArg) {
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

func doFromHelp(helpdoc string) *CLI {
	hlp, _ := script.Echo(helpdoc).Slice()
	parser := helpParser{
		usagepat:          regexp.MustCompile(Usagepat),
		commandsectionpat: regexp.MustCompile(CommandSectionPat),
		flagsectionpat:    regexp.MustCompile(FlagSectionPat),
		moresectionpat:    regexp.MustCompile(MoreSectionPat),
		itemMap:           map[string]cmdLineArg{},
		text:              "",
		lines:             hlp,
		errs:              []parseError{},
		pos:               0,
		i:                 0,
		width:             0,
		line:              0,
	}

	parser.text = parser.lines[0]
	scanf := scanUsage(&parser)
	for {
		if scanf == nil {
			break
		}
		scanf = scanf(&parser)
	}

	fmt.Fprintln(os.Stderr, parser.Errors())
	return parseCommandLineArgs(parser.itemMap, os.Args[1:])

}

//─────────────┤ scanUsage ├─────────────

func scanUsage(h *helpParser) scanfunc {
	loc := h.usagepat.FindStringIndex(h.Line()[h.Pos():])
	if loc == nil {
		return nil
	}
	if loc[0] > 0 {
		return nil
	}
	h.setPos(h.Pos() + loc[1])
	return scanCommand
}

//─────────────┤ scanCommand ├─────────────

func scanCommand(h *helpParser) scanfunc {
	if 2 >= len(h.lines) {
		return nil
	}
	var found bool
	var num int
	for i := 2; i < len(h.lines); i++ { //requires a blank line before 'Commands'
		loc := h.commandsectionpat.FindStringIndex(h.lines[i])
		if loc != nil {
			found = true
			num = i
			break
		}
	}
	if !found {
		return scanFlag
	}

	if num >= len(h.lines) {
		return nil
	}

	for i := num + 1; i < len(h.lines); i++ {
		pos := 0
		line := strings.Trim(h.lines[i], "\t ")
		if len(line) == 0 {
			continue
		}
		loc := h.flagsectionpat.FindStringIndex(line)
		if loc != nil {
			return scanFlag
		}

		intType := reflect.TypeOf(int64(0))
		boolType := reflect.TypeOf(true)
		floatType := reflect.TypeOf(float64(0.0))
		stringType := reflect.TypeOf("")
		intSliceType := reflect.TypeOf([]int64{})
		floatSliceType := reflect.TypeOf([]float64{})
		stringSliceType := reflect.TypeOf([]string{})

		item := cmdLineArg{}
		item.Type = stringType // initial type is string

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
					item.Type = intType
					pos += 1
				}
				if strings.Contains(line[pos:pos+3], ".") {
					item.Type = floatType
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
		loc = r.FindStringIndex(remnant)
		if loc != nil {
			if len(remnant) >= loc[1]+3 {
				if strings.Contains(remnant[loc[1]:loc[1]+3], "...") {
					switch item.Type {
					case stringType:
						item.Type = stringSliceType
					case intType:
						item.Type = intSliceType
					case floatType:
						item.Type = floatSliceType
					default:
						h.setError(newError(UnsupportedType, "unsupported type at line %d", h.line))
					}
				}
			}
		} else {
			item.Type = boolType
		}

		n := strings.Index(remnant, ":")
		if n != -1 {
			item.shortHelp = strings.Trim(remnant[n+1:], "\t \n")
		}
		if len(cmd) > 0 {
			item.name = cmd
			h.appendArg(item)
		}
	}
	return scanFlag
}

//─────────────┤ scanFlag ├─────────────

func scanFlag(h *helpParser) scanfunc {
	if 2 >= len(h.lines) {
		return nil
	}
	var found bool
	var num int
	for i := 2; i < len(h.lines); i++ { //requires a blank line before 'Commands'
		loc := h.flagsectionpat.FindStringIndex(h.lines[i])
		if loc != nil {
			found = true
			num = i
			break
		}
	}

	if !found {
		return nil
	}

	if num >= len(h.lines) {
		return nil
	}

	for i := num + 1; i < len(h.lines); i++ {
		pos := 0
		line := strings.Trim(h.lines[i], "\t ")
		if len(line) == 0 {
			continue
		}
		loc := h.moresectionpat.FindStringIndex(line)
		if loc != nil {
			return nil
		}

		intType := reflect.TypeOf(int64(0))
		boolType := reflect.TypeOf(true)
		floatType := reflect.TypeOf(float64(0.0))
		stringType := reflect.TypeOf("")
		intSliceType := reflect.TypeOf([]int64{})
		floatSliceType := reflect.TypeOf([]float64{})
		stringSliceType := reflect.TypeOf([]string{})

		item := cmdLineArg{}
		item.Type = stringType // initial type is string

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
					item.Type = intType
					pos += 1
				}
				if strings.Contains(line[pos:pos+3], ".") {
					item.Type = floatType
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
		if strings.Contains(line[pos:], "+") {
			item.isDefault = true
		}
		r := regexp.MustCompile(`<[a-zA-Z0-9-_.]*>`)
		loc = r.FindStringIndex(remnant)
		if loc != nil {
			if len(remnant) >= loc[1]+3 {
				if strings.Contains(remnant[loc[1]:loc[1]+3], "...") {
					switch item.Type {
					case stringType:
						item.Type = stringSliceType
					case intType:
						item.Type = intSliceType
					case floatType:
						item.Type = floatSliceType
					default:
						h.setError(newError(UnsupportedType, "unsupported type at line %d", h.line))
					}
				}
			}
		} else {
			item.Type = boolType
		}
		n := strings.Index(remnant, ":")
		if n != -1 {
			item.shortHelp = strings.Trim(remnant[n+1:], "\t \n")
		}
		item.isFlag = true
		if len(flag) > 0 {
			h.appendArg(item)
		}
	}
	return nil
}

// //─────────────┤ zeroValForType ├─────────────

// func zeroValForType(t string) string {
// 	if t == "string" {
// 		return EmptyStr
// 	}

// 	if t == "bool" {
// 		return "false"
// 	}

// 	var pre string
// 	var slice bool
// 	if strings.HasPrefix(t, "*") {
// 		pre = "&"
// 		t = t[1:]
// 	}

// 	if strings.HasPrefix(t, "[") {
// 		slice = true
// 		t = t[1:]
// 	}

// 	if t == "string" {
// 		if slice {
// 			return "[]" + t + "{}"
// 		}
// 		if len(pre) != 0 {
// 			return "nil"
// 		}
// 		return EmptyStr
// 	}
// 	if t == "bool" {
// 		if slice {
// 			return "[]" + t + "{}"
// 		}
// 		if len(pre) != 0 {
// 			return "nil"
// 		}
// 		return "false"
// 	}
// 	if strings.HasPrefix(t, "float") {
// 		if slice {
// 			return "[]" + t + "{}"
// 		}
// 		if len(pre) != 0 {
// 			return "nil"
// 		}
// 		return "0.0"
// 	}
// 	if strings.HasPrefix(t, "complex") {
// 		if slice {
// 			return "[]" + t + "{}"
// 		}
// 		if len(pre) != 0 {
// 			return "nil"
// 		}
// 		return "(0 + 0i)"
// 	}

// 	types := "int8 uint8 byte int16 uint16 int32 rune uint32 int64 uint64 int uint"
// 	loc := helper.FindStringinString(types, t)
// 	if len(loc) != 0 {
// 		return "0"
// 	}
// 	return "nil"
// }
