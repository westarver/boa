package boa

// validateRequirements ts called after all the commands
// and flags have been parsed.
func validateRequirements(cmds map[string]CmdLineItem, cli *CLI) {
	for _, it := range cmds {
		if it.IsRequired {
			_, found := cli.Items[it.Name]
			if !found {
				cli.SetError(Errorf(BeNoRequiredItem, it.Name))
			}
		}
		if it.IsExclusive {
			for _, i := range cli.Items {
				if it.Id == i.Id {
					continue
				}

				if it.IsFlag && i.IsFlag {
					cli.SetError(Errorf(BeNoExclusiveItem, it.Name, i.Name))
					return
				}

				if !i.IsFlag && i.IsFlag {
					cli.SetError(Errorf(BeNoExclusiveItem, it.Name, i.Name))
					break
				}
			}
		}
	}
}
