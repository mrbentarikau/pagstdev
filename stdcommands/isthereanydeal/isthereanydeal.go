package isthereanydeal

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/url"
	"strings"

	"github.com/botlabs-gg/yagpdb/bot/paginatedmessages"
	"github.com/botlabs-gg/yagpdb/commands"
	"github.com/botlabs-gg/yagpdb/common/config"
	"github.com/botlabs-gg/yagpdb/lib/dcmd"
	"github.com/botlabs-gg/yagpdb/lib/discordgo"
	"github.com/botlabs-gg/yagpdb/stdcommands/util"
)

var confItadAPIKey = config.RegisterOption("yagpdb.itad_api_key", "IsThereAnyDeal API key", "")

func ShouldRegister() bool {
	return confItadAPIKey.GetString() != ""
}

var Currency = map[string]string{
	"EUR": "\u20ac",
	"GBP": "\u00a3",
	"USD": "$",
	"CAD": "C$",
	"BRL": "R$",
	"AUD": "A$",
	"TRY": "\u20BA",
	"CNY": "\u00a5",
}

var Command = &commands.YAGCommand{
	CmdCategory:  commands.CategoryFun,
	Name:         "IsThereAnyDeal",
	Aliases:      []string{"itad", "anydeal", "gamedeal"},
	Description:  "Queries [IsThereAnyDeal](https://isthereanydeal.com) API for current prices of targeted game.",
	RequiredArgs: 1,
	Cooldown:     5,
	Arguments: []*dcmd.ArgDef{
		{Name: "Title", Help: "Search prices based on game's title", Type: dcmd.String},
	},
	ArgSwitches: []*dcmd.ArgDef{
		{Name: "c", Help: "Compact output"},
		{Name: "p", Help: "Paginated output"},
		{Name: "country", Help: "Country code for price currency if available, default is US dollar.", Type: dcmd.String},
	},
	DefaultEnabled:            true,
	ApplicationCommandEnabled: true,
	RunFunc: func(data *dcmd.Data) (interface{}, error) {
		var compactView, paginatedView bool
		var itadAPIKey = confItadAPIKey.GetString()
		var itadAPIHost = "https://api.isthereanydeal.com"
		var queryLimit = 25
		var country = "US"

		queryTitle := strings.ToLower(data.Args[0].Str())

		switchCountry := data.Switch("country")
		if switchCountry.Value != nil {
			country = strings.ToUpper(switchCountry.Str())
		}

		if data.Switches["c"].Value != nil && data.Switches["c"].Value.(bool) {
			compactView = true
		}

		if data.Switches["p"].Value != nil && data.Switches["p"].Value.(bool) {
			compactView = false
			paginatedView = true
		}

		querySearch := fmt.Sprintf("%s/v02/search/search/?key=%s&q=%s&limit=%d", itadAPIHost, itadAPIKey, url.QueryEscape(queryTitle), queryLimit)

		body, err := util.RequestFromAPI(querySearch)
		if err != nil {
			return "nil", err
		}

		var itadSearch ItadSearch
		err = json.Unmarshal([]byte(body), &itadSearch)
		if err != nil {
			return nil, err
		}

		if len(itadSearch.Data.Results) == 0 {
			return "Game title not found", nil
		}

		itadSearchResults := itadSearch.Data.Results

		var plainsBuilder strings.Builder
		for _, v := range itadSearchResults {
			if plainsBuilder.Len() != 0 {
				plainsBuilder.WriteString(",")
			}

			plainsBuilder.WriteString(v.Plain)
		}

		queryPrice := fmt.Sprintf("%s/v01/game/prices/?key=%s&plains=%s&country=%s", itadAPIHost, itadAPIKey, plainsBuilder.String(), country)
		body, err = util.RequestFromAPI(queryPrice)
		if err != nil {
			return nil, err
		}

		var itadPrice ItadPrice
		err = json.Unmarshal([]byte(body), &itadPrice)
		if err != nil {
			return nil, err
		}

		if compactView {
			var compactPriceNew string
			compactPrice := itadPrice.Data[itadSearchResults[0].Plain].List[0]
			if itadPrice.Meta.Currency == "EUR" {
				compactPriceNew = fmt.Sprintf("%.2f%s", compactPrice.PriceNew, Currency[itadPrice.Meta.Currency])
			} else {
				compactPriceNew = fmt.Sprintf("%s%.2f", Currency[itadPrice.Meta.Currency], compactPrice.PriceNew)

			}
			compactData := fmt.Sprintf("%s: %s | CUT: %d%% | %s: <%s> | ITAD: <%s>",
				itadSearchResults[0].Title,
				compactPriceNew,
				compactPrice.PriceCut,
				compactPrice.Shop.Name,
				compactPrice.URL,
				itadPrice.Data[itadSearchResults[0].Plain].Urls.Game,
			)
			return compactData, nil
		}

		queryInfo := fmt.Sprintf("%s/v01/game/info/?key=%s&plains=%s", itadAPIHost, itadAPIKey, plainsBuilder.String())
		body, err = util.RequestFromAPI(queryInfo)
		if err != nil {
			return nil, err
		}

		var itadGameInfo ItadGameInfo
		err = json.Unmarshal([]byte(body), &itadGameInfo)
		if err != nil {
			return nil, err
		}

		itadComplete := &ItadComplete{
			Search:   itadSearch,
			Price:    itadPrice,
			GameInfo: itadGameInfo,
		}

		itadEmbed := embedCreator(itadComplete, 0, paginatedView)

		var pm *paginatedmessages.PaginatedMessage
		if paginatedView {
			pm, err = paginatedmessages.CreatePaginatedMessage(
				data.GuildData.GS.ID, data.ChannelID, 1, len(itadComplete.Search.Data.Results), func(p *paginatedmessages.PaginatedMessage, page int) (*discordgo.MessageEmbed, error) {
					i := page - 1
					paginatedEmbed := embedCreator(itadComplete, i, paginatedView)
					return paginatedEmbed, nil
				})
			if err != nil {
				return "Something went wrong", nil
			}
		} else {
			return itadEmbed, nil
		}

		return pm, nil
	},
}

func embedCreator(itadComplete *ItadComplete, i int, paginated bool) *discordgo.MessageEmbed {
	var embedDescription, embedPriceNew, embedPriceOld string
	embedCurrency := Currency[itadComplete.Price.Meta.Currency]
	embedPricingHeader := "`Price Cut | Current | Regular |` Store\n"
	embedDescription = embedPricingHeader

	itadPriceDataListSlice := itadComplete.Price.Data[itadComplete.Search.Data.Results[i].Plain].List

	if len(itadPriceDataListSlice) == 0 {
		embedDescription = "```No tracked stores are currently selling this game in this region/country...```"
	}

	for _, v := range itadPriceDataListSlice {
		if itadComplete.Price.Meta.Currency == "EUR" {
			embedPriceNew = fmt.Sprintf("%.2f%s", v.PriceNew, embedCurrency)
			embedPriceOld = fmt.Sprintf("%.2f%s", v.PriceOld, embedCurrency)
		} else {
			embedPriceNew = fmt.Sprintf("%s%.2f", embedCurrency, v.PriceNew)
			embedPriceOld = fmt.Sprintf("%s%.2f", embedCurrency, v.PriceOld)
		}

		embedDescription += fmt.Sprintf("`%9s | %7s | %7s |` [%s](%s)\n",
			fmt.Sprintf("%d%%", v.PriceCut),
			embedPriceNew,
			embedPriceOld,
			v.Shop.Name,
			v.URL,
		)
	}

	embed := &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			Name:    "IsThereAnyDeal",
			URL:     itadComplete.Price.Data[itadComplete.Search.Data.Results[i].Plain].Urls.Game,
			IconURL: "https://raw.githubusercontent.com/IsThereAnyDeal/AugmentedSteam/master/src/img/itad.png",
		},

		Title:       itadComplete.Search.Data.Results[i].Title,
		Description: embedDescription,

		Color: int(rand.Int63n(16777215)),
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: itadComplete.GameInfo.Data[itadComplete.Search.Data.Results[i].Plain].Image,
		},
	}

	return embed
}
