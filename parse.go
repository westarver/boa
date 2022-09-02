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
<<<<<<< HEAD
		paramType:   cm.paramType,
		value:       nil,
=======
		Type:        cm.Type,
>>>>>>> ddb2b57a0cb42366fc393fe1a983ecea453ad4b8
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
<<<<<<< HEAD
		id:          cm.id,
		paramOpt:    cm.paramOpt,
		paramCount:  0,
	}

	var advance = 1 // we eat at least the command/flag name
	var err error
	switch cm.paramType {
	case TypeBool:
		result.value = true
		result.paramCount = 0
		return advance, &result
	case TypeInt:
		var n int64
		if len(args) <= 1 {
			result.value = int(0)
			result.paramCount = 0
			return advance, &result
		}
		advance = 2
		n, err = strconv.ParseInt(args[1], 10, 64)
		if err != nil { // should we return error?
			n = 0
			advance = 1
		}
		result.value = int(n)
		result.paramCount = 1
		return advance, &result
	case TypeFloat:
		var n float64
		if len(args) <= 1 {
			result.value = 0.0
			result.paramCount = 0
			return advance, &result
		}
		advance = 2
		n, _ = strconv.ParseFloat(args[1], 64)
		if err != nil { // should we return error?
			n = 0.0
			advance = 1
		}
		result.value = n
		result.paramCount = 1
		return advance, &result
	case TypeString:
		if len(args) <= 1 {
			result.value = ""
			result.paramCount = 0
			return 1, &result
		}
		result.value = args[1]
		result.paramCount = 1
		return 2, &result
	case TypeIntSlice:
=======
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
>>>>>>> ddb2b57a0cb42366fc393fe1a983ecea453ad4b8
		var j int
		var val []int
		if len(args) <= 1 {
			result.value = []int{}
			result.paramCount = 0
			return 1, &result
		}
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
<<<<<<< HEAD
		result.paramCount = j
		return j + 1, &result
	case TypeFloatSlice:
=======
		return j + 1, &result
	case FloatSliceType:
>>>>>>> ddb2b57a0cb42366fc393fe1a983ecea453ad4b8
		var j int
		var val []float64
		if len(args) <= 1 {
			result.value = []float64{}
			result.paramCount = 0
			return 1, &result
		}
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
<<<<<<< HEAD
		result.paramCount = j
		return j + 1, &result
	case TypeStringSlice:
=======
		return j + 1, &result
	case StringSliceType:
>>>>>>> ddb2b57a0cb42366fc393fe1a983ecea453ad4b8
		var j int
		var val []string
		if len(args) <= 1 {
			result.value = []string{}
			result.paramCount = 0
			return 1, &result
		}
		for i, a := range args[1:] {
			if a == "--" {
				j = i + 1
				break
			}
			val = append(val, a)
			j = i + 1
		}
		result.value = val
<<<<<<< HEAD
		result.paramCount = j
=======
>>>>>>> ddb2b57a0cb42366fc393fe1a983ecea453ad4b8
		return j + 1, &result
	default:
	}
	return 1, nil
}
