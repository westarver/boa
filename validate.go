package boa

// validateRequirements ts called after all the commands
// and flags have been parsed. The last thing to do before
// returning the filled CLI structure is to make sure the
// requirements and restrictions proscribed in the input
// script are met.
func validateRequirements(cmds map[string]CmdLineItem, cli *CLI) {
	for _, it := range cmds {
		if it.Required {
			_, found := cli.Items[it.Name]
			if !found {
				cli.SetError(NewParseError(BeNoRequiredItem, StringFromCode(BeNoRequiredItem), it.Name))
			}
		}
		if it.Exclusive {
			for _, i := range cli.Items {
				if it.Id == i.Id {
					continue
				}

				if it.IsFlag && i.IsFlag {
					cli.SetError(NewParseError(BeNoExclusiveItem, StringFromCode(BeNoExclusiveItem), it.Name, i.Name))
					return
				}

				if !i.IsFlag && i.IsFlag {
					cli.SetError(NewParseError(BeNoExclusiveItem, StringFromCode(BeNoExclusiveItem), it.Name, i.Name))
					break
				}
			}
		}
	}
}
