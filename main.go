package main

import (
	. "youtube-stalker-bot/models"
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"
)

var requestMaxTries int = 10
var requestDelayOnFail int = 5
var views int = 200
var stats = make(map[int]int)


const gCloadApiUrl string = "https://www.googleapis.com/youtube/v3"
var gCloadApiToken string = os.Getenv("GCLOUD_TOKEN")

const tgBotApiUrl string = "https://api.telegram.org/bot"
var tgBotApiToken string = os.Getenv("TGBOT_TOKEN")

func main(){
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	var offset int
	if gCloadApiToken == "" || tgBotApiToken == "" {
		log.Fatalf("Set GCLOUD_TOKEN and TGBOT_TOKEN env variables")
	}
	for ;; {
		updates, err := getTelegramUpdates(offset)
		if err != nil {
			log.Println(err)
		}
		for _, update := range updates {
			go processUpdate(update)
			offset = update.UpdateId + 1
		}
	}
}

func processUpdate(update Update){
	message := BotMessage{}
	message.ChatId = update.Message.Chat.ChatId
	if update.Message.Text == "/start" {
		message.Text = "Чтобы получить случайные видео, нажми /random"
		botReply(message)
	}

	if update.Message.Text == "/random" {

		day := time.Now().Day()
		stats[day]+=1
		delete(stats, day-3)
		
		message.Text = "Начинаю рандомить 🎲"
		botReply(message)

		videos, err := findRandomVideos()
		
		if err != nil {
			switch {
			case errors.Is(err, ErrorApiQuotaExceeded):
				log.Println(err)
				message.Text = "У меня кончились запросы к ютуб апи, попробуй попозже 😴"
				botReply(message)
			default:
				log.Println(err)
			}
			return
		}
		message.Text = fmt.Sprintf("Найдено %d самых случайных видео", len(videos))
		botReply(message)
		time.Sleep(time.Second * 2)
		
		for _, v := range videos {
			message = BotMessage{
				ChatId: update.Message.Chat.ChatId,
				Text: fmt.Sprintf("title: %s\nviews: %d\nlink: https://www.youtube.com/watch?v=%s", v.Title, v.Views, v.Id),
			}
			botReply(message)
			time.Sleep(time.Second)
		}
	}

	if update.Message.Text == "/stats" {
		day := time.Now().Day()
		message.Text = fmt.Sprintf("Нажали /random\nПозавчера: %d\nВчера: %d\nСегодня: %d", stats[day], stats[day-1], stats[day-2])
		botReply(message)
	}
}

func findRandomVideos() (map[string]*Video, error) {
	var (
		id string
		err error
	)
	results := make(map[string]*Video)

	for try:=1;;try++{
		id = randomYtId()

		log.Printf("Randomly generated id part %s", id)

		err = searchId(results, id)
		if err != nil {
			return nil, err
		}
		
		log.Printf("Found %d videos", len(results))
		if len(results) == 0 {
			continue
		}

		log.Printf("Populating views")
		err = getViews(results)
		if err != nil {
			return nil, err
		}

		log.Printf("Filter videos with less than %d views", views)
		for k, v := range results {
			if v.Views > views {
				delete(results, k)
			}
		}
		
		log.Printf("Videos after filtering %d", len(results))
		if len(results) > 0 {
			return results, nil
		}	
	}
}

func randomYtId() string {
	const ytbase64range string = "0123456789abcdefghijklmnopqrstuvwxyz-_"
	var id []byte		
	
	rand.Seed(time.Now().UnixNano())	
	for i:=0; i<5; i++ {
		id = append(id, ytbase64range[rand.Intn(37)])
	}

	return string(id)
}

func searchId(results map[string]*Video, id string) error {
	var endpoint string = "/search"
	var params string = "?part=snippet&maxResults=50&type=video&key=" + gCloadApiToken + "&q=inurl%3A" + id
	resp, err := getRequest(gCloadApiUrl + endpoint + params)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var r YtSearchResponse
	err = json.Unmarshal(body, &r)
	if err != nil {
		return err
	}
	log.Printf("Found %d results", r.PageInfo.TotalResults)

	for _, item := range r.Items {
		video := new(Video)
		
		video.Id = item.Id.VideoId
		video.UploadDate = item.Snippet.PublishedAt
		video.Title = item.Snippet.Title

		results[video.Id] = video
	}
	return nil
}

func getViews(videos map[string]*Video) error {
	var ids string
	
	for _, video := range videos {
		ids += video.Id + ","
	}

	var endpoint string = "/videos"
	var params string = "?part=statistics&key=" + gCloadApiToken + "&id=" + ids 
	resp, err := getRequest(gCloadApiUrl + endpoint + params)
	if err != nil {
		log.Printf("getViews: %e", err)
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
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
		if err != nil {
			log.Println(err)
			return err
		}
		videos[item.Id].Views = viewCount
	}
	return nil
}

func getTelegramUpdates (offset int) ([]Update, error) {
	var endpoint string = "/getUpdates" + fmt.Sprintf("?offset=%d", offset)
	resp, err := getRequest(tgBotApiUrl + tgBotApiToken + endpoint)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var tgBotApiResponseJson TgBotApiResponse
	err = json.Unmarshal(body, &tgBotApiResponseJson)

	if err != nil {
		return nil, err
	}
	return tgBotApiResponseJson.Result, nil
}

func botReply(message BotMessage) error {
	log.Println("Telegram BOT reply")
	var endpoint string = "/sendMessage"
	jsonValue, _ := json.Marshal(message)

	ledSwitch("default-on")
	err := postRequest(tgBotApiUrl + tgBotApiToken + endpoint, bytes.NewBuffer(jsonValue))
	if err != nil {
		return err
	}
	ledSwitch("none")

	return nil
	
}

var ErrorApiQuotaExceeded = errors.New("YouTube API quota exceeded")

func getRequest(url string) (*http.Response, error) {
	var (
		resp *http.Response
		err error
	)
	for i:=0; i<requestMaxTries; i++ {
		resp, err = http.Get(url)
		if resp.StatusCode == 429 {
			return resp, ErrorApiQuotaExceeded
		}
		if resp.StatusCode != 200 {
			err = fmt.Errorf("POST %s %d", url, resp.StatusCode)
		}
		if err != nil {
			log.Printf("Error occured: %s, try: %d", err, i)
			time.Sleep(time.Second * 10)
			continue
		}
		return resp, nil
	}
	return nil, err
}

func postRequest(url string, body io.Reader) error {
	var (
		resp *http.Response
		err error
	)
	for i:=0; i<requestMaxTries; i++ {
		resp, err = http.Post(url, "application/json", body)
		if resp.StatusCode != 200 {
			err = fmt.Errorf("POST %s %d", url, resp.StatusCode)
		}
		if err != nil {
			log.Printf("Error occured: %s, try: %d", err, i)
			time.Sleep(time.Second * time.Duration(requestDelayOnFail))
			continue
		}
		return nil
	}
	return err
}

func ledSwitch(line string) {
	// LED on = default-on
	// LED off = none
	in, e := os.Create("/sys/devices/platform/gpio-leds/leds/F@ST2704N:red:inet/trigger")
	if e != nil {
		log.Println(e)
	}
	fmt.Fprint(in, line)
	defer in.Close()
}