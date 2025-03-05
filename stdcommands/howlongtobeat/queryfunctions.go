package howlongtobeat

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/mrbentarikau/pagst/commands"
	"github.com/mrbentarikau/pagst/common"
	"github.com/mrbentarikau/pagst/lib/jarowinkler"
	"github.com/mrbentarikau/pagst/stdcommands/util"

	"github.com/PuerkitoBio/goquery"
	"github.com/mediocregopher/radix/v3"
)

func getGameData(searchTitle string, recursiveRequest ...bool) (*HowlongToBeat, error) {
	var hltbQuery *HowlongToBeat

	u := &url.URL{
		Scheme:   hltbScheme,
		Host:     hltbHost,
		Path:     hltbHostPath + hltbAPIKey,
		RawQuery: hltbRawQuery,
	}

	urlStr := u.String()

	client := &http.Client{}

	var jsonQuery HowlongToBeatQuery
	jsonQuery.SearchTerms = []string{searchTitle}
	jsonQuery.SearchType = "games"
	jsonQuery.SearchPage = 1
	jsonQuery.Size = 20
	jsonQuery.UseCache = true
	jsonQuery.SearchOptions.Games.SortCategory = "popular"
	jsonQuery.SearchOptions.Games.RangeCategory = "main"
	jsonQuery.SearchOptions.Games.RangeCategory = "main"
	jsonQuery.SearchOptions.Games.Gameplay.Difficulty = ""
	jsonQuery.SearchOptions.Users.ID = hltbAPIKey

	jsonData, err := json.Marshal(jsonQuery)
	if err != nil {
		return nil, err
	}

	r, _ := http.NewRequest("POST", urlStr, strings.NewReader(string(jsonData))) // URL-encoded payload
	r.Header.Set("Content-Length", strconv.Itoa(len(jsonData)))
	r.Header.Set("User-Agent", common.ConfBotUserAgent.GetString())

	for k, v := range hltbHeaders {
		r.Header.Set(k, v)
	}

	resp, err := client.Do(r)
	if err != nil {
		return nil, err
	}

	var recursive bool
	if len(recursiveRequest) > 0 {
		recursive = recursiveRequest[0]
	}

	if resp.StatusCode != 200 {
		if resp.StatusCode == 404 && !recursive {
			hltbAPIKey, err = getAPIKey()
			if err != nil {
				return nil, err
			}

			hltbQuery, err = getGameData(searchTitle, true)
			if err != nil {
				return nil, err
			}
			return hltbQuery, nil
		} else {
			return nil, commands.NewPublicError("Unable to fetch data from howlongtobeat.com, status code:", resp.StatusCode)
		}
	}
	r.Body.Close()
	defer resp.Body.Close()

	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	queryErr := json.Unmarshal(bytes, &hltbQuery)
	if queryErr != nil {
		return nil, queryErr
	}

	return hltbQuery, nil
}

func parseQueryData(hltbQuery []HowlongToBeatData, gameName string) []HowlongToBeatData {

	for i, j := range hltbQuery {
		hltbQuery[i].CompMainDur = time.Second * time.Duration(j.CompMain)
		hltbQuery[i].CompPlusDur = time.Second * time.Duration(j.CompPlus)
		hltbQuery[i].Comp100Dur = time.Second * time.Duration(j.Comp100)

		hltbQuery[i].CompMainHumanize = common.HumanizeDurationShort(common.DurationPrecisionMinutes, hltbQuery[i].CompMainDur)
		hltbQuery[i].CompPlusHumanize = common.HumanizeDurationShort(common.DurationPrecisionMinutes, hltbQuery[i].CompPlusDur)
		hltbQuery[i].Comp100Humanize = common.HumanizeDurationShort(common.DurationPrecisionMinutes, hltbQuery[i].Comp100Dur)

		hltbQuery[i].JaroWinklerSimilarity = jarowinkler.Similarity([]rune(gameName), []rune(j.GameName))
		hltbQuery[i].GameURL = fmt.Sprintf("%sgame/%d", hltbURL, hltbQuery[i].GameID)
		hltbQuery[i].ImageURL = fmt.Sprintf("%sgames/%s", hltbURL, hltbQuery[i].GameImage)
	}

	return hltbQuery
}

func getAPIKey() (string, error) {
	headers := hltbHeaders
	headers["User-Agent"] = common.ConfBotUserAgent.GetString()

	// query the main page of HLTB for full HTML
	body, err := util.RequestFromAPI(hltbURL, headers)
	if err != nil {
		return "", err
	}

	toReader := bytes.NewReader(body)
	parseData, err := goquery.NewDocumentFromReader(toReader)
	if err != nil {
		return "", err
	}

	// finding specific JS file from HTML
	re := regexp.MustCompile(`_app-`)

	var apiKey, apiKeyURL string
	apiKeyURLSlice := make([]string, 0)
	parseData.Find("script").Each(func(_ int, sel *goquery.Selection) {
		attr, exists := sel.Attr("src")
		if exists && !strings.HasPrefix(attr, "https://cdn.ziffstatic.com") { // let's ignore JS from ziffstatic domain
			apiKeyURLSlice = append(apiKeyURLSlice, hltbURL+attr)
		}
	})

	apiKeyBuf := make([]string, 0)
	for _, u := range apiKeyURLSlice {
		if re.FindString(u) != "" {
			apiKeyURL = u
		} else {
			apiKeyBuf = append(apiKeyBuf, u)
		}
	}

	apiKeyURLSlice = nil
	apiKeyURLSlice = append(apiKeyURLSlice, apiKeyURL)
	apiKeyURLSlice = append(apiKeyURLSlice, apiKeyBuf...)

	var counter int
	var hltbHostPathBuf string
	for _, v := range apiKeyURLSlice {
		counter++
		apiKeyURL = v
		// getting contents of that JS file
		body, err = util.RequestFromAPI(apiKeyURL, headers)
		if err != nil {
			continue
		}

		// finding API key from JS file's contents
		// "/api/ouch/" was previously "/api/s && /api/lookup && /api/search" && /api/find
		// trying regex pattern /api/\w+/ to future-proof it...
		re = regexp.MustCompile(`"/(api/\w+/)"\.concat\("([[:alnum:]]+)"\)\.concat\("([[:alnum:]]+)"\)`)
		apiKeySlice := re.FindSubmatch(body)
		// going to look for second possible way for API key
		if len(apiKeySlice) == 0 {
			re = regexp.MustCompile(`users\s*:\s*{\s*id\s*:\s*"([^"]+)"`)
			apiKeySlice = re.FindSubmatch(body)
		}

		if len(apiKeySlice) >= 4 {
			apiKey = string(apiKeySlice[2]) + string(apiKeySlice[3])
			hltbHostPathBuf = string(apiKeySlice[1])
		} else if len(apiKeySlice) >= 3 {
			apiKey = string(apiKeySlice[2])
			hltbHostPathBuf = string(apiKeySlice[1])
		}

		if apiKey != "" {
			break
		}

	}

	logger.Infof("Parsed all HLTB JS files, count: %d", counter)

	if apiKey == "" && counter == 1 {
		logger.Errorf("Key for HLTB API not received, JS-scraping failed for: %s", apiKeyURL)
		return "", nil
	} else if apiKey == "" {
		return "", fmt.Errorf("key for HLTB API not received, JS-scraping failed for %d js files", len(apiKeyURLSlice))
	}

	logger.Infof("Storing new HLTB API Key to Redis: %s", apiKey)

	err = common.RedisPool.Do(radix.Cmd(nil, "SET", "hltb_APIKey", apiKey))
	if err != nil {
		return "", err
	}

	if hltbHostPath != hltbHostPathBuf {
		hltbHostPath = hltbHostPathBuf
		logger.Infof("Storing new HLTB search endpoint to Redis: %s", hltbHostPath)
		err = common.RedisPool.Do(radix.Cmd(nil, "SET", "hltb_HostPath", hltbHostPath))
		if err != nil {
			return "", err
		}
	}

	return apiKey, nil
}
