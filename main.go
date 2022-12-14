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
var tgBotUsername string
var tgChannelUsername string
var tgChannelId string = os.Getenv("TG_CHANNEL")




var ss *stats.Storage = stats.NewStorage()
var yt *youtube.Client = youtube.NewClient(gCloadApiUrl, gCloadApiToken, ss, 200)
var tg *telegram.Client = telegram.NewClient(tgBotApiUrl, tgBotApiToken)

func main(){

	if gCloadApiToken == "" || tgBotApiToken == "" || tgChannelId == "" {
		log.Fatalln("Set GCLOUD_TOKEN, TGBOT_TOKEN and TG_CHANNEL env variables")
	}

	me, err := tg.GetMe()
	if err != nil {
		log.Fatalln("Can't get bot username: ", err)
	}
	tgBotUsername = me.Username

	chat, err := tg.GetChat(tgChannelId)
	if err != nil {
		log.Fatalln("Can't get bot username: ", err)
	}
	tgChannelUsername = chat.Username
	
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

	message := telegram.BotMessage{}
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

		message.Text = fmt.Sprintf("Title: %s\nViews: %d\nPublished: %s\nLink: https://www.youtube.com/watch?v=%s", video.Title, video.Views, date, video.Id)
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
		if reply := result.Message.ReplyToMessage; reply != nil {
			if reply.From.Username == tgBotUsername {
				tg.SendMessage(telegram.BotMessage{
					ChatId: json.Number(tgChannelId),
					Text: strings.Replace(result.Message.Text, "/post", "", 1) + "\n\n" + result.Message.ReplyToMessage.Text,
				})
				message.Text = fmt.Sprintf("Успешно отправлено в @%s", tgChannelUsername)
			} else {
				message.Text = "Напишите /post в ответ на моё сообщение"
			}
		} else {
			message.Text = "Напишите /post в ответ на сообщение, которое вы хотите запостить"
		}
	}
	tg.SendMessage(message)
}