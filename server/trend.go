package server

import (
	"encoding/json"
	"time"

	"github.com/btlike/api/utils"
	"gopkg.in/olivere/elastic.v3"
)

//format
const (
	TIME = "2006-01-02T15:04:05Z07:00"
)

var (
	monthTrends []trend
	weekTrends  []trend
)

func getMonthTrend(latest time.Time) {
	for {
		section := elastic.NewRangeQuery("CreateTime").Gte(latest.Add(-time.Hour * 24 * 30).Format(TIME))
		var trends []trend
		result, err := utils.Config.ElasticClient.Search().Index("torrent").Type("infohash").Query(section).Sort("Heat", false).Size(1000).Do()
		if err != nil {
			utils.Log().Println(err)
			time.Sleep(time.Hour)
			continue
		}
		if result != nil && result.Hits != nil {
			for _, v := range result.Hits.Hits {
				var esdata esData
				json.Unmarshal(*v.Source, &esdata)
				exist, content := getTorrent(v.Id)
				if !exist {
					continue
				}
				var td torrentData
				err = json.Unmarshal([]byte(content), &td)
				if err != nil {
					continue
				}
				if len(td.Name) > 40 {
					continue
				}
				for _, file := range td.Files {
					if isVideo(file.Name) {
						trends = append(trends, trend{
							ID:         v.Id,
							Name:       td.Name,
							CreateTime: esdata.CreateTime,
							Length:     td.Length,
							Heat:       esdata.Heat,
						})
						if len(trends) >= 100 {
							monthTrends = make([]trend, 0)
							for _, v := range trends {
								monthTrends = append(monthTrends, v)
							}
							goto done
						}
					}
					//只处理第一个文件（也是最大的文件）
					break
				}
			}
			monthTrends = make([]trend, 0)
			for _, v := range trends {
				monthTrends = append(monthTrends, v)
			}
			goto done
		}
	done:
		time.Sleep(time.Hour * 12)
	}
}

func getWeekTrend(latest time.Time) {
	for {
		section := elastic.NewRangeQuery("CreateTime").Gte(latest.Add(-time.Hour * 24 * 7).Format(TIME))
		var trends []trend
		result, err := utils.Config.ElasticClient.Search().Index("torrent").Type("infohash").Query(section).Sort("Heat", false).Size(1000).Do()
		if err != nil {
			utils.Log().Println(err)
			time.Sleep(time.Hour)
			continue
		}
		if result != nil && result.Hits != nil {
			for _, v := range result.Hits.Hits {
				var esdata esData
				json.Unmarshal(*v.Source, &esdata)
				exist, content := getTorrent(v.Id)
				if !exist {
					continue
				}
				var td torrentData
				err = json.Unmarshal([]byte(content), &td)
				if err != nil {
					continue
				}
				if len(td.Name) > 40 {
					continue
				}
				for _, file := range td.Files {
					if isVideo(file.Name) {
						trends = append(trends, trend{
							ID:         v.Id,
							Name:       td.Name,
							CreateTime: esdata.CreateTime,
							Length:     td.Length,
							Heat:       esdata.Heat,
						})
						if len(trends) >= 100 {
							weekTrends = make([]trend, 0)
							for _, v := range trends {
								weekTrends = append(weekTrends, v)
							}
							goto done
						}
					}
					//只处理第一个文件（也是最大的文件）
					break
				}
			}
			weekTrends = make([]trend, 0)
			for _, v := range trends {
				weekTrends = append(weekTrends, v)
			}
			goto done
		}
	done:
		time.Sleep(time.Hour * 6)
	}
}

func getTrend() (err error) {
	result, err := utils.Config.ElasticClient.Search().Index("torrent").Type("infohash").Sort("CreateTime", false).Size(1).Do()
	if err != nil {
		utils.Log().Println(err)
		return
	}

	var latest time.Time
	if result != nil && result.Hits != nil {
		for _, v := range result.Hits.Hits {
			var esdata esData
			json.Unmarshal(*v.Source, &esdata)
			latest = esdata.CreateTime
			if latest.IsZero() {
				latest = time.Now()
			}
		}
	}
	go getMonthTrend(latest)
	go getWeekTrend(latest)
	return
}
