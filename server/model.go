package server

import (
	"time"

	"github.com/btsay/repository"
)

type trend struct {
	Name       string
	ID         string
	Heat       int64
	Length     int64
	CreateTime time.Time
}

type recommend struct {
	ID   int
	Name string
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
