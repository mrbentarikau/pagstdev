package thanks

import (
	"fmt"

	"github.com/botlabs-gg/yagpdb/commands"
	"github.com/botlabs-gg/yagpdb/lib/dcmd"
	"github.com/botlabs-gg/yagpdb/lib/dstate"
	"github.com/botlabs-gg/yagpdb/reputation"
)

var Command = &commands.YAGCommand{
	CmdCategory:          commands.CategoryFun,
	Name:                 "Thanks",
	CmdNoRun:             true,
	RunInDM:              false,
	HideFromHelp:         true,
	HideFromCommandsPage: true,
	DefaultEnabled:       false,

	SlashCommandEnabled:    true,
	ApplicationCommandType: 2,

	Arguments: []*dcmd.ArgDef{
		{Name: "User", Type: &commands.MemberArg{}},
	},

	RunFunc: func(data *dcmd.Data) (interface{}, error) {
		sender := data.GuildData.MS
		target := data.Args[0].Value.(*dstate.MemberState)

		if sender.User.ID == target.User.ID {
			return nil, nil
		}

		conf, err := reputation.GetConfig(data.Context(), data.GuildData.GS.ID)
		if err != nil {
			return "An error occurred while finding the server config", err
		}

		if err = reputation.CanModifyRep(conf, sender, target); err != nil {
			return nil, err
		}

		err = reputation.ModifyRep(data.Context(), conf, data.GuildData.GS.ID, sender, target, 1)
		if err != nil {
			if err == reputation.ErrCooldown {
				// Ignore this error silently
				return nil, nil
			}
			err = commands.NewPublicError("Failed giving rep")
			return nil, err
		}

		content := fmt.Sprintf("Gave +1 %s to **%s**", conf.PointsName, target.User.Mention())
		return content, nil
	},
}
