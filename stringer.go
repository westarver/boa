package boa

import "fmt"

func TypeToString(p ParameterType) string {
	switch p {
	case TypeString, TypeStringSlice:
		return "String"
	case TypeInt, TypeIntSlice:
		return "Integer"
	case TypeFloat, TypeFloatSlice:
		return "Float"
	case TypeTime, TypeTimeSlice:
		return "Time"
	case TypeDate, TypeDateSlice:
		return "Date"
	case TypeTimeDuration, TypeTimeDurationSlice:
		return "TimeDuration"
	case TypePath, TypePathSlice:
		return "File Path"
	case TypeURL, TypeURLSlice:
		return "URL"
	case TypeEmail, TypeEmailSlice:
		return "Email Address"
	case TypeIPv4, TypeIPv4Slice:
		return "IPv4 Address"
	}
	return "Bool"
}

func PrintCli(cli *CLI) {
	fmt.Println("App:", cli.Application)
	PrintMap(cli.Items)
}

func PrintMap(items map[string]CmdLineItem) {
	for _, it := range items {
		PrintItem(it)
	}
}

func PrintItem(it CmdLineItem) {
	fmt.Println("\nID", it.Id)
	fmt.Println("Name", it.Name)
	fmt.Println("Short", it.ShortHelp)
	fmt.Println("Long", it.LongHelp)
	fmt.Println("Type", TypeToString(it.ParamType))
	fmt.Println("Count", it.ParamCount)
	fmt.Println("Optional", it.ParamOpt)
	fmt.Println("IsFlag", it.IsFlag)
	fmt.Println("--------------------")
	fmt.Println("")
}
