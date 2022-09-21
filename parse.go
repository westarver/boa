package boa

import (
	"fmt"
	"net"
	"net/mail"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/rhysd/abspath"
)

func ParseCommandLineArgs(cmds map[string]CmdLineItem, args []string) *CLI {
	var cli = CLI{Items: make(map[string]CmdLineItem, len(args))}
	var err error
	var cm *CmdLineItem

	// first check for compound flags eg. -doe break up to -d -o -e
	// the last one can have arguments depending on its definition
	args = normalizeArgs(args)

	var n = 0
	var m = 0

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
		n += m
		if err != nil {
			cli.SetError(err)
		}

		if cm != nil {
			cli.Items[cm.Name] = *cm
		} else {
			cli.SetError(fmt.Errorf("%s - %s", a, StringFromCode(BeInvalidCommand)))
		}
		if n >= len(args) {
			break
		}
	}

	return &cli
}

func normalizeArgs(args []string) []string {
	var result []string

	for _, a := range args {
		if !strings.HasPrefix(a, "--") {
			if strings.HasPrefix(a, "-") {
				if utf8.RuneCount([]byte(a)) > 2 {
					temp := "-"
					a = a[1:]
					rune, w := utf8.DecodeRuneInString(a)
					temp += string(rune)
					result = append(result, temp)
					a = a[w:]
					for {
						if len(a) == 0 {
							break
						}
						temp := "-"
						rune, w := utf8.DecodeRuneInString(a)
						temp += string(rune)
						result = append(result, temp)
						a = a[w:]
					}
				}
			}
		}
		result = append(result, a)
	}
	return result
}

func getCmdValues(cmds map[string]CmdLineItem, a string, args []string) (int, *CmdLineItem, error) {
	result, exist := cmds[a]
	if !exist {
		result, exist = cmds["--"+a]
		if !exist {
			return 1, nil, nil
		}
	}

	if len(args) < 1 {
		return 1, nil, nil
	}

	// no args allowed
	if result.ParamCount == 0 {
		result.Value = true
		result.ParamType = TypeBool
		return 1, &result, nil
	}

	var err error
	switch result.ParamType {
	case TypeBool:
		result.Value = true
		result.ParamCount = 0
		return 1, &result, nil
	case TypeInt:
		var n int64
		if len(args) <= 1 {
			if !result.ParamOpt {
				return 1, &result, NewParseError(BeNoRequiredInt, StringFromCode(BeNoRequiredInt), a)
			}
			return 1, &result, nil
		}

		n, err = strconv.ParseInt(args[1], 10, 64)
		if err != nil {
			return 2, &result, NewParseError(BeNotAnInt, StringFromCode(BeNotAnInt), args[1], a)
		}
		result.Value = int(n)
		result.ParamCount = 1
		return 2, &result, nil
	case TypeFloat:
		var n float64
		if len(args) <= 1 {
			if !result.ParamOpt {
				return 2, &result, NewParseError(BeNoRequiredFloat, StringFromCode(BeNoRequiredFloat), a)
			}
			return 2, &result, nil
		}

		n, _ = strconv.ParseFloat(args[1], 64)
		if err != nil {
			return 2, &result, NewParseError(BeNotAFloat, StringFromCode(BeNotAFloat), args[1], a)
		}
		result.Value = n
		result.ParamCount = 1
		return 2, &result, nil
	case TypeString:
		if len(args) <= 1 {
			if !result.ParamOpt {
				return 1, &result, NewParseError(BeNoRequiredString, StringFromCode(BeNoRequiredString), a)
			}
			return 1, &result, nil
		}
		result.Value = args[1]
		result.ParamCount = 1
		return 2, &result, nil
	case TypeEmail:
		if len(args) <= 1 {
			if !result.ParamOpt {
				return 1, &result, NewParseError(BeNoRequiredEmail, StringFromCode(BeNoRequiredEmail), a)
			}
			return 1, &result, nil
		}
		email, err := mail.ParseAddress(args[1])
		if err != nil {
			return 2, &result, NewParseError(BeNotAnEmail, StringFromCode(BeNotAnEmail), args[1], a)
		}
		result.Value = *email
		result.ParamCount = 1
		return 2, &result, nil
	case TypePhone:
		if len(args) <= 1 {
			if !result.ParamOpt {
				return 1, &result, NewParseError(BeNoRequiredPhone, StringFromCode(BeNoRequiredEmail), a)
			}
			return 1, &result, nil
		}
		re := regexp.MustCompile(`^(?:(?:\(?(?:00|\+)([1-4]\d\d|[1-9]\d?)\)?)?[\-\.\ \\\/]?)?((?:\(?\d{1,}\)?[\-\.\ \\\/]?){0,})(?:[\-\.\ \\\/]?(?:#|ext\.?|extension|x)[\-\.\ \\\/]?(\d+))?$`)
		b := re.MatchString(args[1])
		if !b {
			return 2, &result, NewParseError(BeNotAPhone, StringFromCode(BeNotAPhone), args[1], a)
		}
		result.Value = args[1]
		result.ParamCount = 1
		return 2, &result, nil
	case TypeTime:
		if len(args) <= 1 {
			if !result.ParamOpt {
				return 1, &result, NewParseError(BeNoRequiredTime, StringFromCode(BeNoRequiredTime), a)
			}
			return 1, &result, nil
		}
		timeval, err := time.Parse(time.Kitchen, args[1])
		if err != nil {
			return 2, &result, NewParseError(BeNotATime, StringFromCode(BeNotATime), args[1], a)
		}
		result.Value = timeval
		result.ParamCount = 1
		return 2, &result, nil
	case TypeTimeDuration:
		if len(args) <= 1 {
			if !result.ParamOpt {
				return 1, &result, NewParseError(BeNoRequiredDuration, StringFromCode(BeNoRequiredDuration), a)
			}
			return 1, &result, nil
		}
		duration, err := time.ParseDuration(args[1])
		if err != nil {
			return 2, &result, NewParseError(BeNotADuration, StringFromCode(BeNotADuration), args[1], a)
		}
		result.Value = duration
		result.ParamCount = 1
		return 2, &result, nil
	case TypeDate:
		if len(args) <= 1 {
			if !result.ParamOpt {
				return 1, &result, NewParseError(BeNoRequiredDate, StringFromCode(BeNoRequiredDate), a)
			}
			return 1, &result, nil
		}
		// below is the way the reference date would be represented
		// in the desired layout; it has no time zone present.
		// Note: without explicit zone, returns time in UTC.
		const format = "Jan-02-2006"
		dateval, err := time.Parse(format, args[1])
		if err != nil {
			return 2, &result, NewParseError(BeNotADate, StringFromCode(BeNotADate), args[1], a)
		}
		result.Value = dateval
		result.ParamCount = 1
		return 2, &result, nil
	case TypePath:
		if len(args) <= 1 {
			if !result.ParamOpt {
				return 1, &result, NewParseError(BeNoRequiredPath, StringFromCode(BeNoRequiredPath), a)
			}
			return 1, &result, nil
		}
		path, err := abspath.ExpandFrom(args[1])
		if err != nil {
			return 2, &result, NewParseError(BeNotAPath, StringFromCode(BeNotAPath), args[1], a)
		}
		result.Value = path.String()
		result.ParamCount = 1
		return 2, &result, nil
	case TypeURL:
		if len(args) <= 1 {
			if !result.ParamOpt {
				return 1, &result, NewParseError(BeNoRequiredURL, StringFromCode(BeNoRequiredURL), a)
			}
			return 1, &result, nil
		}
		url, err := url.ParseRequestURI(args[1])
		if err != nil {
			return 2, &result, NewParseError(BeNotAURL, StringFromCode(BeNotAURL), args[1], a)
		}
		result.Value = url
		result.ParamCount = 1
		return 2, &result, nil
	case TypeIPv4:
		if len(args) <= 1 {
			if !result.ParamOpt {
				return 1, &result, NewParseError(BeNoRequiredIPv4, StringFromCode(BeNoRequiredIPv4), a)
			}
			return 1, &result, nil
		}
		ip := net.ParseIP(args[1])
		if ip == nil {
			return 2, &result, NewParseError(BeNotAnIPv4, StringFromCode(BeNotAnIPv4), args[1], a)
		}
		result.Value = ip
		result.ParamCount = 1
		return 2, &result, nil

	// slice types-------------------------

	case TypeIntSlice:
		var j int
		var val []int
		if len(args) <= 1 {
			if !result.ParamOpt {
				return 1, &result, NewParseError(BeNoRequiredInt, StringFromCode(BeNoRequiredInt), a)
			}
			return 1, &result, nil
		}
		for i, arg := range args[1:] {
			if arg == "--" {
				j = i + 1
				break
			}
			n, err := strconv.ParseInt(arg, 10, 64)
			if err != nil {
				return 2, &result, NewParseError(BeNotAnInt, StringFromCode(BeNotAnInt), arg, a)
			}
			val = append(val, int(n))
			j = i + 1
		}
		result.Value = val
		result.ParamCount = j
		return j + 1, &result, nil
	case TypeFloatSlice:
		var j int
		var val []float64
		if len(args) <= 1 {
			if !result.ParamOpt {
				return 1, &result, NewParseError(BeNoRequiredFloat, StringFromCode(BeNoRequiredFloat), a)
			}
			return 1, &result, nil
		}
		for i, arg := range args[1:] {
			if arg == "--" {
				j = i + 1
				break
			}
			n, err := strconv.ParseFloat(arg, 64)
			if err != nil {
				return 2, &result, NewParseError(BeNotAFloat, StringFromCode(BeNotAFloat), arg, a)
			}
			val = append(val, n)
			j = i + 1
		}
		result.Value = val
		result.ParamCount = j
		return j + 1, &result, nil
	case TypeStringSlice:
		var j int
		var val []string
		if len(args) <= 1 {
			if !result.ParamOpt {
				return 1, &result, NewParseError(BeNoRequiredString, StringFromCode(BeNoRequiredString), a)
			}
			return 1, &result, nil
		}
		for i, arg := range args[1:] {
			if arg == "--" {
				j = i + 1
				break
			}
			val = append(val, arg)
			j = i + 1
		}
		result.Value = val
		result.ParamCount = j
		return j + 1, &result, nil
	case TypeEmailSlice:
		var j int
		var val []mail.Address
		if len(args) <= 1 {
			if !result.ParamOpt {
				return 1, &result, NewParseError(BeNoRequiredEmail, StringFromCode(BeNoRequiredEmail), a)
			}
			return 1, &result, nil
		}
		for i, arg := range args[1:] {
			if arg == "--" {
				j = i + 1
				break
			}
			email, err := mail.ParseAddress(arg)
			if err != nil {
				return 1, &result, NewParseError(BeNotAnEmail, StringFromCode(BeNotAnEmail), arg, a)
			}
			val = append(val, *email)
			j = i + 1
		}
		result.Value = val
		result.ParamCount = j
		return j + 1, &result, nil
	case TypePhoneSlice:
		var j int
		var val []string
		if len(args) <= 1 {
			if !result.ParamOpt {
				return 1, &result, NewParseError(BeNoRequiredPhone, StringFromCode(BeNoRequiredPhone), a)
			}
			return 1, &result, nil
		}

		for i, arg := range args[1:] {
			if arg == "--" {
				j = i + 1
				break
			}
			re := regexp.MustCompile(`^(?:(?:\(?(?:00|\+)([1-4]\d\d|[1-9]\d?)\)?)?[\-\.\ \\\/]?)?((?:\(?\d{1,}\)?[\-\.\ \\\/]?){0,})(?:[\-\.\ \\\/]?(?:#|ext\.?|extension|x)[\-\.\ \\\/]?(\d+))?$`)
			b := re.MatchString(args[1])
			if !b {
				return 2, &result, NewParseError(BeNotAPhone, StringFromCode(BeNotAPhone), arg, a)
			}
			val = append(val, arg)
			j = i + 1
		}
		result.Value = val
		result.ParamCount = j
		return j + 1, &result, nil
	case TypeTimeSlice:
		var j int
		var val []time.Time
		if len(args) <= 1 {
			if !result.ParamOpt {
				return 1, &result, NewParseError(BeNoRequiredTime, StringFromCode(BeNoRequiredTime), a)
			}
			return 1, &result, nil
		}

		for i, arg := range args[1:] {
			if arg == "--" {
				j = i + 1
				break
			}
			timeval, err := time.Parse(time.Kitchen, arg)
			if err != nil {
				return 2, &result, NewParseError(BeNotATime, StringFromCode(BeNotATime), arg, a)
			}
			val = append(val, timeval)
			j = i + 1
		}
		result.Value = val
		result.ParamCount = j
		return j + 1, &result, nil
	case TypeTimeDurationSlice:
		var j int
		var val []time.Duration
		if len(args) <= 1 {
			if !result.ParamOpt {
				return 1, &result, NewParseError(BeNoRequiredDuration, StringFromCode(BeNoRequiredDuration), a)
			}
			return 1, &result, nil
		}

		for i, arg := range args[1:] {
			if arg == "--" {
				j = i + 1
				break
			}
			duration, err := time.ParseDuration(arg)
			if err != nil {
				return 2, &result, NewParseError(BeNotADuration, StringFromCode(BeNotADuration), arg, a)
			}
			val = append(val, duration)
			j = i + 1
		}
		result.Value = val
		result.ParamCount = j
		return j + 1, &result, nil
	case TypeDateSlice:
		var j int
		var val []time.Time
		if len(args) <= 1 {
			if !result.ParamOpt {
				return 1, &result, NewParseError(BeNoRequiredDate, StringFromCode(BeNoRequiredDate), a)
			}
			return 1, &result, nil
		}

		for i, arg := range args[1:] {
			if arg == "--" {
				j = i + 1
				break
			}
			// below is the way the reference date would be represented
			// in the desired layout; it has no time zone present.
			// Note: without explicit zone, returns time in UTC.
			const format = "Jan-02-2006"
			dateval, err := time.Parse(format, arg)
			if err != nil {
				return 2, &result, NewParseError(BeNotADate, StringFromCode(BeNotADate), arg, a)

			}
			val = append(val, dateval)
			j = i + 1

		}
		result.Value = val
		result.ParamCount = j
		return j + 1, &result, nil
	case TypePathSlice:
		var j int
		var val []string
		if len(args) <= 1 {
			if !result.ParamOpt {
				return 1, &result, NewParseError(BeNoRequiredPath, StringFromCode(BeNoRequiredPath), a)
			}
			return 1, &result, nil
		}

		for i, arg := range args[1:] {
			if arg == "--" {
				j = i + 1
				break
			}
			path, err := abspath.ExpandFrom(arg)
			if err != nil {
				return 2, &result, NewParseError(BeNotAPath, StringFromCode(BeNotAPath), arg, a)
			}
			val = append(val, path.String())
			j = i + 1
		}
		result.Value = val
		result.ParamCount = j
		return j + 1, &result, nil
	case TypeURLSlice:
		var j int
		var val []url.URL
		if len(args) <= 1 {
			if !result.ParamOpt {
				return 1, &result, NewParseError(BeNoRequiredURL, StringFromCode(BeNoRequiredURL), a)
			}
			return 1, &result, nil
		}

		for i, arg := range args[1:] {
			if arg == "--" {
				j = i + 1
				break
			}
			url, err := url.ParseRequestURI(args[1])
			if err != nil {
				return 2, &result, NewParseError(BeNotAURL, StringFromCode(BeNotAURL), arg, a)
			}
			val = append(val, *url)
			j = i + 1
		}
		result.Value = val
		result.ParamCount = j
		return j + 1, &result, nil
	case TypeIPv4Slice:
		var j int
		var val []net.IP
		if len(args) <= 1 {
			if !result.ParamOpt {
				return 1, &result, NewParseError(BeNoRequiredIPv4, StringFromCode(BeNoRequiredIPv4), a)
			}
			return 1, &result, nil
		}

		for i, arg := range args[1:] {
			if arg == "--" {
				j = i + 1
				break
			}
			ip := net.ParseIP(arg)
			if err != nil {
				return 2, &result, NewParseError(BeNotAnIPv4, StringFromCode(BeNotAnIPv4), arg, a)
			}
			val = append(val, ip)
			j = i + 1
		}
		result.Value = val
		result.ParamCount = j
		return j + 1, &result, nil

	default:
	}
	return 1, nil, nil
}
