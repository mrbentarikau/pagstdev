package info

import (
	"fmt"

	"github.com/botlabs-gg/yagpdb/commands"
	"github.com/botlabs-gg/yagpdb/common"
	"github.com/botlabs-gg/yagpdb/lib/dcmd"
)

var Command = &commands.YAGCommand{
	CmdCategory: commands.CategoryGeneral,
	Name:        "Info",
	Description: "Responds with bot information",
	RunInDM:     true,
	RunFunc: func(data *dcmd.Data) (interface{}, error) {

		info := fmt.Sprintf(`**PAGSTDB - People Against Generally Shitty Things Discord Bot**

Control panel: <https://%s/manage>
				`, common.ConfHost.GetString())

		return info, nil
	},
}
