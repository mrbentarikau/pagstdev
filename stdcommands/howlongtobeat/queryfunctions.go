package howlongtobeat

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/botlabs-gg/yagpdb/commands"
	"github.com/botlabs-gg/yagpdb/common"
	"github.com/botlabs-gg/yagpdb/lib/jarowinkler"
)

func getGameData(searchTitle string) (string, error) {
	// data := url.Values{}
	// data.Set("searchTerms", searchTitle) //setting default request header query form data, the site uses
	// data.Add("searchType", "games")      // search type - for games, second option would be HLTB users
	// data.Add("searchPage", "1")
	//data.Add("sorthead", "popular")      // sort by release date,rating,popularity or name...  all parameters can be seen via header data, popular's the best
	//data.Add("sortd", "Normal Order")    // sorting, Normal or Reverse
	//data.Add("plat", "")                 // platform, empty string is for all
	//data.Add("length_type", "main")      // length range category, main is fine
	//data.Add("length_min", "")           // game length min
	//data.Add("length_max", "")           // game length max
	//data.Add("detail", "")               // extra information with user_stats ala speedruns, user rating etc...

	u := &url.URL{
		Scheme:   hltbScheme,
		Host:     hltbHost,
		Path:     hltbHostPath,
		RawQuery: hltbRawQuery,
	}

	urlStr := u.String()

	client := &http.Client{}

	var jsonQuery HowlongToBeatQuery
	jsonQuery.SearchTerms = []string{searchTitle}
	jsonQuery.SearchType = "games"
	jsonQuery.SearchPage = 1
	jsonQuery.Size = 25
	jsonQuery.SearchOptions.Games.UserID = 0
	jsonQuery.SearchOptions.Games.Platform = ""
	jsonQuery.SearchOptions.Games.SortCategory = "popular"
	jsonQuery.SearchOptions.Games.RangeCategory = "main"
	jsonQuery.SearchOptions.Games.RangeTime.Max = 0
	jsonQuery.SearchOptions.Games.RangeTime.Min = 0
	jsonQuery.SearchOptions.Games.Modifier = ""
	jsonQuery.SearchOptions.Users.SortCategory = "postcount"
	jsonQuery.SearchOptions.Filter = ""
	jsonQuery.SearchOptions.Sort = 0
	jsonQuery.SearchOptions.Randomizer = 0

	jsonData, err := json.Marshal(jsonQuery)
	if err != nil {
		return "", err
	}

	r, _ := http.NewRequest("POST", urlStr, strings.NewReader(string(jsonData))) // URL-encoded payload
	r.Header.Add("Content-Type", "application/json")
	r.Header.Add("Accept", "*/*")
	r.Header.Add("Content-Length", strconv.Itoa(len(jsonData)))
	r.Header.Add("User-Agent", common.ConfBotUserAgent.GetString())
	r.Header.Add("Authority", hltbURL)
	r.Header.Add("Origin", hltbURL)
	r.Header.Add("Referer", hltbURL)

	resp, err := client.Do(r)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != 200 {
		return "", commands.NewPublicError("Unable to fetch data from howlongtobeat.com", resp.StatusCode)
	}
	r.Body.Close()
	defer resp.Body.Close()

	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	stringBody := string(bytes)
	return stringBody, nil
}

func parseGameData(gameName string, toReader *strings.Reader) ([]hltb, error) {
	var hltbQuery []hltb
	var queryParsed hltb

	parseData, err := goquery.NewDocumentFromReader(toReader)
	if err != nil {
		return nil, err
	}

	parseData.Find("li").Each(func(_ int, sel *goquery.Selection) {
		queryParsed.ImageURL = sel.Find("img").AttrOr("src", "")
		queryParsed.GameURL = hltbURL + sel.Find("a").AttrOr("href", "")

		queryParsed.GameTitle = strings.TrimSpace(sel.Find("h3").Text())
		queryParsed.PureTitle = strings.TrimSpace(sel.Find("a").AttrOr("title", "")) //a tag has game title without &() etc
		queryParsed.JaroWinklerSimilarity = jarowinkler.Similarity([]rune(gameName), []rune(queryParsed.PureTitle))

		/*if sel.Find(".search_list_tidbit_short").Length() > 0 { //maybe for future use
			queryParsed.OnlineGame = true
		}*/

		sel.Find(".search_list_tidbit, .search_list_tidbit_short").Each(func(_ int, divSel *goquery.Selection) {
			gameType := strings.TrimSpace(divSel.Text())
			if gameType == "Main Story" || gameType == "Single-Player" || gameType == "Solo" {
				queryParsed.MainStory = []string{gameType, strings.TrimSpace(divSel.Next().Text())}
			}

			if gameType == "Main + Extra" || gameType == "Co-Op" {
				queryParsed.MainExtra = []string{gameType, strings.TrimSpace(divSel.Next().Text())}
			}

			if gameType == "Completionist" || gameType == "Vs." {
				queryParsed.Completionist = []string{gameType, strings.TrimSpace(divSel.Next().Text())}
			}
		})
		hltbQuery = append(hltbQuery, queryParsed)

	})
	return hltbQuery, nil
}
