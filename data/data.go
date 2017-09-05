package main

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/axgle/mahonia"
	"github.com/ifels/stock-filter/model"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	jar, _       = cookiejar.New(nil)
	xueqiuJar, _ = cookiejar.New(nil)
	url          = "http://hqdigi2.eastmoney.com/EM_Quote2010NumericApplication/index.aspx?type=s&sortType=C&sortRule=-1&pageSize=5000&page=1&jsName=quote_123&style=33&token=44c9d251add88e27b65ed86506f6e5da&_g=0.31431945857925436"
)

func main() {
	getStocks()
}

func getStocks() {
	initXueqiuCookie()
	log.Println("url = ", url)
	client := &http.Client{Jar: jar}
	req, _ := http.NewRequest("GET", url, nil)
	response, err := client.Do(req)

	if err != nil {
		log.Println("err = ", err)
	} else {
		defer response.Body.Close()
		content, err := ioutil.ReadAll(response.Body)
		if err != nil {
			log.Printf("%s", err)
		}
		parseStocks(string(content))
	}
	time.Sleep(time.Duration(3) * time.Hour)
}

func parseStocks(content string) error {
	log.Println("parse stocks....")
	r := regexp.MustCompile(`"([^"]*)"`)
	arr := r.FindAllString(content, -1)

	stockList := []model.Stock{}
	log.Println("arr.len = ", len(arr))
	for i, str := range arr {
		str = strings.Replace(str, `"`, "", -1)
		infos := strings.Split(str, ",")
		if len(infos) > 2 {
			stock := model.Stock{}
			stock.Code = infos[1]
			stock.Name = infos[2]
			log.Println(i, ":", stock.Code, stock.Name)
			stockList = append(stockList, stock)
		}
	}

	session, err := mgo.Dial("")
	if err != nil {
		log.Println("mgo.Dial error ", err.Error())
		return err
	}
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)

	for _, stock := range stockList {
		start := time.Now()

		c := session.DB("stock").C("items")
		query := bson.M{"code": stock.Code}
		savedStock := model.Stock{}
		c.Find(query).One(&savedStock)
		if len(savedStock.TimeStamp) > 0 {
			ts, err := time.ParseInLocation("2006-01-02 15:04:05", savedStock.TimeStamp, time.Local)
			if err == nil {
				duration := time.Now().Sub(ts)
				log.Printf("duration = %v", duration)
				if duration.Minutes() < 30 {
					//距上次抓取时间小于30分钟,则不抓取数据
					continue
				}
			} else {
				log.Println("parse ts error : ", err)
			}
		}
		err := stock.FillStockInfo()
		fillXueqiuHot(&stock)
		log.Printf("%+v\n", stock)
		if err != nil {
			log.Println("err %v", err)
		} else {
			c.Upsert(bson.M{"code": stock.Code}, &stock)
		}
		duration := time.Now().Sub(start)
		log.Println("duration = ", duration)

		time.Sleep(time.Duration(500) * time.Millisecond)
	}
	log.Println("..................................")
	return nil
}

func initXueqiuCookie() {
	client := &http.Client{Jar: xueqiuJar}
	req, _ := http.NewRequest("GET", "http://xueqiu.com", nil)
	client.Do(req)
}

func fillXueqiuHot(stock *model.Stock) error {
	charset := "utf8"
	url := ""
	if strings.HasPrefix(stock.Code, "6") {
		url = fmt.Sprintf("https://xueqiu.com/S/SH%s", stock.Code)
	} else {
		url = fmt.Sprintf("https://xueqiu.com/S/SZ%s", stock.Code)
	}
	client := &http.Client{Jar: xueqiuJar}
	req, _ := http.NewRequest("GET", url+"/follows", nil)
	req.Header.Set("Referer", url)
	rsp, err := client.Do(req)
	if err != nil {
		fmt.Println("err: ", err.Error())
		return err
	}
	fmt.Println("rsp: ", rsp)
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
		fmt.Println("hot.txt", s.Text())
		reg := regexp.MustCompile(`.*的粉丝\((\d+)\)人.*`)
		if reg.MatchString(s.Text()) {
			hot := reg.ReplaceAllString(s.Text(), "$1")
			fmt.Println("hot = ", hot)
			value, err := strconv.ParseInt(hot, 10, 64)
			if err == nil {
				stock.XueqiuHot = value
			}
		}
		return false
	})
	return err
}
