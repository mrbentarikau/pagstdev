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
	},
	ApplicationCommandEnabled: false,
	DefaultEnabled:            false,
	Cooldown:                  5,
	RunFunc: func(data *dcmd.Data) (interface{}, error) {
		var extraHeaders = map[string]string{"X-MAL-CLIENT-ID": confMalAPIKey.GetString()}
		var queryManga bool
		var queryLimit = 25
		var embedPages = 1
		var paginatedView bool
		var nsfw bool

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
			embedPages = 5
		}

		// Initial search for titles
		querySearch := fmt.Sprintf("%s/anime?q=%s&limit=%d&nsfw=%t", malRequestURI, url.QueryEscape(queryTitle), queryLimit, nsfw)
		if queryManga {
			querySearch = fmt.Sprintf("%s/manga?q=%s&limit=%d&nsfw=%t", malRequestURI, url.QueryEscape(queryTitle), queryLimit, nsfw)
		}

		body, err := util.RequestFromAPI(querySearch, extraHeaders)
		if err != nil {
			return "No such title found", err
		}

		err = json.Unmarshal([]byte(body), &malSearch)
		if err != nil {
			return nil, err
		}

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

		// Query for details of the title
		var animeData []AnimeMangaDetails
		var animeDetails AnimeMangaDetails

		for i := 0; i < embedPages; i++ {
			queryDetails := fmt.Sprintf("%s/anime/%d?fields=id,title,main_picture,alternative_titles,start_date,end_date,synopsis,mean,rank,popularity,num_list_users,num_scoring_users,nsfw,created_at,updated_at,media_type,status,genres,num_episodes,start_season,broadcast,source,average_episode_duration,rating,pictures,related_anime,related_manga,studios&nsfw=%t", malRequestURI, malSearch.Data[i].Node.ID, nsfw)

			if queryManga {
				queryDetails = fmt.Sprintf("%s/manga/%d?fields=id,title,main_picture,alternative_titles,start_date,end_date,synopsis,mean,rank,popularity,num_list_users,num_scoring_users,nsfw,created_at,updated_at,media_type,status,genres,num_volumes,num_chapters,authors{first_name,last_name},pictures,background,related_anime,related_manga,serialization{name}&nsfw=%t", malRequestURI, malSearch.Data[i].Node.ID, nsfw)
			}

			body, err = util.RequestFromAPI(queryDetails, extraHeaders)
			if err != nil {
				return nil, err
			}

			err = json.Unmarshal([]byte(body), &animeDetails)
			if err != nil {
				return nil, err
			}

			animeData = append(animeData, animeDetails)
		}

		malData.AMData = animeData

		malEmbed := embedCreator(malData, 0, paginatedView, queryManga)

		var pm *paginatedmessages.PaginatedMessage
		if paginatedView {
			pm, err = paginatedmessages.CreatePaginatedMessage(
				data.GuildData.GS.ID, data.ChannelID, 1, len(malData.AMData), func(p *paginatedmessages.PaginatedMessage, page int) (*discordgo.MessageEmbed, error) {
					i := page - 1
					paginatedEmbed := embedCreator(malData, i, paginatedView, queryManga)
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

func embedCreator(malData MyAnimeListData, i int, paginated, isManga bool) *discordgo.MessageEmbed {
	var authorsBuilder, genresBuilder, relatedBuilder, serializationBuilder, studiosBuilder, synonymsBuilder strings.Builder
	for _, v := range malData.AMData[i].Genres {
		if genresBuilder.Len() != 0 {
			genresBuilder.WriteString(", ")
		}

		genresBuilder.WriteString("`" + v.Name + "`")
	}

	for _, v := range malData.AMData[i].RelatedAnime {
		if relatedBuilder.Len() != 0 {
			relatedBuilder.WriteString(", ")
		}

		relatedBuilder.WriteString(v.Node.Title)
	}

	for _, v := range malData.AMData[i].RelatedManga {
		if relatedBuilder.Len() != 0 {
			relatedBuilder.WriteString(", ")
		}

		relatedBuilder.WriteString(v.Node.Title)
	}

	for _, v := range malData.AMData[i].Authors {
		if authorsBuilder.Len() != 0 {
			authorsBuilder.WriteString("; ")
		}

		authorsBuilder.WriteString(v.Node.LastName + ", " + v.Node.FirstName + " (" + v.Role + ")")
	}

	for _, v := range malData.AMData[i].Serialization {
		if serializationBuilder.Len() != 0 {
			serializationBuilder.WriteString(", ")
		}

		serializationBuilder.WriteString(v.Node.Name)
	}

	for _, v := range malData.AMData[i].Studios {
		if studiosBuilder.Len() != 0 {
			studiosBuilder.WriteString(", ")
		}

		studiosBuilder.WriteString(v.Name)
	}

	for _, v := range malData.AMData[i].AlternativeTitles.Synonyms {
		if synonymsBuilder.Len() != 0 {
			synonymsBuilder.WriteString(", ")
		}

		synonymsBuilder.WriteString("`" + v + "`")
	}

	authorURL := fmt.Sprintf("https://myanimelist.net/anime/%d", malData.AMData[i].ID)
	if isManga {
		authorURL = fmt.Sprintf("https://myanimelist.net/manga/%d", malData.AMData[i].ID)
	}

	embed := &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			Name:    "MyAnimeList",
			URL:     authorURL,
			IconURL: "https://upload.wikimedia.org/wikipedia/commons/7/7a/MyAnimeList_Logo.png",
		},

		Title:       malData.AMData[i].Title,
		Description: common.CutStringShort(malData.AMData[i].Synopsis, 512),

		Thumbnail: &discordgo.MessageEmbedThumbnail{URL: malData.AMData[i].MainPicture.Medium},
		//Image:     &discordgo.MessageEmbedImage{URL: malData.AnimeData[i].Pictures[0].Medium},

		Color: int(rand.Int63n(16777215)),

		Fields: []*discordgo.MessageEmbedField{},
	}

	if malData.AMData[i].MediaType != "" {
		mediaType := malData.AMData[i].MediaType
		if len(mediaType) <= 3 {
			mediaType = strings.ToUpper(mediaType)
		}
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{Name: "Type", Value: cases.Title(language.Und).String(mediaType), Inline: true})
	}

	if malData.AMData[i].NumEpisodes > 0 {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{Name: "Episodes", Value: fmt.Sprintf("%d", malData.AMData[i].NumEpisodes), Inline: true})
	}

	if malData.AMData[i].NumVolumes > 0 {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{Name: "Chapters", Value: fmt.Sprintf("%d", malData.AMData[i].NumVolumes), Inline: true})
	}

	if malData.AMData[i].NumChapters > 0 {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{Name: "Volumes", Value: fmt.Sprintf("%d", malData.AMData[i].NumChapters), Inline: true})
	}

	if malData.AMData[i].Rating != "" {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{Name: "Rating", Value: cases.Title(language.Und).String(malData.AMData[i].Rating), Inline: true})
	}

	if malData.AMData[i].StartDate != "" || malData.AMData[i].EndDate != "" {
		endDate := malData.AMData[i].EndDate
		if endDate == "" {
			endDate = "?"
		}

		if isManga {
			embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{Name: "Published", Value: fmt.Sprintf("%s to %s", malData.AMData[i].StartDate, endDate)})
		} else {
			embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{Name: "Aired", Value: fmt.Sprintf("%s to %s", malData.AMData[i].StartDate, endDate)})
		}
	}

	if malData.AMData[i].AverageEpisodeDuration > 0 {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{Name: "Duration", Value: fmt.Sprintf("%s per ep.", common.HumanizeDurationShort(common.DurationPrecisionMinutes, time.Second*time.Duration(malData.AMData[i].AverageEpisodeDuration))), Inline: true})
	}

	if malData.AMData[i].Source != "" {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{Name: "Source", Value: cases.Title(language.Und).String(malData.AMData[i].Source), Inline: true})
	}

	if studiosBuilder.String() != "" {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{Name: "Studio", Value: studiosBuilder.String(), Inline: true})
	}

	if authorsBuilder.String() != "" {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{Name: "Authors", Value: authorsBuilder.String(), Inline: true})
	}

	if malData.AMData[i].AlternativeTitles.En != "" || malData.AMData[i].AlternativeTitles.Ja != "" || synonymsBuilder.String() != "" {
		var synonyms string
		if synonymsBuilder.String() != "" {
			synonyms = "Synonyms: " + synonymsBuilder.String()
		}
		enAltTitle := malData.AMData[i].AlternativeTitles.En
		jaAltTitle := malData.AMData[i].AlternativeTitles.Ja
		if enAltTitle != "" {
			enAltTitle = "`" + enAltTitle + "`,"
		}

		if jaAltTitle != "" {
			jaAltTitle = "`" + jaAltTitle + "`"
		}

		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{Name: "Alternative Titles", Value: fmt.Sprintf("%s %s\n%s", enAltTitle, jaAltTitle, synonyms)})
	}

	if genresBuilder.String() != "" {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{Name: "Genres", Value: genresBuilder.String()})
	}

	if relatedBuilder.String() != "" {
		if isManga {
			embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{Name: "Related Manga", Value: common.CutStringShort(relatedBuilder.String(), 280)})
		} else {
			embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{Name: "Related Anime", Value: common.CutStringShort(relatedBuilder.String(), 280)})
		}
	}

	if serializationBuilder.String() != "" {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{Name: "Serialization", Value: common.CutStringShort(serializationBuilder.String(), 280)})
	}

	if malData.AMData[i].Mean > 0 {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{Name: "MAL Score", Value: fmt.Sprintf("**%.2f** (by %s users)", malData.AMData[i].Mean, util.HumanizeThousands(malData.AMData[i].NumScoringUsers)), Inline: true})
	}

	if malData.AMData[i].Rank > 0 || malData.AMData[i].Popularity > 0 {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{Name: "MAL Rank & Popularity", Value: fmt.Sprintf("#%d and #%d", malData.AMData[i].Rank, malData.AMData[i].Popularity), Inline: true})
	}

	if !paginated {
		embed.Footer = &discordgo.MessageEmbedFooter{Text: "myanimelist.net"}
		embed.Timestamp = time.Now().Format(time.RFC3339)
	}

	return embed
}
