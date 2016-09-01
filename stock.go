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
	Subjects   string  `json:"subjects"`
	SubjectTip string  `json:"subjectTip"`
	//BossInfo  string `json:"bossInfo"`
	LaunchDate string `json:"launchDate"`
	TimeStamp  string `json:"timeStamp"`
}

var (
	reAge1 = regexp.MustCompile(`.*([0-9]{4})年[月日出\d]*生,.*`)
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
	stock.fillXueqiuHot()
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

	stock.BossBirth = "unknow"
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
		return false
	})

	doc.Find("a").EachWithBreak(func(i int, s *goquery.Selection) bool {
		tarName := s.AttrOr("tar_name", "")
		if strings.EqualFold(tarName, "概念题材") {
			subjects := ""
			s.Find("td").Each(func(i int, ss *goquery.Selection) {
				_, isTip := ss.Attr("colspan")
				if isTip {
					stock.SubjectTip = ss.Text()
				} else {
					if len(subjects) > 0 {
						subjects = subjects + ", "
					}
					subjects = subjects + ss.Text()
				}
			})
			stock.Subjects = subjects
			return false
		}
		return true
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

	stock.TimeStamp = time.Now().Format("2006-01-02 15:04:05")
	return stock.fillCompanyInfo2()
}

func (stock *Stock) fillCompanyInfo2() error {
	charset := "gbk"
	//stockCode = "300340"
	url := fmt.Sprintf("http://basic.10jqka.com.cn/mobile/%s/company.html", stock.Code)
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

	doc.Find("tr").EachWithBreak(func(i int, s *goquery.Selection) bool {
		w01 := s.Find(".w01")
		tl := s.Find(".tl")
		if strings.EqualFold("上市日期", w01.Text()) {
			stock.LaunchDate = tl.Text()
			return false
		}
		return true
	})
	return err
}

func (stock *Stock) fillXueqiuHot() error {
	charset := "gbk"
	url := ""
	if strings.HasPrefix(stock.Code, "6") {
		url = fmt.Sprintf("https://xueqiu.com/S/SH%s/follows", stock.Code)
	} else {
		url = fmt.Sprintf("https://xueqiu.com/S/SZ%s/follows", stock.Code)
	}
	client := &http.Client{Jar: xueqiuJar}
	req, _ := http.NewRequest("GET", url, nil)
	rsp, err := client.Do(req)
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
	doc.Find(".stockInfo span").EachWithBreak(func(i int, s *goquery.Selection) bool {
		// For each item found, get the band and title
		fmt.Println("hot.txt", s.Text)
		return false
	})
	return err
}
