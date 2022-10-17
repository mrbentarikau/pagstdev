package myanimelist

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/botlabs-gg/yagpdb/bot/paginatedmessages"
	"github.com/botlabs-gg/yagpdb/commands"
	"github.com/botlabs-gg/yagpdb/common"
	"github.com/botlabs-gg/yagpdb/common/config"
	"github.com/botlabs-gg/yagpdb/lib/dcmd"
	"github.com/botlabs-gg/yagpdb/lib/discordgo"
	"github.com/botlabs-gg/yagpdb/lib/jarowinkler"
	"github.com/botlabs-gg/yagpdb/stdcommands/util"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var malRequestURI = "https://api.myanimelist.net/v2"
var confMalAPIKey = config.RegisterOption("yagpdb.mal_api_key", "MyAnimeList API key", "")

func ShouldRegister() bool {
	return confMalAPIKey.GetString() != ""
}

var Command = &commands.YAGCommand{
	CmdCategory: commands.CategoryFun,
	Name:        "MyAnimeList",
	Aliases:     []string{"mal"},
	Description: "Queries [MyAnimeList](https://myanimelist.net) for anime and manga.",
	Arguments: []*dcmd.ArgDef{
		{Name: "Title", Type: dcmd.String},
	},
	ArgSwitches: []*dcmd.ArgDef{
		{Name: "manga", Help: "Search for mangas"},
		{Name: "nsfw", Help: "NSFW enabled (MAL does not show some things)"},
		{Name: "p", Help: "Paginated output"},
		{Name: "top100", Help: "Top 100 titles by ranking"},
	},
	ApplicationCommandEnabled: false,
	DefaultEnabled:            false,
	Cooldown:                  5,
	RunFunc: func(data *dcmd.Data) (interface{}, error) {
		var extraHeaders = map[string]string{"X-MAL-CLIENT-ID": confMalAPIKey.GetString()}
		var queryLimit = 25
		var nsfw, paginatedView, queryManga, topHundred bool

		var malSearch AnimeMangaList
		var malData MyAnimeListData

		queryTitle := data.Args[0].Str()

		if data.Switches["manga"].Value != nil && data.Switches["manga"].Value.(bool) {
			queryManga = true
		}

		if data.Switches["nsfw"].Value != nil && data.Switches["nsfw"].Value.(bool) {
			nsfw = true
		}

		if data.Switches["p"].Value != nil && data.Switches["p"].Value.(bool) {
			paginatedView = true
		}

		if data.Switches["top100"].Value != nil && data.Switches["top100"].Value.(bool) {
			topHundred = true
			queryLimit = 100
			paginatedView = true
		}

		queryFields := "id,title,main_picture,alternative_titles,start_date,end_date,synopsis,mean,rank,popularity,num_list_users,num_scoring_users,nsfw,created_at,updated_at,media_type,status,genres,num_volumes,num_chapters,authors{first_name,last_name},num_episodes,start_season,broadcast,source,average_episode_duration,rating,pictures,background,related_anime,related_manga,studios,serialization{name}"

		// Initial search for titles
		var querySearch string
		if !topHundred {
			querySearch = fmt.Sprintf("%s/anime?q=%s&fields=%s&limit=%d&nsfw=%t", malRequestURI, url.QueryEscape(queryTitle), queryFields, queryLimit, nsfw)
			if queryManga {
				querySearch = fmt.Sprintf("%s/manga?q=%s&fields=%s&limit=%d&nsfw=%t", malRequestURI, url.QueryEscape(queryTitle), queryFields, queryLimit, nsfw)
			}
		} else {
			querySearch = fmt.Sprintf("%s/anime/ranking?ranking_type=all&fields=%s&limit=%d'", malRequestURI, queryFields, queryLimit)
			if queryManga {
				querySearch = fmt.Sprintf("%s/manga/ranking?ranking_type=all&fields=%s&limit=%d'", malRequestURI, queryFields, queryLimit)
			}
		}

		if !topHundred && queryTitle == "" {
			return "No title provided...", nil
		}

		body, err := util.RequestFromAPI(querySearch, extraHeaders)
		if err != nil {
			return "No such title found...", err
		}

		err = json.Unmarshal([]byte(body), &malSearch)
		if err != nil {
			return nil, err
		}

		if !topHundred {
			var animeSearchDataNode []AnimeMangaListNode
			for _, node := range malSearch.Data {
				node.Node.JaroWinklerSimilarity = jarowinkler.Similarity([]rune(queryTitle), []rune(node.Node.Title))
				animeSearchDataNode = append(animeSearchDataNode, node)
			}

			// sort the search results more closer to title's similarity
			malSearch.Data = animeSearchDataNode
			sort.SliceStable(malSearch.Data, func(i, j int) bool {
				return malSearch.Data[i].Node.JaroWinklerSimilarity > malSearch.Data[j].Node.JaroWinklerSimilarity
			})
		}

		// Query for details of the title
		var animeData []AnimeMangaDetails
		for _, j := range malSearch.Data {
			animeData = append(animeData, j.Node)
		}
		malData.AMData = animeData

		malEmbed := embedCreator(malData, 0, paginatedView, queryManga, topHundred)

		var pm *paginatedmessages.PaginatedMessage
		if paginatedView {
			pm, err = paginatedmessages.CreatePaginatedMessage(
				data.GuildData.GS.ID, data.ChannelID, 1, len(malData.AMData), func(p *paginatedmessages.PaginatedMessage, page int) (*discordgo.MessageEmbed, error) {
					i := page - 1
					paginatedEmbed := embedCreator(malData, i, paginatedView, queryManga, topHundred)
					return paginatedEmbed, nil
				})
			if err != nil {
				return "Something went wrong making the paginated messages", nil
			}
		} else {
			return malEmbed, nil
		}

		return pm, nil
	},
}

type malBuilder struct {
	strings.Builder
}

func (mb *malBuilder) Sep(sep string) {
	if mb.Len() != 0 {
		mb.WriteString(sep)
	}
}

func amIDNameToBuilder(anIDN []AnimatedMangaIDName) malBuilder {
	var builder malBuilder
	for _, v := range anIDN {
		builder.Sep(", ")
		builder.WriteString("`" + v.Name + "`")
	}
	return builder
}

func embedCreator(malData MyAnimeListData, i int, paginated, isManga, isTopHundred bool) *discordgo.MessageEmbed {
	var authorsBuilder, genresBuilder, relatedBuilder, serializationBuilder, studiosBuilder, synonymsBuilder malBuilder
	var mData = malData.AMData[i]

	genresBuilder = amIDNameToBuilder(mData.Genres)
	studiosBuilder = amIDNameToBuilder(mData.Studios)

	for _, v := range mData.Related() {
		relatedBuilder.Sep(", ")
		relatedBuilder.WriteString(v.Node.Title)
	}

	for _, v := range mData.Authors {
		authorsBuilder.Sep("; ")
		if v.Node.FirstName != "" {
			v.Node.FirstName = ", " + v.Node.FirstName
		}
		if v.Role != "" {
			v.Role = " (" + v.Role + ")"
		}
		authorsBuilder.WriteString(v.Node.LastName + v.Node.FirstName + v.Role)
	}

	for _, v := range mData.Serialization {
		serializationBuilder.Sep(", ")
		serializationBuilder.WriteString(v.Node.Name)
	}

	for _, v := range mData.AlternativeTitles.Synonyms {
		synonymsBuilder.Sep(", ")
		synonymsBuilder.WriteString("`" + v + "`")
	}

	mediaType := mData.MediaType
	if len(mediaType) <= 3 {
		mediaType = strings.ToUpper(mediaType)
	} else {
		mediaType = cases.Title(language.Und).String(mediaType)
	}

	var dateField, relatedField *discordgo.MessageEmbedField
	startDate := mData.StartDate
	endDate := mData.EndDate
	if startDate != "" {
		var timeLayout = "2006-01-02"
		var formatLayout = "Jan 02, 2006"

		startDateTime, err := time.Parse(timeLayout, startDate)
		if err != nil {
			startDate = "?"
		} else {
			startDate = startDateTime.Format(formatLayout)
		}

		if endDate == "" {
			endDate = "?"
		} else {
			endDateTime, err := time.Parse(timeLayout, endDate)
			if err != nil {
				endDate = "?"
			} else {
				endDate = endDateTime.Format(formatLayout)
			}
		}

		if isManga {
			dateField = &discordgo.MessageEmbedField{Name: "Published", Value: fmt.Sprintf("%s to %s", startDate, endDate)}
			relatedField = &discordgo.MessageEmbedField{Name: "Related Manga", Value: common.CutStringShort(relatedBuilder.String(), 280)}
		} else {
			dateField = &discordgo.MessageEmbedField{Name: "Aired", Value: fmt.Sprintf("%s to %s", startDate, endDate)}
			relatedField = &discordgo.MessageEmbedField{Name: "Related Anime", Value: common.CutStringShort(relatedBuilder.String(), 280)}
		}
	}

	var synonyms string
	if synonymsBuilder.String() != "" {
		synonyms = "Synonyms: " + synonymsBuilder.String()
	}
	enAltTitle := mData.AlternativeTitles.En
	jaAltTitle := mData.AlternativeTitles.Ja
	if enAltTitle != "" {
		enAltTitle = "`" + enAltTitle + "`,"
	}

	if jaAltTitle != "" {
		jaAltTitle = "`" + jaAltTitle + "`"
	}

	var authorURL string
	if !isTopHundred {
		authorURL = fmt.Sprintf("https://myanimelist.net/anime/%d", mData.ID)
		if isManga {
			authorURL = fmt.Sprintf("https://myanimelist.net/manga/%d", mData.ID)
		}
	} else {
		authorURL = "https://myanimelist.net/topanime.php"
		if isManga {
			authorURL = "https://myanimelist.net/topmanga.php"
		}
	}

	var titleURL string
	if isTopHundred {
		titleURL = fmt.Sprintf("https://myanimelist.net/anime/%d", mData.ID)
		if isManga {
			titleURL = fmt.Sprintf("https://myanimelist.net/manga/%d", mData.ID)
		}
	}

	embed := &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			Name:    "MyAnimeList",
			URL:     authorURL,
			IconURL: "https://upload.wikimedia.org/wikipedia/commons/7/7a/MyAnimeList_Logo.png",
		},

		Title:       mData.Title,
		URL:         titleURL,
		Description: common.CutStringShort(mData.Synopsis, 512),
		Thumbnail:   &discordgo.MessageEmbedThumbnail{URL: malData.AMData[i].MainPicture.Medium},
		Color:       int(rand.Int63n(16777215)),

		Fields: []*discordgo.MessageEmbedField{
			{Name: "Type", Value: mediaType, Inline: true},
			{Name: "Episodes", Value: fmt.Sprintf("%d", mData.NumEpisodes), Inline: true},
			{Name: "Chapters", Value: fmt.Sprintf("%d", mData.NumVolumes), Inline: true},
			{Name: "Volumes", Value: fmt.Sprintf("%d", mData.NumChapters), Inline: true},
			{Name: "Rating", Value: cases.Title(language.Und).String(mData.Rating), Inline: true},
			dateField,
			{Name: "Duration", Value: fmt.Sprintf("%s per ep.", common.HumanizeDurationShort(common.DurationPrecisionMinutes, time.Second*time.Duration(mData.AverageEpisodeDuration))), Inline: true},
			{Name: "Source", Value: cases.Title(language.Und).String(mData.Source), Inline: true},
			{Name: "Studio", Value: studiosBuilder.String(), Inline: true},
			{Name: "Authors", Value: authorsBuilder.String(), Inline: true},
			{Name: "Alternative Titles", Value: fmt.Sprintf("%s %s\n%s", enAltTitle, jaAltTitle, synonyms)},
			{Name: "Genres", Value: genresBuilder.String()},
			relatedField,
			{Name: "Serialization", Value: common.CutStringShort(serializationBuilder.String(), 280)},
			{Name: "MAL Score", Value: fmt.Sprintf("**%.2f** (by %s users)", mData.Mean, util.HumanizeThousands(mData.NumScoringUsers)), Inline: true},
			{Name: "MAL Rank & Popularity", Value: fmt.Sprintf("#%d and #%d", mData.Rank, mData.Popularity), Inline: true},
		},
	}

	var cleanFields []*discordgo.MessageEmbedField
	for _, v := range embed.Fields {
		if v.Value == "" || v.Value == "0" || (isManga && v.Name == "Duration") {
			continue
		}
		cleanFields = append(cleanFields, v)
	}
	embed.Fields = cleanFields

	if !paginated {
		embed.Footer = &discordgo.MessageEmbedFooter{Text: "myanimelist.net"}
		embed.Timestamp = time.Now().Format(time.RFC3339)
	}

	return embed
}
