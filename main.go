package main

import (
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"youtube-stalker-bot/stats"
	"youtube-stalker-bot/telegram"
	"youtube-stalker-bot/youtube"
	"youtube-stalker-bot/led"
)


const MaxViews int = 200

const gCloadApiUrl string = "https://www.googleapis.com/youtube/v3"
var gCloadApiToken string = os.Getenv("GCLOUD_TOKEN")

const tgBotApiUrl string = "https://api.telegram.org/bot"
var tgBotApiToken string = os.Getenv("TGBOT_TOKEN")

var ss *stats.Storage = stats.NewStorage()
var yt *youtube.Client = youtube.NewClient(gCloadApiUrl, gCloadApiToken, ss, 200)
var tg *telegram.Client = telegram.NewClient(tgBotApiUrl, tgBotApiToken)

func main(){
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	if gCloadApiToken == "" || tgBotApiToken == "" {
		log.Fatalf("Set GCLOUD_TOKEN and TGBOT_TOKEN env variables")
	}
	
	// event loop
	for ;; {
		updates, err := tg.Updates()

		if err != nil {
			log.Println(err)
		}
		for _, update := range updates {
			go processUpdate(&update)
			tg.Offset = update.UpdateId + 1
		}
	}
}


func processUpdate(update *telegram.Update){
	message := telegram.BotMessage{}
	message.ChatId = update.Message.Chat.ChatId
	
	if update.Message.Text == "/start" {
		message.Text = "Чтобы получить случайные видео, нажми /random"
	}

	if update.Message.Text == "/random" {
		ss.Days[0].Clicks+=1
		video, err := yt.TakeFromQueue()
	
		if errors.Is(err, youtube.ErrorApiQuota) {
			message.Text = "У меня кончились запросы к ютуб апи, попробуй попозже 😴"
		}
		if err != nil {
			log.Println(err)
		}

		message.Text = fmt.Sprintf("Title: %s\nViews: %d\nLink: https://www.youtube.com/watch?v=%s", video.Title, video.Views, video.Id)
	}

	if update.Message.Text == "/stats" {

		clicks := fmt.Sprintf("Нажали /random\n- Позавчера: %d\n- Вчера: %d\n- Сегодня: %d\n\n", 
			ss.Days[2].Clicks, ss.Days[1].Clicks, ss.Days[0].Clicks)

		queries := fmt.Sprintf("Запросов к YouTube API\n- Позавчера: %d\n- Вчера: %d\n- Сегодня: %d\n\n", 
			ss.Days[2].ApiQueries, ss.Days[1].ApiQueries, ss.Days[0].ApiQueries)

		inqueue := fmt.Sprintf("Видео в очереди: %d", len(yt.VideoQueue))

		message.Text = clicks + queries + inqueue
	}
	led.LedSwitch("default-on")
	tg.SendMessage(message)
	led.LedSwitch("none")
}