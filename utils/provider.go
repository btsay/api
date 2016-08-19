package utils

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
)

var Keyword keyword

type keyword struct {
	WhiteList []string `json:"white_list"`
	BlackList []string `json:"black_list"`
}

func (p *keyword) InWhiteList(s string) bool {
	for _, v := range p.WhiteList {
		if strings.HasPrefix(v, s) {
			return true
		}
	}
	return false
}

func (p *keyword) InBlackList(s string) bool {
	for _, v := range p.BlackList {
		if strings.Contains(v, s) {
			return true
		}
	}
	return false
}

func initKeyword() {
	url := "http://obu2kw0g0.bkt.clouddn.com/Keyword.json"
	if len(Config.KeywordProvider) != 0 {
		url = Config.KeywordProvider
	}
	resp, err := http.Get(url)
	exit(err)
	data, err := ioutil.ReadAll(resp.Body)
	exit(err)
	err = json.Unmarshal(data, &Keyword)
	exit(err)
}
