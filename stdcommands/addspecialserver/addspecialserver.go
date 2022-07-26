package addspecialserver

import (
	"github.com/botlabs-gg/yagpdb/commands"
	"github.com/botlabs-gg/yagpdb/common"
	"github.com/botlabs-gg/yagpdb/lib/dcmd"
	"github.com/botlabs-gg/yagpdb/stdcommands/util"
	"github.com/mediocregopher/radix/v3"
)

var Command = &commands.YAGCommand{
	Cooldown:             2,
	CmdCategory:          commands.CategoryDebug,
	HideFromCommandsPage: true,
	Name:                 "addspecialserver",
	Description:          ";))",
	HideFromHelp:         true,
	RequiredArgs:         1,
	Arguments: []*dcmd.ArgDef{
		{Name: "server", Type: dcmd.Int},
	},
	RunFunc: util.RequireOwner(func(data *dcmd.Data) (interface{}, error) {
		var whitelisted bool
		err := common.RedisPool.Do(radix.FlatCmd(&whitelisted, "SADD", "special_servers", data.Args[0].Int64()))
		if err != nil {
			return "", err
		}

		if !whitelisted {
			return "Server was already whitelisted", nil
		}
		return "Whitelisted: " + data.Args[0].Str(), nil
	}),
}
