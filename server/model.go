package server

import (
	"time"

	"github.com/btlike/repository"
)

type trend struct {
	Name       string
	ID         string
	Heat       int64
	Length     int64
	CreateTime time.Time
}

type recommend struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type esData struct {
	Name       string
	Length     int64
	Heat       int64
	CreateTime time.Time
}

type searchResp struct {
	Torrent []repository.Torrent
	Count   int64
}

//Infohash define db model
type Infohash struct {
	ID   string `xorm:"'id'"`
	Data string `xorm:"'data'"`
}
