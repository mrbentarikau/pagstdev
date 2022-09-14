package howlongtobeat

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/botlabs-gg/yagpdb/commands"
	"github.com/botlabs-gg/yagpdb/common"
)

func getGameData(searchTitle string) ([]byte, error) {
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
		return nil, err
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
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, commands.NewPublicError("Unable to fetch data from howlongtobeat.com", resp.StatusCode)
	}
	r.Body.Close()
	defer resp.Body.Close()

	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return bytes, nil
}
