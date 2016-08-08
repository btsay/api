package server

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/labstack/echo"
	"github.com/labstack/echo/engine/standard"
	"github.com/labstack/echo/middleware"
	"github.com/btlike/api/utils"
	"github.com/btlike/database/torrent"
	"gopkg.in/olivere/elastic.v3"
)

//define const
const (
	PageSize = 20
)

//ddefine var
var (
	VideoFormats = []string{"webm", "mkv", "flv", "vob", "ogv", "ogg", "drc", "gif",
		"gifv", "mng", "avi", "mov", "wmv", "yuv", "rm", "rmvb", "asf", "amv", "mp4", "m4p",
		"m4v", "mpg", "mp2", "mpeg", "mpe", "mpv", "m2v", "svi", "3gp", "3g2", "mxf", "roq", "nsv", "f4v",
		"f4p", "f4a", "f4b"}

	trends []trend
)

func isChineseChar(str string) bool {
	for _, r := range str {
		if unicode.Is(unicode.Scripts["Han"], r) {
			return true
		}
	}
	return false
}

func initTrend() (err error) {
	go func() {
		for {
			var ts []trend
			// dt := elastic.NewDateRangeAggregation()
			// dt = dt.BetweenWithKey("CreateTime", time.Now().Add(-time.Hour*24*30), time.Now())
			result, err := utils.Config.ElasticClient.Search().Index("torrent").Type("infohash").Sort("Heat", false).Size(1000).Do()
			if err != nil {
				utils.Log().Println(err)
				time.Sleep(time.Hour)
				continue
			}
			if result != nil && result.Hits != nil {
				for _, v := range result.Hits.Hits {
					var esdata esData
					json.Unmarshal(*v.Source, &esdata)

					// if !isChineseChar(esdata.Name) && len(esdata.Name) > 20 {
					// continue
					// }

					// if len(esdata.Name) > 20 {
					// continue
					// }

					exist, content := getTorrent(v.Id)
					if !exist {
						continue
					}
					var td torrentData
					err = json.Unmarshal([]byte(content), &td)
					if err != nil {
						continue
					}
					for _, file := range td.Files {
						if isVideo(file.Name) {
							ts = append(ts, trend{
								ID:         v.Id,
								Name:       td.Name,
								CreateTime: td.CreateTime,
								Length:     td.Length,
								Heat:       esdata.Heat,
							})
							if len(ts) >= 100 {
								trends = make([]trend, 0)
								for _, v := range ts {
									trends = append(trends, v)
								}
								goto done
							}
						}
						//只处理第一个文件（也是最大的文件）
						break
					}
				}
				trends = make([]trend, 0)
				for _, v := range ts {
					trends = append(trends, v)
				}
				goto done
			}
		done:
			time.Sleep(time.Hour)
		}
	}()
	return
}

//Run the server
func Run(address string) {
	err := initTrend()
	if err != nil {
		utils.Log().Println(err)
	}

	e := echo.New()
	e.Get("/list", func(c echo.Context) error {
		keyword := c.QueryParam("keyword")
		keyword, _ = url.QueryUnescape(keyword)
		if keyword == "" {
			return nil
		}

		var page int
		pg := c.QueryParam("page")
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
			return c.JSON(http.StatusOK, resp)
		}

		query := elastic.NewMatchQuery("Name", keyword)
		search := utils.Config.ElasticClient.Search().Index("torrent").Query(query)
		order := c.QueryParam("order")
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
			c.JSON(http.StatusInternalServerError, nil)
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

		return c.JSON(http.StatusOK, resp)
	})

	e.Get("/detail", func(c echo.Context) error {
		id := c.QueryParam("id")
		if id == "" {
			return nil
		}

		var item torrentData
		if utils.Config.Pause {
			return c.JSON(http.StatusOK, item)
		}

		has, content := getTorrent(id)
		if !has {
			return nil
		}

		err := json.Unmarshal([]byte(content), &item)
		if err != nil {
			utils.Log().Println(err)
			return err
		}

		return c.JSON(http.StatusOK, item)
	})

	e.Get("/recommend", func(c echo.Context) error {
		var data []torrent.Recommend
		utils.Config.Engine.OrderBy("id").Find(&data)
		return c.JSON(http.StatusOK, data)
	})

	e.Get("/trend", func(c echo.Context) error {
		return c.JSON(http.StatusOK, trends)
	})

	utils.Log().Println("running on", address)
	e.Use(middleware.CORS())
	e.Run(standard.New(address))
}

func isVideo(name string) bool {
	name = strings.TrimRight(name, ".")
	if name == "" {
		return false
	}

	if index := strings.LastIndex(name, "."); index > 0 {
		format := name[index+1:]
		for _, v := range VideoFormats {
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

func getTorrent(hash string) (exist bool, content string) {
	var has bool
	var err error
	switch hash[0] {
	case '0':
		var data torrent.Infohash0
		data.Infohash = hash
		has, err = utils.Config.Engine.Get(&data)
		if err != nil {
			return false, ""
		}
		if !has {
			return false, ""
		}
		content = data.Data
	case '1':
		var data torrent.Infohash1
		data.Infohash = hash
		has, err = utils.Config.Engine.Get(&data)
		if err != nil {
			return false, ""
		}
		if !has {
			return false, ""
		}
		content = data.Data
	case '2':
		var data torrent.Infohash2
		data.Infohash = hash
		has, err = utils.Config.Engine.Get(&data)
		if err != nil {
			return false, ""
		}
		if !has {
			return false, ""
		}
		content = data.Data
	case '3':
		var data torrent.Infohash3
		data.Infohash = hash
		has, err = utils.Config.Engine.Get(&data)
		if err != nil {
			return false, ""
		}
		if !has {
			return false, ""
		}
		content = data.Data
	case '4':
		var data torrent.Infohash4
		data.Infohash = hash
		has, err = utils.Config.Engine.Get(&data)
		if err != nil {
			return false, ""
		}
		if !has {
			return false, ""
		}
		content = data.Data
	case '5':
		var data torrent.Infohash5
		data.Infohash = hash
		has, err = utils.Config.Engine.Get(&data)
		if err != nil {
			return false, ""
		}
		if !has {
			return false, ""
		}
		content = data.Data
	case '6':
		var data torrent.Infohash6
		data.Infohash = hash
		has, err = utils.Config.Engine.Get(&data)
		if err != nil {
			return false, ""
		}
		if !has {
			return false, ""
		}
		content = data.Data
	case '7':
		var data torrent.Infohash7
		data.Infohash = hash
		has, err = utils.Config.Engine.Get(&data)
		if err != nil {
			return false, ""
		}
		if !has {
			return false, ""
		}
		content = data.Data
	case '8':
		var data torrent.Infohash8
		data.Infohash = hash
		has, err = utils.Config.Engine.Get(&data)
		if err != nil {
			return false, ""
		}
		if !has {
			return false, ""
		}
		content = data.Data
	case '9':
		var data torrent.Infohash9
		data.Infohash = hash
		has, err = utils.Config.Engine.Get(&data)
		if err != nil {
			return false, ""
		}
		if !has {
			return false, ""
		}
		content = data.Data
	case 'A':
		var data torrent.Infohasha
		data.Infohash = hash
		has, err = utils.Config.Engine.Get(&data)
		if err != nil {
			return false, ""
		}
		if !has {
			return false, ""
		}
		content = data.Data
	case 'B':
		var data torrent.Infohashb
		data.Infohash = hash
		has, err = utils.Config.Engine.Get(&data)
		if err != nil {
			return false, ""
		}
		if !has {
			return false, ""
		}
		content = data.Data
	case 'C':
		var data torrent.Infohashc
		data.Infohash = hash
		has, err = utils.Config.Engine.Get(&data)
		if err != nil {
			return false, ""
		}
		if !has {
			return false, ""
		}
		content = data.Data
	case 'D':
		var data torrent.Infohashd
		data.Infohash = hash
		has, err = utils.Config.Engine.Get(&data)
		if err != nil {
			return false, ""
		}
		if !has {
			return false, ""
		}
		content = data.Data
	case 'E':
		var data torrent.Infohashe
		data.Infohash = hash
		has, err = utils.Config.Engine.Get(&data)
		if err != nil {
			return false, ""
		}
		if !has {
			return false, ""
		}
		content = data.Data
	case 'F':
		var data torrent.Infohashf
		data.Infohash = hash
		has, err = utils.Config.Engine.Get(&data)
		if err != nil {
			return false, ""
		}
		if !has {
			return false, ""
		}
		content = data.Data
	}
	return true, content
}
