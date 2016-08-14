package server

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"unicode"

	"github.com/btlike/api/utils"
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
		utils.Log.Println(err)
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

		e := utils.Repostory.CreateHistory(keyword, r.RemoteAddr)
		if e != nil {
			utils.Log.Println(e)
		}

		var resp searchResp
		//返回所有视频都不存在
		if utils.Config.Pause {
			//如果在推荐列表中，直接搜索
			data, _ := utils.Repostory.GetRecommend()
			for _, v := range data {
				if keyword == v {
					goto pass
				}
			}
			if len(data) > 0 {
				//如果不在推荐列表中，在推荐列表中随机选择一个进行搜索
				index := rand.Intn(len(data))
				keyword = data[index]
				goto pass
			}
			w.Write(encoding(resp))
			return
		}

	pass:
		query := elastic.NewMatchPhrasePrefixQuery("Name", keyword)
		search := utils.ElasticClient.Search().Index("torrent").Query(query)
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
				trt, err := utils.Repostory.GetTorrentByInfohash(v.Id)
				if err != nil {
					continue
				}
				if len(trt.Name) == 0 {
					continue
				}
				var tdata esData
				err = json.Unmarshal(*v.Source, &tdata)
				if err != nil {
					utils.Log.Println(err)
				}
				trt.Heat = tdata.Heat
				resp.Torrent = append(resp.Torrent, trt)
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
		item, err := utils.Repostory.GetTorrentByInfohash(id)
		if err != nil {
			return
		}
		w.Write(encoding(item))
		return
	})

	mux.HandleFunc("/recommend", func(w http.ResponseWriter, r *http.Request) {
		var data []recommend
		rec, err := utils.Repostory.GetRecommend()
		if err != nil {
			return
		}
		for k, v := range rec {
			data = append(data, recommend{ID: k + 1, Name: v})
		}
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

	utils.Log.Println("running on", address)
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
