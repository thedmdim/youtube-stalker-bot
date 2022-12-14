package telegram

import "encoding/json"

type Updates struct {
	Results []Result `json:"result"`
}

type Result struct {
	UpdateId int     `json:"update_id"`
	Message  Message `json:"message"`
}

type Message struct {
	From           From     `json:"from"`
	Chat           Chat     `json:"chat"`
	Text           string   `json:"text"`
	ReplyToMessage *Message `json:"reply_to_message"`
}

type Chat struct {
	ChatId json.Number `json:"id"`
}

type From struct {
	Username string `json:"username"`
}

type BotMessage struct {
	// BotMessage представляет собой сообщение которым отвечает бот
	ChatId json.Number `json:"chat_id"`
	Text   string      `json:"text"`
}