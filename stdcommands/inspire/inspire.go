package inspire

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"regexp"
	"time"

	"github.com/botlabs-gg/yagpdb/bot/paginatedmessages"
	"github.com/botlabs-gg/yagpdb/commands"
	"github.com/botlabs-gg/yagpdb/common"
	"github.com/botlabs-gg/yagpdb/lib/dcmd"
	"github.com/botlabs-gg/yagpdb/lib/discordgo"
	"github.com/botlabs-gg/yagpdb/stdcommands/util"
)

var Command = &commands.YAGCommand{
	CmdCategory:               commands.CategoryFun,
	Name:                      "Inspire",
	Aliases:                   []string{"insp"},
	Description:               "Shows 'inspirational' quotes from inspirobot.me...",
	RunInDM:                   false,
	DefaultEnabled:            true,
	ApplicationCommandEnabled: true,
	Cooldown:                  3,
	ArgSwitches: []*dcmd.ArgDef{
		{Name: "mindfulness", Help: "Generates Mindful Quotes!"},
	},

	RunFunc: func(data *dcmd.Data) (interface{}, error) {
		var pm *paginatedmessages.PaginatedMessage
		var paginatedView bool

		query := "https://inspirobot.me/api?generate=true"

		if data.Switches["mindfulness"].Value != nil && data.Switches["mindfulness"].Value.(bool) {
			paginatedView = true
		}

		if paginatedView {
			var ID = time.Now().UTC().Unix()

			query = fmt.Sprintf("https://inspirobot.me/api?generateFlow=1&sessionID=%d", ID)
			wonkyErr := "InspireAPI wonky... ducks are sad : /"

			requestBody, err := util.RequestFromAPI(query)
			if err != nil {
				return wonkyErr, err
			}

			mindfulness, err := handleRequestBody(requestBody)
			if err != nil {
				return wonkyErr, err
			}

			inspireArray := []string{}
			inspireArray = arrayMaker(inspireArray, mindfulness)

			pm, err = paginatedmessages.CreatePaginatedMessage(
				data.GuildData.GS.ID, data.ChannelID, 1, 25, func(p *paginatedmessages.PaginatedMessage, page int) (interface{}, error) {
					if page-1 == len(inspireArray) {
						requestBody, err := util.RequestFromAPI(query)
						if err != nil {
							return nil, err
						}

						mindfulness, err := handleRequestBody(requestBody)
						if err != nil {
							return nil, err
						}

						inspireArray = arrayMaker(inspireArray, mindfulness)
					}
					return createInspireEmbed(inspireArray[page-1], true), nil
				})
			if err != nil {
				return nil, err
			}

			return pm, nil
		}

		//Normal Image Inspire Output
		inspData, err := util.RequestFromAPI(query)
		if err != nil {
			return fmt.Sprintf("%s\nInspiroBot wonky... sad times :/", err), err
		}
		embed := createInspireEmbed(string(inspData), false)

		return embed, nil
	},
}

func handleRequestBody(rBody []byte) (string, error) {
	var mindful MindfulnessMode
	var mindfulness string

	err := json.Unmarshal(rBody, &mindful)
	if err != nil {
		return "", err
	}

	for _, i := range mindful.Data {
		if i.Text != "" {
			mindfulness = i.Text
		}
	}

	if len(mindfulness) > 4000 {
		mindfulness = common.CutStringShort(mindfulness, 4000)
	}

	return mindfulness, nil
}

func createInspireEmbed(data string, mindfulness bool) *discordgo.MessageEmbed {
	embed := &discordgo.MessageEmbed{}

	if mindfulness {
		embed.Title = "Here's an inspirational quote (Mindfulness Mode):"
		embed.Description = "```\n" + data + "```"
		embed.Color = int(11413503)
	} else {
		embed.Color = int(rand.Int63n(0xffffff))
		embed.Description = "Here's an inspirational quote:"
		embed.Image = &discordgo.MessageEmbedImage{
			URL: data,
		}
	}

	return embed
}
func arrayMaker(list []string, data string) []string {
	re := regexp.MustCompile(`\[pause \d+\]`)
	list = append(list, re.ReplaceAllString(data, ""))

	return list
}

type MindfulnessMode struct {
	Data []struct {
		Duration *float64 `json:"duration,omitempty"`
		Image    *string  `json:"image,omitempty"`
		Type     *string  `json:"type"`
		Time     *float64 `json:"time"`
		Text     string   `json:"text,omitempty"`
	} `json:"data"`
}
