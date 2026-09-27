package xkcd

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

type Comic struct {
	Title      string `json:"title,omitempty"`
	SafeTitle  string `json:"safe_title,omitempty"`
	Transcript string `json:"transcript,omitempty"`
	Img        string `json:"img,omitempty"`
	Alt        string `json:"alt,omitempty"`
	News       string `json:"news,omitempty"`
	Link       string `json:"link,omitempty"`
	Year       string `json:"year,omitempty"`
	Month      string `json:"month,omitempty"`
	Day        string `json:"day,omitempty"`
	Num        int    `json:"num,omitempty"`
}

const baseAPI = "https://xkcd.com"
const JSONAPI = "/info.0.json"

var ErrIndexNotFound = errors.New("index file not found")

func getInfo(client *http.Client, num int) (*http.Response, error) {
	var url string
	if num == 0 {
		url = fmt.Sprintf("%s%s", baseAPI, JSONAPI)
	} else {
		url = fmt.Sprintf("%s/%d%s", baseAPI, num, JSONAPI)
	}
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to get info for comic #%d", num)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("error with code: %d. Num query: %d", resp.StatusCode, num)
	}

	return resp, nil
}

func parseResponse(response *http.Response) (*Comic, error) {
	comic := Comic{}
	err := json.NewDecoder(response.Body).Decode(&comic)
	defer response.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to parse response")
	}

	return &comic, nil
}

func getMaxNum(client *http.Client) (int, error) {
	resp, err := getInfo(client, 0)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	endComic, err := parseResponse(resp)
	if err != nil {
		return 0, err
	}

	return endComic.Num, nil
}

func HandleComics(onProgress func(current, max int)) ([]Comic, error) {
	var err error
	client := http.Client{Timeout: 10 * time.Second}
	maximum, err := getMaxNum(&client)
	if err != nil {
		return nil, err
	}
	comics := make([]Comic, 0, maximum)
	for num := 1; num <= maximum; num++ {
		if num == 404 {
			comics = append(comics, Comic{})
			continue
		}
		resp, err := getInfo(&client, num)
		if err != nil {
			return nil, err
		}

		currentComic, err := parseResponse(resp)
		if err != nil {
			return nil, err
		}
		comics = append(comics, *currentComic)
		onProgress(num, maximum)
	}
	return comics, nil
}

func SaveIndex(comics []Comic, path string) error {
	data, err := json.MarshalIndent(comics, " ", "     ")
	if err != nil {
		return fmt.Errorf("failed to marshal data")
	}

	err = os.WriteFile(path, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write data to json file")
	}
	return nil
}

func LoadIndex(path string) ([]Comic, error) {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrIndexNotFound
		}
		return nil, fmt.Errorf("error accessing the file")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error while reading file")
	}
	var comics []Comic

	err = json.Unmarshal(data, &comics)
	if err != nil {
		return nil, fmt.Errorf("unmarshal data error")
	}

	return comics, nil
}

func match(c *Comic, term string) bool {
	lowerTerm := strings.ToLower(term)

	return strings.Contains(strings.ToLower(c.Title), lowerTerm) ||
		strings.Contains(strings.ToLower(c.SafeTitle), lowerTerm) ||
		strings.Contains(strings.ToLower(c.Transcript), lowerTerm) ||
		strings.Contains(strings.ToLower(c.Img), lowerTerm) ||
		strings.Contains(strings.ToLower(c.Alt), lowerTerm) ||
		strings.Contains(strings.ToLower(c.News), lowerTerm) ||
		strings.Contains(strings.ToLower(c.Link), lowerTerm) ||
		strings.Contains(c.Year, lowerTerm) ||
		strings.Contains(c.Month, lowerTerm) ||
		strings.Contains(c.Day, lowerTerm)

}

func Search(comics []Comic, term string) []Comic {
	res := make([]Comic, 0, len(comics))
	for _, i := range comics {
		if match(&i, term) {
			res = append(res, i)
		}
	}
	return res
}

func (c *Comic) String() string {
	return fmt.Sprintf("Comic: #%d\nTitle: %s\nImage: %s\nTranscription: %s\nURL: %s\nCreated at: %s-%s-%s",
		c.Num, c.Title, c.Img, c.Transcript, c.Link, c.Day, c.Month, c.Year)
}
