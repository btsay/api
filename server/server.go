package server

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/btlike/api/utils"
	"github.com/btlike/database/torrent"
	"github.com/rs/cors"
	"gopkg.in/olivere/elastic.v3"
)

//define const
const (
	PageSize = 20
)

//define var
var (
	videoFormats = []string{"webm", "mkv", "flv", "vob", "ogv", "ogg", "drc", "gif",
		"gifv", "mng", "avi", "mov", "wmv", "yuv", "rm", "rmvb", "asf", "amv", "mp4", "m4p",
		"m4v", "mpg", "mp2", "mpeg", "mpe", "mpv", "m2v", "svi", "3gp", "3g2", "mxf", "roq", "nsv", "f4v",
		"f4p", "f4a", "f4b"}
)

func isChineseChar(str string) bool {
	for _, r := range str {
		if unicode.Is(unicode.Scripts["Han"], r) {
			return true
		}
	}
	return false
}

func encoding(v interface{}) []byte {
	b, _ := json.Marshal(v)
	return b
}

//Run the server
func Run(address string) {
	err := getTrend()
	if err != nil {
		utils.Log().Println(err)
		return
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/list", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		keyword := r.Form.Get("keyword")
		keyword, _ = url.QueryUnescape(keyword)
		if keyword == "" {
			return
		}

		var page int
		pg := r.Form.Get("page")
		if pg == "" {
			page = 1
		} else {
			page, _ = strconv.Atoi(pg)
			if page == 0 {
				page = 1
			}
			if page > 20 {
				page = 20
			}
		}

		var history torrent.History
		history.Keyword = keyword
		_, e := utils.Config.Engine.Insert(&history)
		if e != nil {
			utils.Log().Println(e)
		}

		var resp searchResp
		//返回所有视频都不存在
		if utils.Config.Pause {
			//如果在推荐列表中，直接搜索
			var data []torrent.Recommend
			utils.Config.Engine.OrderBy("id").Find(&data)
			for _, v := range data {
				if keyword == v.Name {
					goto pass
				}
			}
			if len(data) > 0 {
				//如果不在推荐列表中，在推荐列表中随机选择一个进行搜索
				index := rand.Intn(len(data))
				keyword = data[index].Name
				goto pass
			}
			w.Write(encoding(resp))
			return
		}

	pass:
		query := elastic.NewMatchPhrasePrefixQuery("Name", keyword)
		search := utils.Config.ElasticClient.Search().Index("torrent").Query(query)
		order := r.Form.Get("order")
		if order == "l" {
			search = search.Sort("CreateTime", false)
		}
		if order == "m" {
			search = search.Sort("Length", false)
		}
		if order == "h" {
			search = search.Sort("Heat", false)
		}

		searchResult, err := search.
			From((page - 1) * PageSize).
			Size(PageSize).
			Do() // execute
		if err != nil {
			// Handle error
			w.WriteHeader(500)
		}

		if searchResult.Hits != nil {
			resp.Count = searchResult.Hits.TotalHits
			for _, v := range searchResult.Hits.Hits {
				has, content := getTorrent(v.Id)
				if !has {
					continue
				}
				var item torrentData
				err = json.Unmarshal([]byte(content), &item)
				if err != nil {
					utils.Log().Println(err)
					continue
				}
				var tdata esData
				err = json.Unmarshal(*v.Source, &tdata)
				if err != nil {
					utils.Log().Println(err)
				}
				item.Heat = tdata.Heat

				resp.Torrent = append(resp.Torrent, item)
			}
		}
		w.Write(encoding(resp))
		return
	})

	mux.HandleFunc("/detail", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()

		id := r.Form.Get("id")
		if id == "" {
			return
		}

		var item torrentData
		has, content := getTorrent(id)
		if !has {
			return
		}

		err := json.Unmarshal([]byte(content), &item)
		if err != nil {
			utils.Log().Println(err)
			return
		}
		w.Write(encoding(item))
		return
	})

	mux.HandleFunc("/recommend", func(w http.ResponseWriter, r *http.Request) {
		var data []torrent.Recommend
		utils.Config.Engine.OrderBy("id").Find(&data)
		w.Write(encoding(data))
		return
	})

	mux.HandleFunc("/trend", func(w http.ResponseWriter, r *http.Request) {
		tp := r.Form.Get("type")
		if tp == "week" {
			w.Write(encoding(weekTrends))
		} else {
			w.Write(encoding(monthTrends))
		}
		return
	})

	utils.Log().Println("running on", address)
	handler := cors.Default().Handler(mux)
	err = http.ListenAndServe(address, handler)
	if err != nil {
		panic(err)
	}
}

func isVideo(name string) bool {
	name = strings.TrimRight(name, ".")
	if name == "" {
		return false
	}

	if index := strings.LastIndex(name, "."); index > 0 {
		format := name[index+1:]
		for _, v := range videoFormats {
			if v == format {
				return true
			}
		}
	}
	return false
}

type trend struct {
	Name       string
	ID         string
	Heat       int64
	Length     int64
	CreateTime time.Time
}

type esData struct {
	Name       string
	Length     int64
	Heat       int64
	CreateTime time.Time
}

type searchResp struct {
	Torrent []torrentData
	Count   int64
}

type torrentData struct {
	Infohash   string
	Name       string
	CreateTime time.Time
	Length     int64
	FileCount  int64
	Heat       int64

	Files []file
}

type file struct {
	Name   string
	Length int64
}

//Infohash define db model
type Infohash struct {
	ID   string `xorm:"'id'"`
	Data string `xorm:"'data'"`
}

func getTorrent(hash string) (has bool, content string) {
	switch hash[0] {
	case '0':
		data := torrent.Infohash0{Infohash: hash}
		has, _ = utils.Config.Engine.Get(&data)
		content = data.Data
	case '1':
		data := torrent.Infohash1{Infohash: hash}
		has, _ = utils.Config.Engine.Get(&data)
		content = data.Data
	case '2':
		data := torrent.Infohash2{Infohash: hash}
		has, _ = utils.Config.Engine.Get(&data)
		content = data.Data
	case '3':
		data := torrent.Infohash3{Infohash: hash}
		has, _ = utils.Config.Engine.Get(&data)
		content = data.Data
	case '4':
		data := torrent.Infohash4{Infohash: hash}
		has, _ = utils.Config.Engine.Get(&data)
		content = data.Data
	case '5':
		data := torrent.Infohash5{Infohash: hash}
		has, _ = utils.Config.Engine.Get(&data)
		content = data.Data
	case '6':
		data := torrent.Infohash6{Infohash: hash}
		has, _ = utils.Config.Engine.Get(&data)
		content = data.Data
	case '7':
		data := torrent.Infohash7{Infohash: hash}
		has, _ = utils.Config.Engine.Get(&data)
		content = data.Data
	case '8':
		data := torrent.Infohash8{Infohash: hash}
		has, _ = utils.Config.Engine.Get(&data)
		content = data.Data
	case '9':
		data := torrent.Infohash9{Infohash: hash}
		has, _ = utils.Config.Engine.Get(&data)
		content = data.Data
	case 'A':
		data := torrent.Infohasha{Infohash: hash}
		has, _ = utils.Config.Engine.Get(&data)
		content = data.Data
	case 'B':
		data := torrent.Infohashb{Infohash: hash}
		has, _ = utils.Config.Engine.Get(&data)
		content = data.Data
	case 'C':
		data := torrent.Infohashc{Infohash: hash}
		has, _ = utils.Config.Engine.Get(&data)
		content = data.Data
	case 'D':
		data := torrent.Infohashd{Infohash: hash}
		has, _ = utils.Config.Engine.Get(&data)
		content = data.Data
	case 'E':
		data := torrent.Infohashe{Infohash: hash}
		has, _ = utils.Config.Engine.Get(&data)
		content = data.Data
	case 'F':
		data := torrent.Infohashf{Infohash: hash}
		has, _ = utils.Config.Engine.Get(&data)
		content = data.Data
	}
	return
}
