package main

type Update struct {
	UpdateId int `json:"update_id"`
	Message Message `json:"message"`
}

type Message struct {
	Chat Chat `json:"chat"`
	Text string `json:"text"`
}

type Chat struct {
	ChatId int64 `json:"id"`
}

type TgBotApiResponse struct {
	Result []Update `json:"result"`
}

type BotMessage struct {
	ChatId int64 `json:"chat_id"`
	Text string `json:"text"`
}