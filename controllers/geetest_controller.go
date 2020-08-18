package controllers

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"gt3-server-golang-gin-sdk/controllers/sdk"
	"github.com/gomodule/redigo/redis"
    "time"
	"fmt"
	"io/ioutil"
	"net/url"
	"errors"
	"encoding/json"
)


var pool *redis.Pool

// 建立redis连接池
func NewPool(server string) *redis.Pool {
    return &redis.Pool{
        MaxIdle:     3,
        MaxActive:   5,
        IdleTimeout: 240 * time.Second,
        Dial: func() (redis.Conn, error) {
            c, err := redis.Dial("tcp", server)
            if err != nil {
                c.Close()
                return nil, err
			}
            return c, err
        },
    }
}

// 发送GET请求
func httpGet(getURL string, params map[string]string) (string, error) {
	q := url.Values{}
	if params != nil {
			for key, val := range params {
					q.Add(key, val)
			}
	}
	req, err := http.NewRequest(http.MethodGet, getURL, nil)
	if err != nil {
			return "", errors.New("NewRequest fail")
	}
	req.URL.RawQuery = q.Encode()
	client := &http.Client{Timeout: time.Duration(5) * time.Second}
	res, err := client.Do(req)
	if err != nil {
			return "", err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
			return "", err
	}
	if res.StatusCode == 200 {
			return string(body), nil
	}
	return "", nil
}

// 从geetest获取bypass状态
func CheckBypassStatus(){
	redisStatus := "fail"
	for true{
		pool = NewPool(REDIS_SERVER)
		conn := pool.Get()
		defer conn.Close()
		params := make(map[string]string)
		params["gt"] = GEETEST_ID
		resBody, err := httpGet(BYPASS_URL, params)
		if resBody == ""{
			redisStatus = "fail"
		} else {
			resMap := make(map[string]interface{})
			err = json.Unmarshal([]byte(resBody), &resMap)
			if err != nil {
				redisStatus = "fail"
		}
		if resMap["status"] == "success" {
			redisStatus = "success"
		} else {
			redisStatus = "fail"
			}
		}
		s, err := conn.Do("SET", GEETEST_BYPASS_STATUS_KEY, redisStatus)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("bypass状态已经获取并存入redis,当前状态为-", redisStatus)
		fmt.Println(s)
		time.Sleep(time.Duration(CYCLE_TIME) * time.Second)
	}
}

// 获取redis缓存的bypass状态
func GetBypassCache() (status string) {
	pool = NewPool(REDIS_SERVER)
	conn := pool.Get()
	defer conn.Close()
	status, err := redis.String(conn.Do("GET", GEETEST_BYPASS_STATUS_KEY))
	if err != nil {
        fmt.Println(err)
        return
    }
	return status
}


// 验证初始化接口，GET请求
func FirstRegister(c *gin.Context) {
	/*
	   必传参数
	       digestmod 此版本sdk可支持md5、sha256、hmac-sha256，md5之外的算法需特殊配置的账号，联系极验客服
	   自定义参数,可选择添加
		   user_id 客户端用户的唯一标识，确定用户的唯一性；作用于提供进阶数据分析服务，可在register和validate接口传入，不传入也不影响验证服务的使用；若担心用户信息风险，可作预处理(如哈希处理)再提供到极验
		   client_type 客户端类型，web：电脑上的浏览器；h5：手机上的浏览器，包括移动应用内完全内置的web_view；native：通过原生sdk植入app应用的方式；unknown：未知
		   ip_address 客户端请求sdk服务器的ip地址
	*/
	bypassStatus := GetBypassCache()
	var result *sdk.GeetestLibResult
	gtLib := sdk.NewGeetestLib(GEETEST_ID, GEETEST_KEY)
	digestmod := "md5"
	userID := "test"
	params := map[string]string{
		"digestmod":   digestmod,
		"user_id":     userID,
		"client_type": "web",
		"ip_address":  "127.0.0.1",
	}
	if bypassStatus == "success" {
		result = gtLib.Register(digestmod, params)
	} else {
		result = gtLib.LocalRegister()
	}
	// 注意，不要更改返回的结构和值类型
	c.Header("Content-Type", "application/json;charset=UTF-8")
	c.String(http.StatusOK, result.Data)
}

// 二次验证接口，POST请求
func SecondValidate(c *gin.Context) {
	gtLib := sdk.NewGeetestLib(GEETEST_ID, GEETEST_KEY)
	challenge := c.PostForm(sdk.GEETEST_CHALLENGE)
	validate := c.PostForm(sdk.GEETEST_VALIDATE)
	seccode := c.PostForm(sdk.GEETEST_SECCODE)
	bypassStatus := GetBypassCache()
	var result *sdk.GeetestLibResult
	if bypassStatus == "success" {
		result = gtLib.SuccessValidate(challenge, validate, seccode)
	} else {
		result = gtLib.FailValidate(challenge, validate, seccode)
	}
	// 注意，不要更改返回的结构和值类型
	if result.Status == 1 {
		c.JSON(http.StatusOK, gin.H{"result": "success", "version": sdk.VERSION})
	} else {
		c.JSON(http.StatusOK, gin.H{"result": "fail", "version": sdk.VERSION, "msg": result.Msg})
	}
}
