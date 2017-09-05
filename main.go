package main

import (
	"github.com/codegangsta/negroni"
	"github.com/ifels/stock-filter/model"
	"github.com/unrolled/render"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"log"
	"net/http"
	"strings"
)

var (
	r = render.New(render.Options{})
)

func main() {
	mux := http.NewServeMux()
	n := negroni.Classic()
	mux.HandleFunc("/stocks", handleStocksRequest)
	n.UseHandler(mux)
	n.Run(":10000")
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
	subjects := req.Form.Get("subjects")
	if len(sort) > 0 {
		sort = strings.ToLower(sort)
	} else {
		sort = "totalvalue"
	}
	var stockList []model.Stock
	c := session.DB("stock").C("items")
	query := bson.M{}
	c.Find(query).Sort(sort).All(&stockList)
	//log.Println("pkgList.size = ", len(stockList))

	log.Println("city = ", city)
	log.Println("market = ", market)
	if len(city) == 0 && len(market) == 0 {
		r.JSON(w, http.StatusOK, stockList)
	} else {
		var result []model.Stock
		for _, stock := range stockList {
			if len(city) != 0 && !strings.Contains(stock.City, city) {
				continue
			}

			if len(subjects) != 0 && !strings.Contains(stock.Subjects, subjects) {
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
