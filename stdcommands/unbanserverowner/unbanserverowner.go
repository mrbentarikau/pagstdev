package unbanserverowner

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
	Name:                 "unbanserverowner",
	Description:          ";))",
	HideFromHelp:         true,
	RequiredArgs:         1,
	Arguments: []*dcmd.ArgDef{
		{Name: "server", Type: dcmd.String},
	},
	RunFunc: util.RequireOwner(func(data *dcmd.Data) (interface{}, error) {

		var unbanned bool
		err := common.RedisPool.Do(radix.Cmd(&unbanned, "SREM", "banned_server_owners", data.Args[0].Str()))
		if err != nil {
			return nil, err
		}

		if !unbanned {
			return "User wasn't banned", nil
		}

		return "Unbanned owner", nil
	}),
}
