package boa

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"

	trace "github.com/westarver/tracer"
)

func parseCommandLineArgs(cmds map[string]cmdLineArg, args []string) *CLI {
	var trace = trace.New(os.Stderr)                                 //<rmv/>
	trace.Trace("---------------entering parseCommandLineArgs")      //<rmv/>````````
	defer trace.Trace("---------------leaving parseCommandLineArgs") //<rmv/>
	var cli = CLI{items: make(map[string]any, len(args))}
	trace.Trace("cmd line as passed ", args) //<rmv/>
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
		a = strings.TrimPrefix(a, "--")
		args[n] = a                        // in case the alias was normalized the proper value must be passed to getCmdValues
		trace.Trace("args[n] ", a, " ", n) //<rmv/>

		m, cm := getCmdValues(cmds, a, args[n:])
		trace.Trace("gobbled up ", m) //<rmv/>
		n += m
		trace.Trace("n = ", n) //<rmv/>
		if cm != nil {
			cli.items[cm.(GenericArg).Name()] = cm
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

	trace.Trace("args", args)           //<rmv/>
	trace.Trace("len args ", len(args)) //<rmv/>

	cm, exist := cmds[a]
	if !exist {
		trace.Trace("returning cmds[a] not exist") //<rmv/>
		return 1, nil
	}

	t := cm.Type
	trace.Trace("type ", t) //<rmv/>

	if len(args) < 1 {
		trace.Trace("returning nil len < 1") //<rmv/>
		return 1, nil
	}

	if len(args) > 1 {
		if args[1] == "--" {
			trace.Trace("returning nil found --") //<rmv/>
			return 2, nil
		}
	}
	switch t {
	case boolType:
		trace.Trace("matched bool ", boolType) //<rmv/>
		cm1 := CmdLineItem[bool]{value: true}
		trace.Trace("cmd line item ", cm1) //<rmv/>
		return 1, cm1
	case intType:
		trace.Trace("matched int ", intType) //<rmv/>
		n, _ := strconv.ParseInt(args[1], 10, 64)
		cm1 := CmdLineItem[int64]{value: n}
		trace.Trace("cmd line item ", cm1) //<rmv/>
		return 2, cm1
	case floatType:
		trace.Trace("matched float ", floatType) //<rmv/>
		n, _ := strconv.ParseFloat(args[1], 64)
		cm1 := CmdLineItem[float64]{value: n}
		trace.Trace("cmd line item ", cm1) //<rmv/>
		return 2, cm1
	case stringType:
		trace.Trace("matched string ", stringType) //<rmv/>
		cm1 := CmdLineItem[string]{value: args[1]}
		cm1.value = args[1]
		trace.Trace("cmd line item ", cm1) //<rmv/>
		return 2, cm1
	case intSliceType:
		trace.Trace("matched []int ", intSliceType) //<rmv/>
		var j int
		cm1 := CmdLineItem[[]int64]{}
		for i, a := range args[1:] {
			if a == "--" {
				j = i + 1
				break
			}
			n, _ := strconv.ParseInt(a, 10, 64)
			cm1.value = append(cm1.value, n)
			j = i + 1
		}
		trace.Trace("cmd line item ", cm1) //<rmv/>
		return j + 1, cm1
	case floatSliceType:
		trace.Trace("matched []float ", floatSliceType) //<rmv/>
		var j int
		cm1 := CmdLineItem[[]float64]{}
		for i, a := range args[1:] {
			if a == "--" {
				j = i + 1
				break
			}
			n, _ := strconv.ParseFloat(a, 64)
			cm1.value = append(cm1.value, n)
			j = i + 1
		}
		trace.Trace("cmd line item ", cm1) //<rmv/>
		return j + 1, cm1
	case stringSliceType:
		trace.Trace("matched []string ", stringSliceType) //<rmv/>
		var j int
		cm1 := CmdLineItem[[]string]{}
		for i, a := range args[1:] {
			if a == "--" {
				j = i + 1
				break
			}
			cm1.value = append(cm1.value, a)
			j = i + 1
		}
		trace.Trace("cmd line item ", cm1) //<rmv/>
		return j + 1, cm1
	default:
	}

	trace.Trace("returning nil no match") //<rmv/>
	return 1, nil
}
