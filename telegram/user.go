package telegram

type GetUser struct {
	Result User `json:"result"`
}

type User struct {
	Username string `json:"username"`
}