package main

import (
	"bytes"
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
var views int = 200

var gCloadApi string = os.Getenv("GCLOUD_TOKEN")
const cseId string = "41646304e52844d02"

const tgBotApiUrl string = "https://api.telegram.org/bot"
var tgBotApiToken string = os.Getenv("TGBOT_TOKEN")

func main(){
	var offset int
	if gCloadApi == "" || tgBotApiToken == "" {
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
	var message BotMessage
	if update.Message.Text == "/start" {
		message = BotMessage{
			ChatId: update.Message.Chat.ChatId,
			Text: "Получить случайные видео\nTo get random video\n/random",
		}
		botReply(message)
	}

	if update.Message.Text == "/random" {
		videos := findRandomVideos()
		message = BotMessage{
			ChatId: update.Message.Chat.ChatId,
			Text: fmt.Sprintf("Randomly found %d", len(videos)),
		}
		botReply(message)
		time.Sleep(time.Second * 2)
		for _, v := range videos{
			message = BotMessage{
				ChatId: update.Message.Chat.ChatId,
				Text: fmt.Sprintf("Title: %s\nViews: %d\nLink: https://www.youtube.com/watch?v=%s", v.Title, v.Views, v.Id),
			}
			botReply(message)
			time.Sleep(time.Second)
		}
	}
}


func findRandomVideos() map[string]*Video {
	for try:=1;;try++{
		log.Printf("Find random video try number %d", try)

		id := randomYtId()

		log.Printf("Search id %s with CSE", id)
		results, err := searchId(id)
		if err != nil {
			log.Println(err)
			continue
		}
		
		log.Printf("Found %d videos", len(results))
		if len(results) == 0 {
			continue
		}
		log.Printf("Populating views")
		err = getViews(results)
		if err != nil {
			log.Println(err)
			continue
		}

		log.Printf("Filter videos with less than %d views", views)
		for k, v := range results {
			if v.Views > views {
				delete(results, k)
			}
		}
		
		log.Printf("Videos after filtering %d", len(results) > 0)
		if len(results) > 0 {
			return results
		}		
	}
}

// get random youtube video id
func randomYtId() string {
	const ytbase64range string = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz-_"
	var id string		
	
	rand.Seed(time.Now().UnixNano())	
	for i:=0; i<5; i++ {
		id+=string(ytbase64range[rand.Intn(63)])
	}

	log.Printf("Randomly generated id part %s", id)
	return id
}


func searchId(id string) (map[string]*Video, error) {
	log.Println("Searching video id")

	urlQuery := fmt.Sprintf("https://www.googleapis.com/customsearch/v1?key=%s&cx=%s&q=inurl%%3A%s", gCloadApi, cseId, id)
	resp, err := getRequest(urlQuery)
	if err != nil {
		log.Printf("searchId %e", err)
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	var r gSearchResponse
	err = json.Unmarshal(body, &r)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	log.Printf("Found %s results", r.SearchInformation.TotalResults)

	result := make(map[string]*Video)
	for _, item := range r.Items {
		video := new(Video)
		
		video.Id = item.Pagemap.VideoObject[0].VideoId
		video.Link =  item.Link
		video.UploadDate = item.Pagemap.VideoObject[0].UploadDate
		video.Title = item.Pagemap.VideoObject[0].Title

		result[item.Pagemap.VideoObject[0].VideoId] = video
	}
	return result, nil
}


func getViews(videos map[string]*Video) error {
	log.Println("Getting views")

	var ids string
	for _, video := range videos {
		ids += video.Id + ","
	}
	urlQuery := fmt.Sprintf("https://www.googleapis.com/youtube/v3/videos?key=%s&part=statistics,contentDetails,snippet&id=%s", gCloadApi, ids)

	resp, err := getRequest(urlQuery)
	if err != nil {
		log.Printf("getViews: %e", err)
		return err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var r ytResponse
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
	err := postRequest(tgBotApiUrl + tgBotApiToken + endpoint, bytes.NewBuffer(jsonValue))
	if err != nil {
		return err
	}
	return nil
	
}

func getRequest(url string) (*http.Response, error) {
	var (
		resp *http.Response
		err error
	)
	for i:=0; i<requestMaxTries; i++ {
		log.Printf("GET %s try %d", url, i)
		resp, err = http.Get(url)
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			err = errors.New(fmt.Sprintf("GET %s %d", url, resp.StatusCode))
			continue
		}
		if err != nil {
			log.Printf("Error occured: %e, try: %d", err, i)
			time.Sleep(time.Second * time.Duration(requestMaxTries))
			continue
		}
	}
	return resp, err
}

func postRequest(url string, body io.Reader) error {
	var (
		resp *http.Response
		err error
	)
	for i:=0; i<requestMaxTries; i++ {
		log.Printf("POST %s try %d", url, i)
		resp, err = http.Post(url, "application/json", body)
		if resp.StatusCode != 200 {
			err = errors.New(fmt.Sprintf("POST %s %d", url, resp.StatusCode))
			continue
		}
		if err != nil {
			log.Printf("Error occured: %e, try: %d", err, i)
			time.Sleep(time.Second * time.Duration(requestMaxTries))
			continue
		}
	}
	return err
}