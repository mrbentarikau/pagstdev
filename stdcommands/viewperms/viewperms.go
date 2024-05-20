package viewperms

import (
	"fmt"
	"strings"

	"github.com/mrbentarikau/pagst/bot"
	"github.com/mrbentarikau/pagst/commands"
	"github.com/mrbentarikau/pagst/common"
	"github.com/mrbentarikau/pagst/lib/dcmd"
	"github.com/mrbentarikau/pagst/lib/discordgo"
	"github.com/mrbentarikau/pagst/lib/dstate"
)

var Command = &commands.YAGCommand{
	CmdCategory: commands.CategoryDebug,
	Name:        "ViewPerms",
	Description: "Shows you or the target's permissions in a given channel (default current channel)",
	Arguments: []*dcmd.ArgDef{
		{Name: "target", Type: dcmd.UserID, Default: int64(0)},
	},
	ArgSwitches: []*dcmd.ArgDef{
		{Name: "channel", Type: dcmd.Channel},
	},
	RunFunc: func(data *dcmd.Data) (interface{}, error) {
		var target *dstate.MemberState

		if targetID := data.Args[0].Int64(); targetID == 0 {
			target = data.GuildData.MS
		} else {
			var err error
			target, err = bot.GetMember(data.GuildData.GS.ID, targetID)
			if err != nil {
				if common.IsDiscordErr(err, discordgo.ErrCodeUnknownMember) {
					return "Unknown member", nil
				}

				return nil, err
			}
		}

		cID := data.ChannelID
		if c := data.Switch("channel"); c.Value != nil {
			cID = c.Value.(*dstate.ChannelState).ID
		}

		perms, err := data.GuildData.GS.GetMemberPermissions(cID, target.User.ID, target.Member.Roles)
		if err != nil {
			return "Unable to calculate perms", err
		}

		humanized := common.HumanizePermissions(int64(perms))
		return fmt.Sprintf("Perms of %s in <#%d>\n`%d`\n%s", target.User.Username, cID, perms, strings.Join(humanized, ", ")), nil
	},
}
