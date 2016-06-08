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
    3. ./start.sh
```

demo:  
> http://ifels.cn:10000/stocks?sort=price&market=sz&city=深圳


参数:  
```
sort,支持按以下参数递增排序：  
    code,       股票代码
    name,       股票名称
    price,      股票当前价格
    totalValue, 总市值
    tradeValue, 流通市值
    city,       公司所在城市
    bossName,   董事长名称
    bossBirth   董事长年龄

market, 深圳或上海：  
    sz,  深圳
    sh,  上海

city, 城市： 
    如：深圳、北京、广东、江西等
```
