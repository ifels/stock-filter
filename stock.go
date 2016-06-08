package main

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/axgle/mahonia"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Stock struct {
	Code       string  `json:"code"`
	Name       string  `json:"name"`
	City       string  `json:"city"`
	TotalValue float32 `json:"totalValue"` //总市值
	TradeValue float32 `json:"tradeValue"` //流通市值
	Price      float32 `json:"price"`      //当前股价
	BossName   string  `json:"bossName"`
	BossBirth  string  `json:"bossBirth"`
	//BossInfo  string `json:"bossInfo"`
}

var (
	reAge1 = regexp.MustCompile(`.*([0-9]{4})年[^,]*生,.*`)
	reAge2 = regexp.MustCompile(`.*[^先]生于([0-9]{4})年.*`)
	reAge3 = regexp.MustCompile(`.*,([0-9]{2})岁,.*`)
)

func (stock *Stock) FillStockInfo() error {
	url := ""
	if strings.HasPrefix(stock.Code, "6") {
		url = fmt.Sprintf("http://qt.gtimg.cn/q=sh%s", stock.Code)
	} else {
		url = fmt.Sprintf("http://qt.gtimg.cn/q=sz%s", stock.Code)
	}
	log.Println(url)
	rsp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer rsp.Body.Close()

	charset := "gbk"
	if mahonia.GetCharset(charset) == nil {
		return fmt.Errorf("%s charset not suported \n", charset)
	}
	dec := mahonia.NewDecoder(charset)
	rd := dec.NewReader(rsp.Body)
	content, err := ioutil.ReadAll(rd)
	if err != nil {
		return err
	}
	str := string(content)
	r := regexp.MustCompile(`.*"([^"]*)".*`)
	str = r.ReplaceAllString(str, "$1")
	arr := strings.Split(strings.TrimSpace(str), "~")
	log.Println("content = ", str)
	if len(arr) < 46 {
		return fmt.Errorf("get stockInfo error \n")
	}
	if strings.EqualFold(stock.Code, arr[2]) {
		stock.Price = getFloat32(arr[3])
		if stock.Price == 0 {
			stock.Price = getFloat32(arr[4])
		}
		stock.TradeValue = getFloat32(arr[44])
		stock.TotalValue = getFloat32(arr[45])
	}
	return stock.fillCompanyInfo()
}

func getFloat32(value string) float32 {
	v, err := strconv.ParseFloat(value, 32)
	if err == nil {
		return float32(v)
	} else {
		return 0
	}
}

func (stock *Stock) fillCompanyInfo() error {
	charset := "gbk"
	//stockCode = "300340"
	url := fmt.Sprintf("http://basic.10jqka.com.cn/mobile/%s/companyn.html", stock.Code)
	log.Println(url)
	rsp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer rsp.Body.Close()

	if mahonia.GetCharset(charset) == nil {
		return fmt.Errorf("%s charset not suported \n", charset)
	}

	dec := mahonia.NewDecoder(charset)
	rd := dec.NewReader(rsp.Body)

	doc, err := goquery.NewDocumentFromReader(rd)
	if err != nil {
		return fmt.Errorf("when create from reader error %s ", err.Error())
	}
	doc.Find(".namebox").EachWithBreak(func(i int, s *goquery.Selection) bool {
		stock.BossName = strings.TrimSpace(s.Text())
		return false
	})

	doc.Find(".mng-intro").EachWithBreak(func(i int, s *goquery.Selection) bool {
		info := s.Find("p").Text()
		log.Println(info)

		//stock.BossInfo = strings.TrimSpace(info)
		if reAge1.MatchString(info) {
			stock.BossBirth = reAge1.ReplaceAllString(info, "$1")
			return false
		}
		if reAge2.MatchString(info) {
			stock.BossBirth = reAge2.ReplaceAllString(info, "$1")
			return false
		}
		if reAge3.MatchString(info) {
			ageStr := reAge3.ReplaceAllString(info, "$1")
			age, err := strconv.Atoi(ageStr)
			log.Println("age = ", age)
			if err == nil {
				stock.BossBirth = strconv.Itoa(time.Now().Year() - age)
				return false
			}
		}
		stock.BossBirth = "unknow"
		return false
	})

	doc.Find("tr").EachWithBreak(func(i int, s *goquery.Selection) bool {
		th := s.Find("th")
		td := s.Find("td")
		if strings.EqualFold("所属区域", th.Text()) {
			stock.City = td.Text()
			return false
		}
		return true
	})
	return err
}
