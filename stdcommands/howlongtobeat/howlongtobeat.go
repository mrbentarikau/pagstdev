package howlongtobeat

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"time"

	"github.com/botlabs-gg/yagpdb/bot/paginatedmessages"
	"github.com/botlabs-gg/yagpdb/commands"
	"github.com/botlabs-gg/yagpdb/common"
	"github.com/botlabs-gg/yagpdb/lib/dcmd"
	"github.com/botlabs-gg/yagpdb/lib/discordgo"
	"github.com/botlabs-gg/yagpdb/lib/jarowinkler"
)

var (
	hltbScheme = "https"
	hltbHost   = "howlongtobeat.com"
	hltbURL    = fmt.Sprintf("%s://%s/", hltbScheme, hltbHost)
	//hltbHostPath = "search_results"
	hltbHostPath = "api/search"
	//hltbRawQuery = "page=1"
	hltbRawQuery = ""
)

// Command var needs a comment for lint :)
var Command = &commands.YAGCommand{
	CmdCategory:         commands.CategoryFun,
	Name:                "HowLongToBeat",
	Aliases:             []string{"hltb"},
	RequiredArgs:        1,
	Description:         "Game information based on query from howlongtobeat.com.\nResults are sorted by popularity, it's their default. Without -p returns the first result.\nSwitch -p gives paginated output using the Jaro-Winkler similarity metric sorting max 20 results.",
	DefaultEnabled:      true,
	SlashCommandEnabled: true,
	Arguments: []*dcmd.ArgDef{
		{Name: "Game-Title", Type: dcmd.String},
	},
	ArgSwitches: []*dcmd.ArgDef{
		{Name: "c", Help: "Compact output"},
		{Name: "p", Help: "Paginated output"},
	},
	RunFunc: func(data *dcmd.Data) (interface{}, error) {
		var compactView, paginatedView bool
		gameName := data.Args[0].Str()

		if data.Switches["c"].Value != nil && data.Switches["c"].Value.(bool) {
			compactView = true
		}

		if data.Switches["p"].Value != nil && data.Switches["p"].Value.(bool) {
			compactView = false
			paginatedView = true
		}

		getData, err := getGameData(gameName)
		if err != nil {
			return nil, err
		}

		var hltbQuery HowlongToBeat
		queryErr := json.Unmarshal(getData, &hltbQuery)
		if queryErr != nil {
			return nil, queryErr
		}

		for i, j := range hltbQuery.Data {
			hltbQuery.Data[i].CompMainHumanize = common.HumanizeDurationShort(common.DurationPrecisionMinutes, time.Second*time.Duration(j.CompMain))
			hltbQuery.Data[i].CompPlusHumanize = common.HumanizeDurationShort(common.DurationPrecisionMinutes, time.Second*time.Duration(j.CompPlus))
			hltbQuery.Data[i].Comp100Humanize = common.HumanizeDurationShort(common.DurationPrecisionMinutes, time.Second*time.Duration(j.Comp100))
			hltbQuery.Data[i].JaroWinklerSimilarity = jarowinkler.Similarity([]rune(gameName), []rune(j.GameName))
		}

		if compactView {
			compactData := fmt.Sprintf("%s: %s | %s | %s | <%s>",
				normaliseTitle(hltbQuery.Data[0].GameName),
				hltbQuery.Data[0].CompMainHumanize,
				hltbQuery.Data[0].CompPlusHumanize,
				hltbQuery.Data[0].Comp100Humanize,
				fmt.Sprintf("%sgame/%d", hltbURL, hltbQuery.Data[0].GameID),
			)
			return compactData, nil
		}

		hltbEmbed := embedCreator(hltbQuery, 0, paginatedView)

		var pm *paginatedmessages.PaginatedMessage
		if paginatedView {
			pm, err = paginatedmessages.CreatePaginatedMessage(
				data.GuildData.GS.ID, data.ChannelID, 1, len(hltbQuery.Data), func(p *paginatedmessages.PaginatedMessage, page int) (*discordgo.MessageEmbed, error) {
					i := page - 1
					sort.SliceStable(hltbQuery.Data, func(i, j int) bool {
						return hltbQuery.Data[i].JaroWinklerSimilarity > hltbQuery.Data[j].JaroWinklerSimilarity
					})
					paginatedEmbed := embedCreator(hltbQuery, i, paginatedView)
					return paginatedEmbed, nil
				})
			if err != nil {
				return "Something went wrong", nil
			}
		} else {
			return hltbEmbed, nil
		}

		return pm, nil
	},
}

func embedCreator(hltbQuery HowlongToBeat, i int, paginated bool) *discordgo.MessageEmbed {
	hltbURL := fmt.Sprintf("%s://%s/", hltbScheme, hltbHost)
	embed := &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			Name: normaliseTitle(hltbQuery.Data[i].GameName),
			URL:  fmt.Sprintf("%sgame/%d", hltbURL, hltbQuery.Data[i].GameID),
		},

		Color: int(rand.Int63n(16777215)),
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: fmt.Sprintf("%sgames/%s", hltbURL, hltbQuery.Data[i].GameImage),
		},
	}
	if hltbQuery.Data[i].CompMain > 0 {
		embed.Fields = append(embed.Fields,
			&discordgo.MessageEmbedField{Name: "Main Story", Value: hltbQuery.Data[i].CompMainHumanize})
	}
	if hltbQuery.Data[i].CompPlus > 0 {
		embed.Fields = append(embed.Fields,
			&discordgo.MessageEmbedField{Name: "Main + Extra", Value: hltbQuery.Data[i].CompPlusHumanize})
	}
	if hltbQuery.Data[i].Comp100 > 0 {
		embed.Fields = append(embed.Fields,
			&discordgo.MessageEmbedField{Name: "Completionist", Value: hltbQuery.Data[i].Comp100Humanize})
	}
	if hltbQuery.Data[i].ReviewScore > 0 {
		embed.Fields = append(embed.Fields,
			&discordgo.MessageEmbedField{Name: "Review Score", Value: fmt.Sprintf("%d%%", hltbQuery.Data[i].ReviewScore)})
	}
	if paginated {
		embed.Description = fmt.Sprintf("Term similarity: %.1f%%", hltbQuery.Data[i].JaroWinklerSimilarity*100)
	}
	return embed
}

func normaliseTitle(t string) string {
	return strings.Join(strings.Fields(t), " ")
}
