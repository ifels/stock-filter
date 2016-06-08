# stock-filter
stock filter

依赖：
  go get -v github.com/codegangsta/negroni
  go get -v github.com/unrolled/render
  go get -v gopkg.in/mgo.v2
  go get -v github.com/PuerkitoBio/goquery
  go get -v github.com/axgle/mahonia

运行:
1. cd stock-filter
2. go build
3. ./star.sh

demo:
http://ifels.cn:10000/stocks?sort=Price&market=sz&city=%E6%B7%B1%E5%9C%B3

参数:
sort,支持按以下字段递增排序：
    code, name, city, totalValue, tradeValue,price, bossName, bossBirth

market, 可选上海还是深圳：
    sz,  深圳
    sh,  上海

city, 城市
  如：深圳、北京、上海、广东等
