package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"youtube-stalker-bot/stats"
	"youtube-stalker-bot/telegram"
	"youtube-stalker-bot/youtube"
)


const MaxViews int = 200

const gCloadApiUrl string = "https://www.googleapis.com/youtube/v3"
var gCloadApiToken string = os.Getenv("GCLOUD_TOKEN")

const tgBotApiUrl string = "https://api.telegram.org/bot"
var tgBotApiToken string = os.Getenv("TGBOT_TOKEN")
var tgChannelId string = os.Getenv("TG_CHANNEL")


var ss *stats.Storage = stats.NewStorage()
var yt *youtube.Client = youtube.NewClient(gCloadApiUrl, gCloadApiToken, ss, 200)
var tg *telegram.Client = telegram.NewClient(tgBotApiUrl, tgBotApiToken)

func main(){

	if gCloadApiToken == "" || tgBotApiToken == "" || tgChannelId == "" {
		log.Fatalln("Set GCLOUD_TOKEN, TGBOT_TOKEN and TG_CHANNEL env variables")
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


func processUpdate(result *telegram.Result){
	
	if result == nil {
		return
	}

	message := telegram.OutgoingMessage{}
	message.ChatId = result.Message.Chat.ChatId

	defer func() {
        if r := recover(); r != nil {
            log.Println("Recovered: %", r)
        }
    }()

	if result.Message.Text == "/random" {
		ss.IncreaseTodaysClicks()

		video, err := yt.TakeFromQueue()
	
		if errors.Is(err, youtube.ErrorApiQuota) {
			message.Text = "У меня кончились запросы к ютуб апи, попробуй попозже 😴"
		}
		if err != nil {
			log.Println(err)
		}

		parts := strings.Split(strings.Split(video.UploadDate, "T")[0], "-")
		date := fmt.Sprintf("%s.%s.%s", parts[2], parts[1], parts[0])

		message.Text = fmt.Sprintf("Title: %s\nViews: %d\nPublished: %s\n\nhttps://www.youtube.com/watch?v=%s", video.Title, video.Views, date, video.Id)
	}

	if result.Message.Text == "/stats" {

		clicks := fmt.Sprintf("Нажали /random\n- Позавчера: %d\n- Вчера: %d\n- Сегодня: %d\n\n", 
			ss.Days[2].Clicks, ss.Days[1].Clicks, ss.Days[0].Clicks)

		queries := fmt.Sprintf("Запросов к YouTube API\n- Позавчера: %d\n- Вчера: %d\n- Сегодня: %d\n\n", 
			ss.Days[2].ApiQueries, ss.Days[1].ApiQueries, ss.Days[0].ApiQueries)

		inqueue := fmt.Sprintf("Видео в очереди: %d", len(yt.VideoQueue))

		message.Text = clicks + queries + inqueue
	}

	if strings.HasPrefix(result.Message.Text, "/post") {
		reply := result.Message.ReplyToMessage
		if result.Message.Text != "/post" || reply != nil {
			post := telegram.OutgoingMessage{}
			post.ChatId = json.Number(tgChannelId)
			if text := strings.Replace(result.Message.Text, "/post", "", 1); text != "" {
				post.Text += "\n\n" + text
			}
			if reply := result.Message.ReplyToMessage; reply != nil {
				post.Text += "\n\n" + reply.Text
			}
			post.Text += "\n\n@" + result.Message.From.Username
			tg.SendMessageBlink(post)
			message.Text = "Отправлено в предложку!"
		} else {
			message.Text = "Предложить пост:\n/post <текст>,\n/post в ответ на сообщение которое хотите прикрепить"
		}
	}
	tg.SendMessageBlink(message)
}