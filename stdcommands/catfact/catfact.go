package catfact

import (
	"encoding/json"
	"io"
	"math/rand"
	"net/http"

	"github.com/botlabs-gg/yagpdb/commands"
	"github.com/botlabs-gg/yagpdb/common"
	"github.com/botlabs-gg/yagpdb/lib/dcmd"
	"github.com/botlabs-gg/yagpdb/lib/discordgo"
)

type CatStuff struct {
	Fact string `json:"fact"`
	File string `json:"file"`
}

var Command = &commands.YAGCommand{
	CmdCategory:               commands.CategoryFun,
	Name:                      "CatFact",
	Aliases:                   []string{"cf", "cat", "catfacts"},
	Description:               "Cat Facts from local database or catfact.ninja API",
	DefaultEnabled:            true,
	ApplicationCommandEnabled: true,
	ArgSwitches: []*dcmd.ArgDef{
		{Name: "raw", Help: "Raw output"},
	},

	RunFunc: func(data *dcmd.Data) (interface{}, error) {
		var catStuff *CatStuff
		var cf, cPicLink string
		var err error

		queryFact := "https://catfact.ninja/fact"
		queryPic := "https://aws.random.cat/meow"

		cf = Catfacts[rand.Intn(len(Catfacts))]
		cPicLink = "https://" + common.ConfHost.GetString() + "/static/img/quacknotfound.png"

		request := rand.Intn(2)
		if request > 0 {
			catStuff, err = catFactFromAPI(queryFact)
			cf = catStuff.Fact
			if err == nil {
				cf = catStuff.Fact
			}
		}

		if data.Switches["raw"].Value != nil && data.Switches["raw"].Value.(bool) {
			return cf, nil
		}

		catStuff, err = catFactFromAPI(queryPic)
		if err == nil {
			cPicLink = catStuff.File
		}

		embed := &discordgo.MessageEmbed{
			Description: cf,
			Color:       int(rand.Int63n(0xffffff)),
			Image: &discordgo.MessageEmbedImage{
				URL: cPicLink,
			},
		}

		return embed, nil
	},
}

func catFactFromAPI(query string) (*CatStuff, error) {
	var cf CatStuff

	req, err := http.NewRequest("GET", query, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", common.ConfBotUserAgent.GetString())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, commands.NewPublicError("Not 200!")
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	queryErr := json.Unmarshal([]byte(body), &cf)
	if queryErr != nil {
		return nil, err
	}

	return &cf, nil
}
