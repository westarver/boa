package boa

import (
	"encoding/json"
	"strings"
)

func AppDataName() string {
	return "BOA-APP-DATA"
}

func FromJSON(json []byte, args []string) *CLI {
	items, err := CollectItemsFromJSON(json)
	if err != nil {
		return nil
	}

	var app string

	// get rid of the app-data record before passing *CLI to caller
	// if an app needs this record it can be obtained by calling
	// CollectItemsFromJSON directly
	if appdata, ok := items[AppDataName()]; ok {
		app = appdata.Alias //app name is in alias field in that special item
		delete(items, AppDataName())
	}
	cli := ParseCommandLineArgs(items, args)
	if cli == nil {
		return nil
	}
	cli.Application = app

	validateRequirements(items, cli)
	cli.AllHelp = make(map[string]string)
	for _, item := range items {
		cli.AllHelp[item.Name] = formatHelp(item.Name, item.Alias, item.ShortHelp, item.LongHelp)
	}

	return cli
}

type sliceWrap struct {
	Commands []CmdLineItem `json:"commands"`
}

func CollectItemsFromJSON(jsonBytes []byte) (map[string]CmdLineItem, error) {
	var jslice sliceWrap

	if err := json.Unmarshal(jsonBytes, &jslice); err != nil {
		return nil, err
	}

	jmap := make(map[string]CmdLineItem)
	for _, v := range jslice.Commands {
		jmap[v.Name] = v
	}
	return jmap, nil
}

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
