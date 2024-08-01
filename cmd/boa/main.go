package main

import (
	"fmt"

	"github.com/westarver/boa"
)

func main() {
	var args = []string{"name", "--first", "Jimmy", "-abc", "--", "--topping=choc"}
	cmds := make(map[string]boa.CmdLineItem)
	cmds["name"] = boa.CmdLineItem{Name: "name", ParamType: boa.TypeString, ParamCount: -1, DefaultValue: "James", ChNames: []string{"--first"}}
	cmds["-a"] = boa.CmdLineItem{Name: "-a", ParamType: boa.TypeBool, ParamCount: 0}
	cmds["-b"] = boa.CmdLineItem{Name: "-b", ParamType: boa.TypeBool, ParamCount: 0}
	cmds["-c"] = boa.CmdLineItem{Name: "-c", ParamType: boa.TypeInt, ParamCount: -1, DefaultValue: "999"} //[]int{1, 2, 3}}
	cmds["--topping"] = boa.CmdLineItem{Name: "--topping", ParamType: boa.TypeString, ParamCount: 1}
	cli := boa.ParseCommandLineArgs(cmds, args)
	boa.PrintCli(cli)
	fmt.Println(cli.Errors())
}
