package resultcode

//返回码
const (
	SUCCESS     = 0  //成功
	ERR_NORMAL  = -1 //通用失败
	ERR_RELOGIN = -2 //失败需要重新登录
)
