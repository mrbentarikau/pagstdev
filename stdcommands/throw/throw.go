package throw

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"

	"github.com/botlabs-gg/yagpdb/commands"
	"github.com/botlabs-gg/yagpdb/common"
	"github.com/botlabs-gg/yagpdb/lib/dcmd"
	"github.com/botlabs-gg/yagpdb/lib/discordgo"
)

var Command = &commands.YAGCommand{
	CmdCategory: commands.CategoryFun,
	Name:        "Throw",
	Description: "Throwing things is cool.",
	Arguments: []*dcmd.ArgDef{
		{Name: "Target", Type: dcmd.User},
	},
	DefaultEnabled: true,

	SlashCommandEnabled:    true,
	ApplicationCommandType: 2,

	RunFunc: func(data *dcmd.Data) (interface{}, error) {
		target := "a random person nearby"
		if data.Args[0].Value != nil {
			if data.GuildData.MS.User.ID == data.Args[0].Value.(*discordgo.User).ID {
				target = "themself"
			} else {
				target = data.Args[0].Value.(*discordgo.User).Username
			}
		}

		var resp string

		rng := rand.Intn(100)
		if rng < 5 {
			resp = fmt.Sprintf("TRIPLE THROW! Threw **%s**, **%s** and **%s** at **%s**", randomThing(), randomThing(), randomThing(), target)
		} else if rng < 15 {
			resp = fmt.Sprintf("DOUBLE THROW! Threw **%s** and **%s** at **%s**", randomThing(), randomThing(), target)
		} else {
			resp = fmt.Sprintf("Threw **%s** at **%s**", randomThing(), target)
		}

		return resp, nil
	},
}

func randomThing() string {
	randNum := rand.Intn(3)

	if randNum > 1 {
		return common.RandomNoun()
	} else if randNum > 0 {
		if qrt, err := queryRandomThing(); err == nil {
			return qrt
		}
	}

	return throwThings[rand.Intn(len(throwThings))]
}

func queryRandomThing() (string, error) {
	var qrtQuery = "https://roger.redevised.com/api/v1/"

	req, err := http.NewRequest("GET", qrtQuery, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("User-Agent", common.ConfBotUserAgent.GetString())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != 200 {
		return "", commands.NewPublicError("HTTP Response was not 200, but: ", resp.StatusCode)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}
