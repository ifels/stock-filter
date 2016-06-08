package main

import (
	"github.com/codegangsta/negroni"
	"github.com/unrolled/render"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"syscall"
	"time"
)

var (
	r      = render.New(render.Options{})
	jar, _ = cookiejar.New(nil)
	url    = "http://hqdigi2.eastmoney.com/EM_Quote2010NumericApplication/index.aspx?type=s&sortType=C&sortRule=-1&pageSize=5000&page=1&jsName=quote_123&style=33&token=44c9d251add88e27b65ed86506f6e5da&_g=0.31431945857925436"
)

func main() {
	termChan := make(chan os.Signal, 1)
	signal.Notify(termChan, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)

	go getStocks()

	mux := http.NewServeMux()
	n := negroni.Classic()

	mux.HandleFunc("/stocks", handleStocksRequest)

	n.UseHandler(mux)
	n.Run(":10000")
	<-termChan
}

func handleStocksRequest(w http.ResponseWriter, req *http.Request) {
	req.ParseForm()
	session, err := mgo.Dial("")
	if err != nil {
		panic(err)
	}
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)

	sort := req.Form.Get("sort")
	city := req.Form.Get("city")
	market := req.Form.Get("market")
	if len(sort) > 0 {
		sort = strings.ToLower(sort)
	} else {
		sort = "totalvalue"
	}
	var stockList []Stock
	c := session.DB("stock").C("items")
	query := bson.M{}
	c.Find(query).Sort(sort).All(&stockList)
	//log.Println("pkgList.size = ", len(stockList))

	log.Println("city = ", city)
	log.Println("market = ", market)
	if len(city) == 0 && len(market) == 0 {
		r.JSON(w, http.StatusOK, stockList)
	} else {
		var result []Stock
		for _, stock := range stockList {
			if len(city) != 0 && !strings.Contains(stock.City, city) {
				continue
			}
			if len(market) != 0 {
				if (strings.EqualFold(market, "sz") && strings.HasPrefix(stock.Code, "6")) || (strings.EqualFold(market, "sh") && !strings.HasPrefix(stock.Code, "6")) {
					continue
				}
			}
			result = append(result, stock)
		}
		r.JSON(w, http.StatusOK, result)
	}

}

func getStocks() {
	for {
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
}

func parseStocks(content string) error {
	log.Println("parse stocks....")
	r := regexp.MustCompile(`"([^"]*)"`)
	arr := r.FindAllString(content, -1)

	stockList := []Stock{}
	log.Println("arr.len = ", len(arr))
	for i, str := range arr {
		str = strings.Replace(str, `"`, "", -1)
		infos := strings.Split(str, ",")
		if len(infos) > 2 {
			stock := Stock{}
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
		c := session.DB("stock").C("items")
		query := bson.M{"code": stock.Code}
		savedStock := Stock{}
		c.Find(query).One(&savedStock)
		if len(savedStock.TimeStamp) > 0 {
			ts, err := time.ParseInLocation("2006-01-02 15:04:05", savedStock.TimeStamp, time.Local)
			if err == nil {
				duration := time.Now().Sub(ts)
				log.Printf("duration = %v", duration)
				if duration.Hours() < 2 {
					//聚上次抓取时间小于两小时,则不抓取数据
					continue
				}
			} else {
				log.Println("parse ts error : ", err)
			}
		}
		err := stock.FillStockInfo()
		log.Printf("%+v\n", stock)
		if err != nil {
			log.Println("err %v", err)
		} else {
			c.Upsert(bson.M{"code": stock.Code}, &stock)
		}
		time.Sleep(time.Duration(8) * time.Second)
	}
	return nil
}
