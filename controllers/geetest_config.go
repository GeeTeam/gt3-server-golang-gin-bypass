package controllers

const (
	GEETEST_ID  = "c9c4facd1a6feeb80802222cbb74ca8e"
	GEETEST_KEY = "e4e298788aa8c768397639deb9b249a9"
	REDIS_SERVER = "127.0.0.1:6379"  // 对bypass状态进行缓存的redis服务地址
	BYPASS_URL = "http://bypass.geetest.com/v1/bypass_status.php"  // 向geetest发送获取bypass状态请求的url
	CYCLE_TIME = 10  // 轮询发送获取bypass状态请求的时间间隔(单位为秒)
	GEETEST_BYPASS_STATUS_KEY = "gt_server_bypass_status"  // bypass状态存入redis时使用的key值
)
