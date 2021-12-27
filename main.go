package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/axgle/mahonia"
	_ "github.com/go-sql-driver/mysql"
	"xorm.io/xorm"
)

type DealData struct {
	ExTime     string `xorm:"VARCHAR(255) NOT NULL PK" json:"extime"`  // Time.Time
	Code       string `xorm:"VARCHAR(6) NOT NULL PK" json:"code"`      // string
	Name       string `xorm:"VARCHAR(255) NOT NULL" json:"name"`       // string
	TClose     string `xorm:"VARCHAR(255) NOT NULL" json:"tclose"`     // float32
	High       string `xorm:"VARCHAR(255) NOT NULL" json:"high"`       // float32
	Low        string `xorm:"VARCHAR(255) NOT NULL" json:"low"`        // float32
	TOpen      string `xorm:"VARCHAR(255) NOT NULL" json:"topen"`      // float32
	LClose     string `xorm:"VARCHAR(255) NOT NULL" json:"lclose"`     // float32
	Chg        string `xorm:"VARCHAR(255) NOT NULL" json:"chg"`        // float32
	PChg       string `xorm:"VARCHAR(255) NOT NULL" json:"pchg"`       // float32
	Voturnover string `xorm:"VARCHAR(255) NOT NULL" json:"voturnover"` // int64
	Vaturnover string `xorm:"VARCHAR(255) NOT NULL" json:"vaturnover"` // float64
}

/*
type conn func() *xorm.Engine

var (
	db *conn
)
*/

func main() {

	var tmpresult map[string]map[string][]map[string]string
	content := Get("http://87.push2.eastmoney.com/api/qt/clist/get?cb=jQuery112406526563715394427_1631116233755&pn=1&pz=10000&po=1&np=1&ut=bd1d9ddb04089700cf9c27f6f7426281&fltt=2&invt=2&fid=f3&fs=m:0+t:6,m:0+t:80,m:1+t:2,m:1+t:23&fields=f12,f14")

	regx, _ := regexp.Compile(`.*?({.*}).*`)

	data := regx.FindAllStringSubmatch(content, -1)
	json.Unmarshal([]byte(data[0][1]), &tmpresult)

	result := tmpresult["data"]["diff"]
	for _, v := range result {
		code := v["f12"]
		t := code[:1]
		var prefix string
		if t == "3" || t == "0" {
			prefix = "1"
		}
		if t == "6" {
			prefix = "0"
		}
		if t == "4" || t == "8" {
			prefix = "2"
		}
		crawl(prefix + code)
	}

}

var db *xorm.Engine

func crawl(code string) interface{} {

	url := fmt.Sprintf("http://quotes.money.163.com/service/chddata.html?code=%s&start=19700101&end=&fields=TCLOSE;HIGH;LOW;TOPEN;LCLOSE;CHG;PCHG;VOTURNOVER;VATURNOVER", code)

	content := Get(url)
	bodystr := mahonia.NewDecoder("gbk").ConvertString(content)
	data := strings.Split(strings.Replace(bodystr, "'", "", -1), "\r\n")
	for i, v := range data[:len(data)-2] {
		if i > 0 {
			arr := strings.Split(strings.Replace(v, "None", "", -1), ",")
			ok, err := conn().Table("deal_data").Insert(&DealData{ExTime: arr[0], Code: arr[1], Name: arr[2], TClose: arr[3], High: arr[4], Low: arr[5], TOpen: arr[6], LClose: arr[7], Chg: arr[8], PChg: arr[9], Voturnover: arr[10], Vaturnover: arr[11]})
			if ok == 0 && err != nil {
				fmt.Println("处理数据出现错误,", err)
				continue
			}
			fmt.Println(strings.Split(v, ","))
		}
	}
	return true
}

func Get(url string) string {
	client := http.Client{Timeout: 5 * time.Second}
	resp, error := client.Get(url)
	defer resp.Body.Close()
	if error != nil {
		panic(error)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic("读取请求内容出错")
	}
	return string(body)
}

func conn() *xorm.Engine {
	connString := fmt.Sprintf("%s:%s@(%s:%d)/%s?%s", "root", "password", "127.0.0.1", 3306, "dealdata", "utf8mb4")
	db, err := xorm.NewEngine("mysql", connString)
	db.ShowSQL(true)
	if err != nil {
		log.Printf("Open %s failed, err:%v\n", "mysql", err)
	}
	err = db.Sync2(new(DealData))
	if err != nil {
		log.Printf("Sync database and fields failed:%v\n", err)
		panic(err)
	}
	return db
}
