package youtube

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"	
	"strconv"
	"strings"
	"time"
	"youtube-stalker-bot/stats"
)

var ErrorApiQuota error = errors.New("YouTube API quota exceeded")

func request(url string) (*http.Response, error) {
	// Just wrap http.Get to add http code errors
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	if res.StatusCode == 429 {
		return nil, ErrorApiQuota
	}
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("%d %s", res.StatusCode, url)
	}
	return res, err
}

type Client struct {
	BaseURL string
	ApiKey     string
	SS *stats.Storage
	VideoQueue map[string]*Video
	MaxViews int
}

func NewClient(BaseURL string, ApiKey string, StatsStorage *stats.Storage, MaxViews int) *Client {
	return &Client{
		BaseURL: BaseURL,
		ApiKey: ApiKey,
		SS: StatsStorage,
		VideoQueue: make(map[string]*Video),
		MaxViews: MaxViews,
	}
}


func (c *Client) TakeFromQueue() (*Video, error) {
	for ; len(c.VideoQueue) == 0; {
		err := c.findVideos()
		if err != nil {
			return nil, err
		}
	}

	var RandomVideo *Video 
	for k, v := range c.VideoQueue {
		RandomVideo = v
		delete(c.VideoQueue, k)
		break
	}
	return RandomVideo, nil
}

func (c *Client) findVideos() error{
	var err error
	results := make(map[string]*Video)

	for {
		err = c.searchId(results)
		if err != nil {
			return err
		}
		if len(results)==0{
			continue
		}
		err = c.getViews(results)
		if err != nil {
			return err
		}
		
		if len(results)==0 {
			continue
		}
		c.filterVideo(results)
		if len(results)==0 {
			continue
		}
		for k, v := range results {
			c.VideoQueue[k] = v
		}
		return nil
	}
}

func (c *Client) randomYtId() string {
	/* 
	we don't need uppercase and downcase both presented
	because api search isn't case sensetive
	*/
	const base64range string = "0123456789abcdefghijklmnopqrstuvwxyz-_"
	var id []byte

	rand.Seed(time.Now().UnixNano())
	for i := 0; i < 5; i++ {
		id = append(id, base64range[rand.Intn(37)])
	}

	return string(id)
}

func (c *Client) searchId(results map[string]*Video) error {

	endpoint := "/search"

	params := fmt.Sprintf("?part=snippet&maxResults=50&type=video&key=%s&q=%s", c.ApiKey, url.QueryEscape("inurl:" + c.randomYtId()))
	res, err := request(c.BaseURL + endpoint + params)

	if err != nil {
		return err
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	var r YtSearchResponse
	err = json.Unmarshal(body, &r)
	if err != nil {
		return err
	}

	for _, item := range r.Items {
		video := new(Video)
		
		video.Id = item.Id.VideoId
		video.UploadDate = item.Snippet.PublishedAt
		video.Title = item.Snippet.Title

		results[video.Id] = video
	}

	c.SS.Days[0].ApiQueries+=1
	return nil
}

func (c *Client) getViews(videos map[string]*Video) error {

	endpoint := "/videos"

	var ids []string
	for _, video := range videos {
		ids = append(ids, video.Id)
	}

	params := fmt.Sprintf(`?key=%s&id=%s&part=statistics`, c.ApiKey, strings.Join(ids, ","))

	res, err := request(c.BaseURL + endpoint + params)

	if err != nil {
		return err
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	var r YtListResponse
	err = json.Unmarshal(body, &r)
	if err != nil {
		log.Println(err)
		return err
	}

	for _, item := range r.Items {

		viewCount, err := strconv.Atoi(item.Statistics.ViewCount)
		if err == nil {
			videos[item.Id].Views = viewCount
		}
	}

	c.SS.Days[0].ApiQueries+=1
	return nil
}

func (c *Client) filterVideo(videos map[string]*Video) {
	for k, v := range videos {
		if v.Views > c.MaxViews {
			delete(videos, k)
		}
	}
}
