package boa

import (
	"fmt"
	"os"
	"reflect"
	"strconv"

	trace "github.com/westarver/tracer"
)

func ParseCommandLineArgs(cmds map[string]cmdLineArg, args []string) *CLI {
	//var trace = trace.New(os.Stderr)                                 //<rmv/>
	//trace.Trace("---------------entering ParseCommandLineArgs")      //<rmv/>````````
	//defer trace.Trace("---------------leaving ParseCommandLineArgs") //<rmv/>
	var cli = CLI{Items: make(map[string]any, len(args))}
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
		//a = strings.TrimPrefix(a, "--")
		args[n] = a // in case the alias was normalized the proper value must be passed to getCmdValues
		//trace.Trace("args[", n, "] ", a) //<rmv/>

		m, cm := getCmdValues(cmds, a, args[n:])
		//trace.Trace("gobbled up ", m) //<rmv/>
		n += m
		//trace.Trace("n = ", n) //<rmv/>
		if cm != nil {
			cli.Items[cm.(GenericArg).Name()] = cm
		} else {
			cli.SetError(fmt.Errorf("%s - invalid command line parameter", a))
		}
		if n >= len(args) {
			break
		}
	}
	return &cli
}

func getCmdValues(cmds map[string]cmdLineArg, a string, args []string) (int, any) {
	var trace = trace.New(os.Stderr)                             //<rmv/>
	trace.Trace("--------------------entering getCmdValues")     //<rmv/>
	defer trace.Trace("-------------------leaving getCmdValues") //<rmv/>

	intType := reflect.TypeOf(int64(0))
	boolType := reflect.TypeOf(true)
	floatType := reflect.TypeOf(float64(0.0))
	stringType := reflect.TypeOf("")
	intSliceType := reflect.TypeOf([]int64{})
	floatSliceType := reflect.TypeOf([]float64{})
	stringSliceType := reflect.TypeOf([]string{})

	cm, exist := cmds[a]
	if !exist {
		cm, exist = cmds["--"+a]
		if !exist {
			return 1, nil
		}
	}

	if len(args) < 1 {
		//trace.Trace("returning nil len < 1") //<rmv/>
		return 1, nil
	}

	switch cm.Type {
	case boolType:
		//trace.Trace("matched bool ", a) //<rmv/>
		cm1 := CmdLineItem[bool]{
			value:       true,
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
		//trace.Trace("\ncmd line item ", cm1) //<rmv/>
		return 1, cm1
	case intType:
		//trace.Trace("matched int ", intType) //<rmv/>
		n, _ := strconv.ParseInt(args[1], 10, 64)
		cm1 := CmdLineItem[int]{
			value:       int(n),
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
		//trace.Trace("\ncmd line item ", cm1) //<rmv/>
		return 2, cm1
	case floatType:
		//trace.Trace("matched float ", floatType) //<rmv/>
		n, _ := strconv.ParseFloat(args[1], 64)
		cm1 := CmdLineItem[float64]{
			value:       n,
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
		//trace.Trace("\ncmd line item ", cm1) //<rmv/>
		return 2, cm1
	case stringType:
		//trace.Trace("matched string ", stringType) //<rmv/>
		var str = ""
		if len(args) > 1 {
			str = args[1]
		}
		cm1 := CmdLineItem[string]{
			value:       str,
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
		//trace.Trace("\ncmd line item ", cm1) //<rmv/>
		return 2, cm1
	case intSliceType:
		//trace.Trace("matched []int ", intSliceType) //<rmv/>
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
		cm1 := CmdLineItem[[]int]{
			value:       val,
			name:        cm.name,
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
		//trace.Trace("\ncmd line item ", cm1) //<rmv/>
		return j + 1, cm1
	case floatSliceType:
		//trace.Trace("matched []float ", floatSliceType) //<rmv/>
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
		cm1 := CmdLineItem[[]float64]{
			value:       val,
			name:        cm.name,
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
		//trace.Trace("\ncmd line item ", cm1) //<rmv/>
		return j + 1, cm1
	case stringSliceType:
		//trace.Trace("matched []string ", stringSliceType) //<rmv/>
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
		cm1 := CmdLineItem[[]string]{
			value:       val,
			name:        cm.name,
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
		//trace.Trace("\ncmd line item ", cm1) //<rmv/>
		return j + 1, cm1

	default:
	}
	//trace.Trace("returning nil no match") //<rmv/>
	return 1, nil
}
