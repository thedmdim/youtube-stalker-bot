package models

type YtListResponse struct {
	Items []ytLisItem `json:"items"`
}

type ytLisItem struct {
	Id string `json:"id"`
	Statistics Statistics `json:"statistics"`
}

type Statistics struct {
	ViewCount string `json:"viewCount"`
}