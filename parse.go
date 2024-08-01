package boa

import (
	"net"
	"net/mail"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/rhysd/abspath"
)

// Optional parameters are designated via negative numbers.
// A command with 2 parameters will use a ParamCount
// of 2 if required or -2 if optional. Optional parameters are used if a
// default parameter is defined or if an empty string is acceptable
// Passing 0 parameters when 1 or more are required results in an error
// Using the list-end sentinel value of '--' is necessary unless the cmd is
// at the last one.     |== lets boa know the expected param is missing
// Example: app --test -- more=more another // requires a param but has a default
const (
	ZeroOrMore = -99  // variable number of args but all are optional
	OneOrMore  = -100 // variable number of args but at least one is required
	OneOrNone  = -1   // fixed number of params but none are required
)

func ParseCommandLineArgs(cmds map[string]CmdLineItem, args []string) *CLI {
	// first get rid of '=' signs
	// then check for compound flags eg. -doe; break up to -d -o -e
	// the last one can have arguments depending on its definition
	args = normalizeArgs(args)

	var cli = CLI{Items: make(map[string]CmdLineItem, len(args))}
	var err error
	var cm *CmdLineItem

	n := 0
	m := 0

	for i := 0; i < len(args); i++ {
		a := args[n]
		// deal with alias passed in
		for _, c := range cmds {
			if c.Alias == a {
				a = c.Name
				break
			}
		}
		args[n] = a // in case the alias was transformed the proper value must be passed to getCmdValues

		m, cm, err = getCmdValues(cmds, a, args[n:])
		n += m // skip the args consumed in the call above
		if err != nil {
			cli.SetError(err)
		}

		if cm != nil {
			cli.Items[cm.Name] = *cm
		}

		if n >= len(args) {
			break
		}
	}

	return &cli
}

func noeq(args []string) []string {
	var result []string
	for _, a := range args {
		a = strings.Trim(a, " ")
		// take care of the pesky'=' sign as in --name=joe
		if strings.Contains(a, "=") {
			noeq := strings.Split(a, "=")
			result = append(result, noeq...)
		} else {
			result = append(result, a)
		}
	}

	return result
}

func normalizeArgs(args []string) []string {
	if len(args) == 0 {
		return nil
	}

	args = noeq(args)

	var result []string
	for _, a := range args {
		if strings.HasPrefix(a, "--") {
			result = append(result, a)
			continue
		} // double dash

		if !strings.HasPrefix(a, "-") {
			result = append(result, a)
			continue
		} // no dashes

		// take care of args entered as -abc; three separate args
		a = strings.Trim(a, "- ")
		for _, r := range a {
			result = append(result, "-"+string(r))
		}
	}

	return result
}

func getCmdValues(cmds map[string]CmdLineItem, a string, args []string) (int, *CmdLineItem, error) {
	result, exist := cmds[a]
	if !exist {
		result, exist = cmds["--"+a]
		if !exist {
			return 1, nil, Errorf(BeInvalidCommand, a)
		}
	}

	if len(args) < 1 {
		return 0, nil, nil
	}

	// no args allowed
	if result.ParamCount == 0 {
		result.Value = true
		result.ParamType = TypeBool
		return 1, &result, nil
	}

	switch result.ParamType {
	case TypeInt:
		var n int64

		i, res, err := parseArg(args, &result, Errorf(BeNoRequiredInt, args[0]))
		if err != nil {
			return i, &result, err
		}
		n, err = strconv.ParseInt(res, 10, 64)
		if err != nil {
			return i, &result, Errorf(BeNotAnInt, BeNotAnInt.String(), args[1], a)
		}

		result.Value = int(n)
		return i, &result, nil

	case TypeFloat:
		var n float64

		i, res, err := parseArg(args, &result, Errorf(BeNoRequiredFloat, a))
		if err != nil {
			return i, &result, err
		}
		n, err = strconv.ParseFloat(res, 64)
		if err != nil {
			return i, &result, Errorf(BeNotAFloat, BeNotAFloat.String(), args[1], a)

		}

		result.Value = n
		return i, &result, nil

	case TypeString:
		i, res, err := parseArg(args, &result, Errorf(BeNoRequiredString, a))
		if err != nil {
			return i, &result, err
		}
		result.Value = res
		return i, &result, nil

	case TypeEmail:
		i, res, err := parseArg(args, &result, Errorf(BeNoRequiredEmail, a))
		if err != nil {
			return i, &result, err
		}

		email, err := mail.ParseAddress(res)
		if err != nil || email == nil {
			return i, &result, Errorf(BeNotAnEmail, BeNotAnEmail.String(), res, a)
		}

		result.Value = *email
		return i, &result, nil

	case TypePhone:
		i, res, err := parseArg(args, &result, Errorf(BeNoRequiredPhone, a))
		if err != nil {
			return i, &result, err
		}

		re := regexp.MustCompile(`^(?:(?:\(?(?:00|\+)([1-4]\d\d|[1-9]\d?)\)?)?[\-\.\ \\\/]?)?((?:\(?\d{1,}\)?[\-\.\ \\\/]?){0,})(?:[\-\.\ \\\/]?(?:#|ext\.?|extension|x)[\-\.\ \\\/]?(\d+))?$`)
		b := re.MatchString(res)
		if !b {
			return i, &result, Errorf(BeNotAPhone, BeNotAPhone.String(), args[1], a)
		}

		result.Value = res
		return i, &result, nil

	case TypeTime:
		i, res, err := parseArg(args, &result, Errorf(BeNoRequiredPhone, a))
		if err != nil {
			return i, &result, err
		}

		timeval, err := time.Parse(time.Kitchen, res)
		if err != nil {
			return i, &result, Errorf(BeNotATime, args[1], a)
		}

		result.Value = timeval
		return i, &result, nil

	case TypeTimeDuration:
		i, res, err := parseArg(args, &result, Errorf(BeNoRequiredDuration, a))
		if err != nil {
			return i, &result, err
		}

		// A duration string is a possibly signed sequence of decimal
		// numbers, each with optional fraction and a unit suffix,
		// such as "300ms", "-1.5h" or "2h45m". Valid time units are
		// "ns", "us" (or "µs"), "ms", "s", "m", "h".
		duration, err := time.ParseDuration(res)
		if err != nil {
			return i, &result, Errorf(BeNotADuration, args[1], a)
		}
		result.Value = duration
		return i, &result, nil

	case TypeDate:
		i, res, err := parseArg(args, &result, Errorf(BeNoRequiredDate, a))
		if err != nil {
			return i, &result, err
		}

		// below is the way the reference date would be represented
		// in the desired layout; it has no time zone present.
		// Note: without explicit zone, returns time in UTC.
		const format = "Jan-02-2006"
		dateval, err := time.Parse(format, res)
		if err != nil {
			return i, &result, Errorf(BeNotADate, args[1], a)
		}
		result.Value = dateval
		return i, &result, nil

	case TypePath:
		i, res, err := parseArg(args, &result, Errorf(BeNoRequiredPath, a))
		if err != nil {
			return i, &result, err
		}

		path, err := abspath.ExpandFrom(res)
		if err != nil {
			return i, &result, Errorf(BeNotAPath, args[1], a)
		}
		result.Value = path.String()
		return i, &result, nil

	case TypeURL:
		i, res, err := parseArg(args, &result, Errorf(BeNoRequiredURL, a))
		if err != nil {
			return i, &result, err
		}

		url, err := url.ParseRequestURI(res)
		if err != nil || url == nil {
			return i, &result, Errorf(BeNotAURL, args[1], a)
		}

		result.Value = url
		return i, &result, nil

	case TypeIPv4:
		i, res, err := parseArg(args, &result, Errorf(BeNoRequiredIPv4, a))
		if err != nil {
			return i, &result, err
		}

		ip := net.ParseIP(res)
		if ip == nil {
			return i, &result, Errorf(BeNotAnIPv4, args[1], a)
		}
		result.Value = ip
		return i, &result, nil

	// slice types-------------------------

	case TypeIntSlice:
		var vals []int

		i, vs, err := parseSlice(args, &result, Errorf(BeNoRequiredInt, a))
		if err != nil {
			return i, &result, err
		}

		for _, v := range vs {
			n, err := strconv.ParseInt(v, 10, 64)
			if err != nil {
				return i, &result, Errorf(BeNotAnInt, v, a)
			}
			vals = append(vals, int(n))
		}

		result.Value = vals
		return i, &result, nil

	case TypeFloatSlice:
		var vals []float64

		i, vs, err := parseSlice(args, &result, Errorf(BeNoRequiredFloat, a))
		if err != nil {
			return i, &result, err
		}

		for _, v := range vs {
			n, err := strconv.ParseFloat(v, 64)
			if err != nil {
				return i, &result, Errorf(BeNotAFloat, v, a)
			}
			vals = append(vals, float64(n))
		}

		result.Value = vals
		return i, &result, nil

	case TypeStringSlice:
		var vals []string

		i, vs, err := parseSlice(args, &result, Errorf(BeNoRequiredString, a))
		if err != nil {
			return i, &result, err
		}

		vals = append(vals, vs...)
		result.Value = vals
		return i, &result, nil

	case TypeEmailSlice:
		var vals []mail.Address

		i, vs, err := parseSlice(args, &result, Errorf(BeNoRequiredEmail, a))
		if err != nil {
			return i, &result, err
		}

		for _, v := range vs {
			email, err := mail.ParseAddress(v)
			if err != nil {
				return i, &result, Errorf(BeNotAnEmail, v, a)
			}
			if email != nil {
				vals = append(vals, *email)
			}
		}
		result.Value = vals
		return i, &result, nil

	case TypePhoneSlice:
		var vals []string

		i, vs, err := parseSlice(args, &result, Errorf(BeNoRequiredPhone, a))
		if err != nil {
			return i, &result, err
		}

		for _, v := range vs {
			re := regexp.MustCompile(`^(?:(?:\(?(?:00|\+)([1-4]\d\d|[1-9]\d?)\)?)?[\-\.\ \\\/]?)?((?:\(?\d{1,}\)?[\-\.\ \\\/]?){0,})(?:[\-\.\ \\\/]?(?:#|ext\.?|extension|x)[\-\.\ \\\/]?(\d+))?$`)
			b := re.MatchString(v)
			if !b {
				return i, &result, Errorf(BeNotAPhone, v, a)
			}
			vals = append(vals, v)
		}
		result.Value = vals
		return i, &result, nil

	case TypeTimeSlice:
		var vals []time.Time

		i, vs, err := parseSlice(args, &result, Errorf(BeNoRequiredTime, a))
		if err != nil {
			return i, &result, err
		}

		for _, v := range vs {
			timeval, err := time.Parse(time.Kitchen, v)
			if err != nil {
				return i, &result, Errorf(BeNotATime, v, a)
			}
			vals = append(vals, timeval)
		}
		result.Value = vals
		return i, &result, nil

	case TypeTimeDurationSlice:
		var vals []time.Duration

		i, vs, err := parseSlice(args, &result, Errorf(BeNoRequiredDuration, a))
		if err != nil {
			return i, &result, err
		}

		for _, v := range vs {
			duration, err := time.ParseDuration(v)
			if err != nil {
				return i, &result, Errorf(BeNotADuration, v, a)
			}
			vals = append(vals, duration)
		}
		result.Value = vals
		return i, &result, nil

	case TypeDateSlice:
		var vals []time.Time

		i, vs, err := parseSlice(args, &result, Errorf(BeNoRequiredDate, a))
		if err != nil {
			return i, &result, err
		}

		for _, v := range vs {
			const format = "Jan-02-2006"
			dateval, err := time.Parse(format, v)
			if err != nil {
				return i, &result, Errorf(BeNotADate, v, a)
			}
			vals = append(vals, dateval)
		}
		result.Value = vals
		return i, &result, nil

	case TypePathSlice:
		var vals []string

		i, vs, err := parseSlice(args, &result, Errorf(BeNoRequiredPath, a))
		if err != nil {
			return i, &result, err
		}

		for _, v := range vs {
			path, err := abspath.ExpandFrom(v)
			if err != nil {
				return i, &result, Errorf(BeNotAPath, v, a)
			}
			vals = append(vals, path.String())
		}
		result.Value = vals
		return i, &result, nil

	case TypeURLSlice:
		var vals []url.URL

		i, vs, err := parseSlice(args, &result, Errorf(BeNoRequiredURL, a))
		if err != nil {
			return i, &result, err
		}

		for _, v := range vs {
			url, err := url.ParseRequestURI(v)
			if err != nil {
				return 2, &result, Errorf(BeNotAURL, v, a)
			}
			if url != nil {
				vals = append(vals, *url)
			}
		}
		result.Value = vals
		return i, &result, nil

	case TypeIPv4Slice:
		var vals []net.IP

		i, vs, err := parseSlice(args, &result, Errorf(BeNoRequiredIPv4, a))
		for _, v := range vs {
			ip := net.ParseIP(v)
			if ip == nil {
				return i, &result, err
			}
			vals = append(vals, ip)
		}
		result.Value = vals
		return i, &result, nil
	}

	return 1, nil, nil
}

func parseArg(args []string, cmd *CmdLineItem, err error) (int, string, error) {
	//get sub-commands if any for comparison against the args
	var hasChild bool
	if cmd.ChNames != nil {
		hasChild = true
	}

	match := func(a string) (string, bool) {
		for _, c := range cmd.ChNames {
			if c == a {
				return c, true
			}
		}
		return "", false
	}

	if len(args) <= 1 { // no arg given at the last cmd
		if cmd.DefaultValue != "" { // use default value if defined
			return 1, cmd.DefaultValue, nil
		}
		if cmd.ParamCount < 0 && cmd.ParamCount > -100 { // optional params 0 or more
			return 1, "", err
		}
		return 1, "", nil
	}
	if args[1] == "--" { // use DefaultValue even if it is an empty string
		return 2, cmd.DefaultValue, nil
	}
	var i = 0
	if hasChild {
		var chl []string
		for _, a := range args[1:] {
			if ch, ok := match(a); ok { // sub commands have to be used first, before the actual parameter
				chl = append(chl, ch)
				i++
				continue
			}
			cmd.ChNames = chl // cmd.ChNames now holds the sub commands actually used in this instance
			return 2 + i, args[i+1], nil
		}
	}
	return 2 + i, args[1], nil
}

func parseSlice(args []string, cmd *CmdLineItem, err error) (int, []string, error) {
	var vals []string
	//get sub-commands if any for comparison against the args
	var hasChild bool
	if cmd.ChNames != nil {
		hasChild = true
	}

	match := func(a string) (string, bool) {
		for _, c := range cmd.ChNames {
			if c == a {
				return c, true
			}
		}
		return "", false
	}

	if len(args) < 2 {
		if cmd.DefaultValue != "" {
			vals = append(vals, cmd.DefaultValue)
			return 1, vals, nil
		}
		if cmd.ParamCount < 0 && cmd.ParamCount > OneOrMore { // params are optional
			return 1, vals, nil
		}
		return 1, vals, err
	}

	j := 1
	for n, arg := range args {
		var chl []string
		if hasChild {
			for _, a := range args[n:] {
				if a == "--" {
					break
				}
				if ch, ok := match(a); ok { // sub commands have to be used first, before the actual parameter
					chl = append(chl, ch)
				}
			}
		}

		if hasChild {
			cmd.ChNames = chl // cmd.ChNames now holds the sub commands actually used in this instance
		}

		if arg == "--" {
			j++
			break
		}

		vals = append(vals, arg)
		j++
	}
	return j, vals, nil
}
