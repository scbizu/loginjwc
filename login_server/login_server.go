package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo"
	"github.com/labstack/echo/engine/standard"
	"github.com/scbizu/login_jwc"
)

const (
	//CodePath ..
	CodePath = "/jwc/"
	gate     = "http://210.33.60.5/"
)

//Code ...
type Code struct {
	//Href ...
	Href string `json:"href"`
	//Cookies ...
	Cookies string `json:"cookie"`
	//VIEWSTATE ...
	VIEWSTATE string `json:"VIEWSTATE"`
	//VIEWSTATEGENERATOR
	VIEWSTATEGENERATOR string `json:"VIEWSTATEGENERATOR"`
}

func main() {
	e := echo.New()
	e.Static("/", "example")
	e.File("/", "example/example.html")
	//get gif code
	e.GET("/jwc/code", func(c echo.Context) error {
		t := time.Now().UnixNano()
		hash := md5.New()
		hash.Write([]byte(strconv.Itoa(int(t))))
		token := hex.EncodeToString(hash.Sum(nil))

		loginS := jwclogin.NewGate(gate, "")
		sp, err := loginS.Getsp()
		if err != nil {
			return c.JSON(403, "cannot reach to jwc .")
		}
		var cookies []*http.Cookie
		err = json.Unmarshal([]byte(sp["cookie"]), &cookies)
		if err != nil {
			return c.JSON(403, "invaild cookie .")
		}
		loginS.GetVRCode(cookies, token)
		code := new(Code)

		code.Cookies = sp["cookie"]
		code.Href = CodePath + token
		code.VIEWSTATE = sp["VIEWSTATE"]
		code.VIEWSTATEGENERATOR = sp["VIEWSTATEGENERATOR"]
		jsoncode, err := json.Marshal(code)
		if err != nil {
			return c.JSON(403, "invaild checkcode .")
		}
		return c.JSON(200, string(jsoncode))
	})
	// post data
	e.Post("/jwc/login", func(c echo.Context) error {
		Gate := jwclogin.NewGate(gate, "")
		code := c.FormValue("code")
		stuno := c.FormValue("stuno")
		passwd := c.FormValue("passwd")
		VIEWSTATE := c.FormValue("VIEWSTATE")
		VIEWSTATEGENERATOR := c.FormValue("VIEWSTATEGENERATOR")
		cookie := c.FormValue("cookies")
		if code == "" || stuno == "" || passwd == "" || VIEWSTATE == "" {
			return c.JSON(403, "not enough params.")
		}
		stu := jwclogin.NewStu(stuno, passwd, Gate)
		client := new(http.Client)
		var cookies []*http.Cookie
		err := json.Unmarshal([]byte(cookie), &cookies)
		if err != nil {
			return c.JSON(403, "cookie invaild ")
		}
		flag, err := stu.Login(Gate.GateURL, client, code, VIEWSTATE, VIEWSTATEGENERATOR, cookies)
		if err != nil {
			return c.JSON(403, "bad post.")
		}
		return c.JSON(200, flag)
	})
	e.Run(standard.New(":8091"))
}
