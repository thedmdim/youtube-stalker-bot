package main

type ytSearchResponse struct {
	Items []ytSearchItem `json:"items"`
	PageInfo PageInfo `json:"pageInfo"`
}

type PageInfo struct {
	TotalResults int `json:"totalResults"`
}

type ytSearchItem struct {
	Id Id `json:"id"`
	Snippet Snippet `json:"snippet"`
}

type Id struct {
	VideoId string `json:"videoId"`
}

type Snippet struct {
	PublishedAt string `json:"publishedAt"`
	Title string `json:"title"`
}