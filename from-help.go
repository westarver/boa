package boa

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/bitfield/script"
)

const (
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
	app               string
	text              string
	lines             []string
	errs              []ParseError
	pos               int
	line              int
	cmd               int
	flg               int
	lng               int
	more              int
}

func (h *helpParser) appendArg(a CmdLineItem) {
	h.itemMap[a.Name] = a
}

func (h *helpParser) nextLine() string {
	h.line++
	if h.line >= len(h.lines) {
		h.setError(NewParseError(BeEofError, StringFromCode(BeEofError)))
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

func (h *helpParser) setError(err ParseError) {
	h.errs = append(h.errs, err)
}

func (h *helpParser) Errors() string {
	var errs []string
	for _, e := range h.errs {
		errs = append(errs, e.Error())
	}
	return strings.Join(errs, "\n")
}

func CollectItems(helpstring string) (map[string]CmdLineItem, string, string) {
	lines, _ := script.Echo(helpstring).Slice()
	parser := helpParser{
		usagepat:          regexp.MustCompile(Usagepat),
		commandsectionpat: regexp.MustCompile(CommandSectionPat),
		flagsectionpat:    regexp.MustCompile(FlagSectionPat),
		moresectionpat:    regexp.MustCompile(MoreSectionPat),
		longsectionpat:    regexp.MustCompile(LongSectionPat),
		itemMap:           map[string]CmdLineItem{},
		app:               "",
		text:              "",
		lines:             lines,
		errs:              []ParseError{},
		pos:               0,
		line:              0,
		cmd:               0,
		flg:               0,
		lng:               0,
		more:              0,
	}

	getLimits(&parser)
	parser.text = parser.lines[0]
	scanf := scanUsage(&parser)
	for {
		if scanf == nil {
			break
		}
		scanf = scanf(&parser)
	}

	return parser.itemMap, parser.Errors(), parser.app
}

func FromHelp(helpstring string, args []string) *CLI {
	items, errs, app := CollectItems(helpstring)
	cli := ParseCommandLineArgs(items, args)
	validateRequirements(items, cli)
	cli.Application = app
	cli.SetError(errors.New(errs))
	cli.AllHelp = make(map[string]string)
	for _, item := range items {
		cli.AllHelp[item.Name] = formatHelp(item.Name, item.Alias, item.ShortHelp, item.LongHelp)
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

	s = comb + strings.Repeat(" ", spc) + s
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

func getLimits(h *helpParser) {
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
	h.cmd = cmd
	h.flg = flg
	h.lng = lng
	h.more = more
}

//─────────────┤ scanUsage ├─────────────

func scanUsage(h *helpParser) scanfunc {
	loc := h.usagepat.FindStringIndex(h.lines[0])
	if loc == nil {
		h.setError(NewParseError(BeWrongFileFormat, "First line in input script must be Usage: etc. followed by blank line. See example"))
		return nil
	}
	if loc[0] > 0 {
		h.setError(NewParseError(BeWrongFileFormat, "First line in input script must be Usage: etc. followed by blank line. See example"))
		return nil
	}
	h.app = strings.Fields(h.lines[0])[1]
	h.setPos(loc[1])
	if 2 >= len(h.lines) {
		h.setError(NewParseError(BeWrongFileFormat, "First line in input script must be Usage: etc. followed by blank line. See example"))
		return nil
	}

	return scanCommandAndFlag
}

//regex patterns for scanning individual command/flag lines

// `[*+#.!^%\\/@&]`
var meta = regexp.MustCompile(`[*+#.!^%\\/@&]`)

//`[^*+#.!^%\\/@&]`
var nonMeta = regexp.MustCompile(`[^*+#.!^%\\/@&]`)

//`\[?[[:word:]\s\|-]+\]?`
var nameAlias = regexp.MustCompile(`\[?[[:word:]\s\|-]+\]?`)

//`\[?[0-9]*\]?`
var paramPat = regexp.MustCompile(`\[?[0-9]*\]?`)

//─────────────┤ scanCommandAndFlag ├─────────────

func scanCommandAndFlag(h *helpParser) scanfunc {
	if h.flg < h.cmd { // flags must come after commands
		h.setError(ErrorFromCode(BeWrongFileFormat))
		return nil
	}

	if h.cmd == -1 && h.flg == -1 { // no commands or flags!
		h.setError(ErrorFromCode(BeWrongFileFormat))
		return nil
	}

	var inFlag bool
	if h.cmd == -1 && h.flg != -1 { // no command section only flags
		inFlag = true
	}

	var flagStart = len(h.lines) // default to eof

	if h.flg > h.cmd {
		flagStart = h.flg
	}

	var limit int
	if h.lng != -1 && h.lng > h.cmd {
		limit = h.lng
	} else if h.more != -1 {
		limit = h.more
	} else {
		limit = len(h.lines)
	}

	for i := h.cmd + 1; i < limit; i++ {
		if i == flagStart {
			continue
		}
		if i > flagStart {
			inFlag = true
		}

		pos := 0
		line := strings.Trim(h.lines[i], "\t ")
		if len(line) == 0 {
			continue
		}

		item := CmdLineItem{ParamType: TypeString} // default to string type

		if inFlag {
			item.IsFlag = true
		}

		// if all goes well the paramType field will be populated
		pos, err := scanMeta(line, &item)
		if err != nil {
			switch err.(ParseError).Code {
			case BeBadMetaLine, BeMetaNotStart:
				h.setError(err.(ParseError))
				return nil
			}
		}
		line = line[pos:]

		// populate the name, alias fields
		pos, err = scanName(line, &item)
		if err != nil {
			if err.(ParseError).Code == BeNoCommandName {
				h.setError(err.(ParseError))
				return nil
			}
		}

		var cmd = item.Name
		line = line[pos:]

		// populate the paramCount field
		pos, err = scanParams(line, &item)
		if err != nil {
			h.setError(NewParseError(0, err.Error()+" at line %d", h.line))
			return nil
		}

		line = line[pos:]

		// correct the paramType field if not 1 or 0
		if item.ParamCount != 0 && item.ParamCount != 1 {
			switch item.ParamType {
			case TypeString:
				item.ParamType = TypeStringSlice
			case TypeInt:
				item.ParamType = TypeIntSlice
			case TypeFloat:
				item.ParamType = TypeFloatSlice
			case TypeDate:
				item.ParamType = TypeDateSlice
			case TypeTime:
				item.ParamType = TypeTimeSlice
			case TypeTimeDuration:
				item.ParamType = TypeTimeDurationSlice
			case TypePath:
				item.ParamType = TypePathSlice
			case TypeURL:
				item.ParamType = TypeURLSlice
			case TypeEmail:
				item.ParamType = TypeEmailSlice
			case TypePhone:
				item.ParamType = TypePhoneSlice
			case TypeIPv4:
				item.ParamType = TypeIPv4Slice
			default:
				h.setError(NewParseError(BeUnsupportedType, StringFromCode(BeUnsupportedType), h.line))
			}
		}
		n := strings.Index(line, ":")
		if n != -1 {
			item.ShortHelp = strings.Trim(line[n+1:], "\t \n")
		}

		if len(cmd) > 0 {
			item.LongHelp = scanLong(h, cmd)
			item.Id = i
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
	start, end := getLimitsForName(h.lines[:limit], name, h.lng)
	if start == -1 {
		return ""
	}

	long = h.lines[start:end]
	for i := range long {
		long[i] = strings.TrimPrefix(long[i], name)
		long[i] = strings.Trim(long[i], "\t ")
		long[i] = strings.Trim(long[i], ":")
		long[i] = strings.Trim(long[i], "\t ")
	}
	return strings.Join(long, "\n")
}

func getLimitsForName(lines []string, name string, i int) (int, int) {
	r := regexp.MustCompile(`^` + name + `[[:blank:]]*:`)
	var start, end int
	var found bool
	for j := i + 1; j < len(lines); j++ {
		loc := r.FindStringIndex(lines[j])
		if loc != nil {
			str := lines[j][loc[0]:loc[1]]
			str = strings.Trim(str, "\t ")
			str = strings.TrimSuffix(str, ":")
			str = strings.Trim(str, "\t ")
			if str == name {
				start = j
				found = true
				break
			}
		}
	}
	if !found {
		return -1, -1
	}

	if start < len(lines)-1 {
		var j int
		for i := start; i < len(lines); i++ {
			if strings.Trim(lines[i], "\t ") == "" {
				break
			}
			j++
		}
		end = start + j
	} else {
		end = start
	}
	if start == 0 && !found {
		return -1, -1
	}
	return start, end
}

// scan for meta characters before command/flag name, interpret.
// * means exclusive, + means default, # means int, . means float
// ! means date, % means time, ^ is duration, \ means url, / is path,
// @ is email, & is IPv4
func scanMeta(line string, item *CmdLineItem) (int, error) {

	// 1st non-meta char terminates the meta string. Any following
	// meta chars are parsed as part of the normal command/flag
	// description and will fail as only word characters are permitted
	notmeta := nonMeta.FindStringIndex(line)
	if notmeta == nil {
		return len(line), NewParseError(BeBadMetaLine, StringFromCode(BeBadMetaLine))
	}

	metastr := line[:notmeta[0]]
	if len(metastr) == 0 {
		item.Exclusive = false
		item.IsDefault = false
		return 0, nil
	}

	loc := meta.FindStringIndex(metastr)
	if loc == nil {
		item.Exclusive = false
		item.IsDefault = false
		return 0, nil
	}

	if loc[0] != 0 {
		return 0, NewParseError(BeMetaNotStart, StringFromCode(BeMetaNotStart))
	}

	var metarunes []rune
	for i := 0; i < len(metastr); i++ {
		metarunes = append(metarunes, rune(metastr[i]))
	}

	for i := 0; i < len(metarunes); i++ {
		switch metarunes[i] {
		case '*':
			item.Exclusive = true
		case '+':
			item.IsDefault = true
		case '#':
			item.ParamType = TypeInt
		case '.':
			item.ParamType = TypeFloat
		case '!':
			item.ParamType = TypeDate
		case '%':
			item.ParamType = TypeTime
		case '^':
			item.ParamType = TypeTimeDuration
		case '/':
			item.ParamType = TypePath
		case '\\':
			item.ParamType = TypeURL
		case '@':
			item.ParamType = TypeEmail
		case '&':
			item.ParamType = TypeIPv4
		}
	}

	return len(metarunes), nil
}

func scanName(line string, item *CmdLineItem) (int, error) {
	nameloc := nameAlias.FindStringIndex(line)
	if nameloc == nil {
		return 0, ErrorFromCode(BeNoCommandName)
	}
	namestr := line[nameloc[0]:nameloc[1]]
	if !strings.HasPrefix(namestr, "[") {
		item.Required = true
	} else {
		namestr = strings.TrimSuffix(strings.TrimPrefix(namestr, "["), "]")
	}

	al := strings.Split(namestr, "|")
	item.Name = strings.TrimSpace(al[0])

	if len(al) > 1 {
		item.Alias = strings.TrimSpace(al[1])
	}

	return nameloc[1], nil
}

func scanParams(line string, item *CmdLineItem) (int, error) {
	colonpos := strings.Index(line, ":")
	if colonpos == -1 {
		item.ParamCount = 0
		return 0, fmt.Errorf("no colon followed by short description found after %s", item.Name)
	}

	line = line[:colonpos]
	line = strings.TrimSpace(line)

	if line == "" {
		item.ParamCount = 0
		item.ParamType = TypeBool // if no params the command is true if found on command line
		return colonpos - 1, nil
	}

	if line == "..." {
		item.ParamCount = -1
		return colonpos - 1, nil
	}

	parloc := paramPat.FindStringIndex(line)

	if parloc == nil {
		item.ParamCount = 0
		cf := "command"
		if item.IsFlag {
			cf = "flag"
		}
		return colonpos - 1, fmt.Errorf("unrecognized %s, %s found after %s", cf, line, item.Name)
	}

	parstr := line[parloc[0]:parloc[1]]
	parstr = strings.TrimSpace(parstr)
	if strings.HasPrefix(parstr, "[") {
		parstr = strings.TrimSuffix(strings.TrimPrefix(parstr, "["), "]")
		item.ParamOpt = true
	}
	if parstr == "" {
		item.ParamCount = 0
		return colonpos - 1, nil
	}

	ct, err := strconv.ParseInt(parstr, 10, 64)
	if err != nil {
		return colonpos - 1, err
	}

	item.ParamCount = int(ct)
	return colonpos - 1, nil
}
