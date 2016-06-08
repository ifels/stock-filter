# stock-filter

依赖：  
```
    go get -v github.com/codegangsta/negroni
    go get -v github.com/unrolled/render
    go get -v gopkg.in/mgo.v2
    go get -v github.com/PuerkitoBio/goquery
    go get -v github.com/axgle/mahonia
```

运行:  
```
    1. cd stock-filter
    2. go build
    3. ./star.sh
```

demo:  
    http://ifels.cn:10000/stocks?sort=Price&market=sz&city=深圳

参数:  
sort,支持按以下字段递增排序：  
    code, name, city, totalValue, tradeValue,price, bossName, bossBirth

market, 深圳或上海：  
    sz,  深圳
    sh,  上海

city, 城市  
    如：深圳、北京、上海、广东等
