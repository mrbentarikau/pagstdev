package dogfact

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

type DogStuff struct {
	Facts   []string `json:"facts"`
	Message string   `json:"message"`
}

var Command = &commands.YAGCommand{
	CmdCategory:               commands.CategoryFun,
	Name:                      "DogFact",
	Aliases:                   []string{"dog", "dogfacts"},
	Description:               "Dog Facts from local database or dog-api.kinduff.com API",
	ApplicationCommandEnabled: true,
	DefaultEnabled:            true,
	ArgSwitches: []*dcmd.ArgDef{
		{Name: "raw", Help: "Raw output"},
	},

	RunFunc: func(data *dcmd.Data) (interface{}, error) {
		var dogStuff *DogStuff
		var df, dPicLink string
		var err error

		queryFact := "https://dog-api.kinduff.com/api/facts"
		queryPic := "https://dog.ceo/api/breeds/image/random"

		df = dogfacts[rand.Intn(len(dogfacts))]

		request := rand.Intn(2)
		if request > 0 {
			dogStuff, err = dogFactFromAPI(queryFact)
			if err == nil && len(dogStuff.Facts) > 0 {
				df = dogStuff.Facts[0]
			}
		}

		if data.Switches["raw"].Value != nil && data.Switches["raw"].Value.(bool) {
			return df, nil
		}

		dogStuff, err = dogFactFromAPI(queryPic)
		if err == nil {
			dPicLink = dogStuff.Message
		} else {
			return df, nil
		}

		embed := &discordgo.MessageEmbed{
			Description: df,
			Color:       int(rand.Int63n(0xffffff)),
			Image: &discordgo.MessageEmbedImage{
				URL: dPicLink,
			},
		}

		return embed, nil
	},
}

func dogFactFromAPI(query string) (*DogStuff, error) {
	var ds DogStuff

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

	queryErr := json.Unmarshal([]byte(body), &ds)
	if queryErr != nil {
		return nil, err
	}

	return &ds, nil
}
