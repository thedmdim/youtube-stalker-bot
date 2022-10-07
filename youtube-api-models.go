package main

type ytResponse struct {
	Items []VideoInfo `json:"items"`
}

type VideoInfo struct {
	Id string `json:"id"`
	Snippet Snippet `json:"snippet"`
	ContentDetails ContentDetails `json:"contentDetails"`
	Statistics Statistics `json:"statistics"`
}

type Snippet struct {
	PublishedAt string `json:"publishedAt"`
	Title string `json:"title"`
}

type ContentDetails struct {
	Duration string `json:"duration"`
}

type Statistics struct {
	ViewCount string `json:"viewCount"`
}