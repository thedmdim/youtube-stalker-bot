package telegram

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type Client struct {
	BaseURL string
	ApiKey  string
	Offset int
}

func NewClient(BaseURL string, ApiKey string) *Client {
	return &Client{
		BaseURL: BaseURL,
		ApiKey: ApiKey,
	}
}

func (c *Client) Updates() ([]Result, error) {

	endpoint := "/getUpdates"

	params := fmt.Sprintf("?offset=%d", c.Offset)

	resp, err := http.Get(c.BaseURL + c.ApiKey + endpoint + params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var r Updates
	err = json.Unmarshal(body, &r)

	if err != nil {
		return nil, err
	}
	return r.Results, nil
}

func (c *Client) getUser(endpoint string, param string) (*User, error) {

	resp, err := http.Get(c.BaseURL + c.ApiKey + endpoint + param)

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var r GetUser
	err = json.Unmarshal(body, &r)

	if err != nil {
		return nil, err
	}
	return &r.Result, nil
}

func (c *Client) GetMe() (*User, error) {
	return c.getUser("/getMe", "")
}

func (c *Client) GetChat(chatId string) (*User, error) {
	return c.getUser("/getChat", fmt.Sprintf("?chat_id=%s", chatId))
}

func (c *Client) SendMessage(message BotMessage) error {
	
	endpoint := "/sendMessage"

	jsonValue, _ := json.Marshal(message)

	_, err := http.Post(c.BaseURL + c.ApiKey + endpoint, "application/json", bytes.NewBuffer(jsonValue))
	if err != nil {
		return err
	}
	return nil
}