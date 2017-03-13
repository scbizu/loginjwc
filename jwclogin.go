package jwclogin

import (
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"regexp"

	"github.com/scbizu/mahonia"
)

const (
	//模拟登陆第一个入口地址
	loginURLGate string = "http://210.33.60.5/"
	//模拟登陆第一个入口验证码地址
	vrcodeURLGate string = "http://210.33.60.5/CheckCode.aspx"
	//默认登录页
	defaultURL string = "http://210.33.60.5/default2.aspx"
)

//Student structure ...
type Student struct {
	//Username => student name
	Username string `json:"username"`
	//Password => password name
	Password string `json:"password"`

	*LoginGate
}

//LoginGate ...
type LoginGate struct {
	//GateURL
	GateURL string `json:"gateurl"`

	//DefaultGate
	DefaultGate string `json:"defaulturl"`
}

//NewGate init a gate
func NewGate(gate string, otherGate string) *LoginGate {
	logingate := new(LoginGate)
	logingate.DefaultGate = otherGate
	logingate.GateURL = gate
	return logingate
}

//NewStu load student information ..
func NewStu(stuno string, password string, gate *LoginGate) *Student {
	stu := new(Student)
	stu.Password = password
	stu.Username = stuno
	stu.LoginGate = gate
	return stu
}

//Getsp get the VIEWSTATE param ...
func (gate LoginGate) Getsp() (map[string]string, error) {
	view, err := http.Get(gate.GateURL)
	if err != nil {
		return nil, errors.New("发送请求失败(获取SP)")
	}
	cookie := view.Cookies()
	jsoncookie, err := json.Marshal(cookie)
	if err != nil {
		return nil, err
	}
	body, err := ioutil.ReadAll(view.Body)
	if err != nil {
		return nil, errors.New("获取body体失败啦～(获取SP)")
	}
	regular := `<input.type="hidden".name="__VIEWSTATE".value="(.*)" />`
	pattern := regexp.MustCompile(regular)
	VIEWSTATE := pattern.FindAllStringSubmatch(string(body), -1)
	//拿__VIEWSTATEGENERATOR
	retor := `<input.type="hidden".name="__VIEWSTATEGENERATOR".value="(.*)" />`
	patterntor := regexp.MustCompile(retor)
	VIEWSTATEGENERATOR := patterntor.FindAllStringSubmatch(string(body), -1)
	res := make(map[string]string)
	if len(VIEWSTATE) > 0 {
		res["VIEWSTATE"] = VIEWSTATE[0][1]
	} else {
		res["VIEWSTATE"] = ""
	}
	if len(VIEWSTATEGENERATOR) > 0 {
		res["VIEWSTATEGENERATOR"] = VIEWSTATEGENERATOR[0][1]
	} else {
		res["VIEWSTATEGENERATOR"] = ""
	}
	res["cookie"] = string(jsoncookie)
	return res, nil
}

//GetVRCode will fetch vrcode and save it ...
func (gate LoginGate) GetVRCode(cookies []*http.Cookie, token string) {
	gateURL, err := url.Parse(gate.GateURL)
	if err != nil {
		panic(err)
	}
	req, _ := http.NewRequest("GET", gateURL.Scheme+"://"+gateURL.Hostname()+"/CheckCode.aspx", nil)
	for _, v := range cookies {
		req.AddCookie(v)
	}
	// 获取验证码
	// var verifyCode string
	c := &http.Client{}
	res, err := c.Do(req)
	if err != nil {
		os.Exit(1)
	}
	file, er := os.Create("./example/" + token + ".gif")
	if er != nil {
		os.Exit(1)
	}
	io.Copy(file, res.Body)
}

//CodeReader will manually write the vrcode by user
// func CodeReader() string {
// 	var verifyCode string
// 	// for {
// 	fmt.Scanf("%s", &verifyCode)
// 	println(verifyCode)
// 	// if verifyCode != "" {
// 	// color.Green("验证码输入成功,正在请求教务处...")
// 	return verifyCode
// 	// }
// 	// }
// }

//Login post user info and return logined cookie(if success)
func (stu *Student) Login(Rurl string, c *http.Client, verifyCode string, VIEWSTATE string, VIEWSTATEGENERATOR string, tempCookies []*http.Cookie) (string, error) {

	postValue := url.Values{}
	cd := mahonia.NewEncoder("gb2312")
	rb := cd.ConvertString("学生")
	//准备POST的数据
	postValue.Add("txtUserName", stu.Username)
	postValue.Add("TextBox2", stu.Password)
	postValue.Add("txtSecretCode", verifyCode)
	postValue.Add("__VIEWSTATE", VIEWSTATE)
	postValue.Add("__VIEWSTATEGENERATOR", VIEWSTATEGENERATOR)
	postValue.Add("Button1", "")
	postValue.Add("lbLanguage", "")
	postValue.Add("hidPdrs", "")
	postValue.Add("hidsc", "")
	postValue.Add("RadioButtonList1", rb)
	//开始POST   这次POST到登陆界面   带上第一次请求的cookie 和 验证码  和 一些必要的数据
	postURL, _ := url.Parse(Rurl)
	Jar, _ := cookiejar.New(nil)
	Jar.SetCookies(postURL, tempCookies)
	c.Jar = Jar
	_, err := c.PostForm(Rurl, postValue)
	if err != nil {
		return "", errors.New("POST Response Lost")
	}
	// Scookies := resp.Cookies()
	// return Scookies, nil
	flag := stu.checkLogin(c, stu.Username)
	if flag != "fail" {
		cd := mahonia.NewDecoder("GBK")
		flag = cd.ConvertString(flag)
		return flag, nil
	}
	return "can not log in.", errors.New("Login failed")
}

func (gate *LoginGate) getStuname(c *http.Client, StuNo string) string {
	var restuName string
	LoggedURL := gate.GateURL + "/xs_main.aspx?xh=" + StuNo
	req, err := http.NewRequest("GET", LoggedURL, nil)
	if err != nil {
		panic(err)
	}
	finalRes, err := c.Do(req)
	if err != nil {
		panic(err)
	}
	allData, err := ioutil.ReadAll(finalRes.Body)
	if err != nil {
		panic(err)
	}
	defer finalRes.Body.Close()
	cd := mahonia.NewEncoder("gb2312")
	rb := cd.ConvertString("<span.id=\"xhxm\">(.*)同学</span>")
	//Regexp
	regular := rb
	pattern := regexp.MustCompile(regular)
	stuName := pattern.FindAllStringSubmatch(string(allData), -1)
	if len(stuName) > 0 {
		restuName = stuName[0][1]
	}
	return restuName
}

func (stu *Student) checkLogin(c *http.Client, stuno string) string {
	if stu.getStuname(c, stuno) != "" {
		return stu.getStuname(c, stuno)
	}
	return "fail"
}
