package main

type gSearchResponse struct {
	SearchInformation SearchInformation `json:"searchInformation"`
	Items []SearchResult `json:"items"`
	Error
}

type SearchInformation struct {
	TotalResults string `json:"totalResults"`
}

type SearchResult struct {
	Link string `json:"link"`
	Pagemap Pagemap `json:"pagemap"`
}

type Pagemap struct {
	VideoObject []VideoObjectItem `json:"videoobject"`
}


type VideoObjectItem struct {
	Title string `json:"name"`
	UploadDate string `json:"uploaddate"`
	VideoId string `json:"videoid"`
}

type Error struct {
	
}