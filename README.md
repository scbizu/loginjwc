## jwc_login

ZAFU JWC的教务处登陆接口

### implement

解耦`cmd_zafu`的模拟登陆模块,做成了Service来向前端/移动端提供登录的服务。

### how to use

> 详见example目录下的实现

|Route|Request Method|Request params|Response datatype|Response content|Others|
|---|---|---|---|---|----|
|/jwc/code|GET|EMPTY|JSON|{"href":"","cookie":"","VIEWSTATE":"","VIEWSTATEGENERATOR":""}|获取主页的一些必要信息|
|/jwc/login|POST|code,stuno,passwd, VIEWSTATE,VIEWSTATEGENERATOR,cookies(which Response to U in last step)|JSON|a certain name of the student|获取学生的名字|

你需要做的就是

* 发Request到/jwc/code 获取一些页面信息
* 发起POST请求的时候带上上一步返回的信息(别忘了学生帐号和密码)即可

### preview

You can test it at  [This Site](http://api.scnace.cc/jwc)
