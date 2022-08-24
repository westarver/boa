package boa

import (
	"fmt"
	"strconv"
)

func ParseCommandLineArgs(cmds map[string]CmdLineItem, args []string) *CLI {
	var cli = CLI{Items: make(map[string]CmdLineItem, len(args))}
	var n = 0
	for i := 0; i < len(args); i++ {
		a := args[n]
		// deal with alias passed in
		for _, c := range cmds {
			if c.alias == a {
				a = c.name
				break
			}
		}

		args[n] = a // in case the alias was normalized the proper value must be passed to getCmdValues

		m, cm := getCmdValues(cmds, a, args[n:])
		n += m
		if cm != nil {
			cli.Items[cm.Name()] = *cm
		} else {
			cli.SetError(fmt.Errorf("%s - invalid command line parameter", a))
		}
		if n >= len(args) {
			break
		}
	}
	return &cli
}

func getCmdValues(cmds map[string]CmdLineItem, a string, args []string) (int, *CmdLineItem) {
	cm, exist := cmds[a]
	if !exist {
		cm, exist = cmds["--"+a]
		if !exist {
			return 1, nil
		}
	}

	if len(args) < 1 {
		return 1, nil
	}

	result := CmdLineItem{
		Type:        cm.Type,
		name:        a,
		alias:       cm.alias,
		shortHelp:   cm.shortHelp,
		longHelp:    cm.longHelp,
		isDefault:   cm.isDefault,
		isFlag:      cm.isFlag,
		exclusive:   cm.exclusive,
		required:    cm.required,
		requiredOr:  cm.requiredOr,
		requiredAnd: cm.requiredAnd,
	}

	switch cm.Type {
	case BoolType:
		result.value = true
		return 1, &result
	case IntType:
		n, _ := strconv.ParseInt(args[1], 10, 64)
		result.value = int(n)
		return 2, &result
	case FloatType:
		n, _ := strconv.ParseFloat(args[1], 64)
		result.value = n
		return 2, &result
	case StringType:
		var str = ""
		if len(args) > 1 {
			str = args[1]
		}
		result.value = str
		return 2, &result
	case IntSliceType:
		var j int
		var val []int
		for i, a := range args[1:] {
			if a == "--" {
				j = i + 1
				break
			}
			n, _ := strconv.ParseInt(a, 10, 64)
			val = append(val, int(n))
			j = i + 1
		}
		result.value = val
		return j + 1, &result
	case FloatSliceType:
		var j int
		var val []float64
		for i, a := range args[1:] {
			if a == "--" {
				j = i + 1
				break
			}
			n, _ := strconv.ParseFloat(a, 64)
			val = append(val, n)
			j = i + 1
		}
		result.value = val
		return j + 1, &result
	case StringSliceType:
		var j int
		var val []string
		for i, a := range args[1:] {
			if a == "--" {
				j = i + 1
				break
			}
			val = append(val, a)
			j = i + 1
		}
		result.value = val
		return j + 1, &result
	default:
	}
	return 1, nil
}
