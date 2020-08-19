package config

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"minGateway/util"
	"sync"
)

type ProxyConfig struct {
	CoreConf  Core                 `toml:"core"`
	LimitReq  LimitReq             `toml:"limitReq"`
	CcDefense CcDefense            `toml:"ccDefense"`
	LogConf   Log                  `toml:"log"`
	HttpProxy map[string]ProxyInfo `toml:"proxyInfo"`
	SslBase   SslBase              `toml:"sslBase"`
	SslCert   map[string]SslCert   `toml:"sslCert"`
}

/*
#最大连接数，为零，则不限流
limitMaxConn = 30000
#读超时,读完消息head和body的全部时间限制，为零，则没有超时，单位秒
readTimeout = 5
#写超时，从读完消息开始到消息返回的用时限制，为零，则没有超时，单位秒
writeTimeout = 10
#闲置超时，IdleTimeout是启用keep-alives状态后(默认启用)等待下一个请求的最长时间。
#如果IdleTimeout为零，则使用ReadTimeout的值。 如果两者均为零，则没有超时。单位秒
idleTimeout = 120
#最大头字节，为0则使用默认
maxHeaderBytes = 0
#主动防御，如"5,30,1800"设置为3秒钟内访问10次，则把此IP加入黑名单1800秒。未定义或为空则不开启
activeDefense = "5,20,1800"
#设置读取消息头中IP转发字段名称，按照数组的顺序查找，如果为空则获取的是TCP连接的IP地址
#如果消息是通过前端代理服务器转发或者cdn转发，则需要从消息头中获取IP地址（注意确保IP的真实性）
ipForwardeds = ["Ali-Cdn-Real-Ip","X-Forwarded-For","X-Real-Ip","X-Real-IP"]
*/
type Core struct {
	LimitMaxConn   int      `toml:"limitMaxConn"`
	ReadTimeout    int      `toml:"readTimeout"`
	WriteTimeout   int      `toml:"writeTimeout"`
	IdleTimeout    int      `toml:"idleTimeout"`
	MaxHeaderBytes int      `toml:"maxHeaderBytes"`
	IpForwardeds   []string `toml:"ipForwardeds"`
}

type LimitReq struct {
	Enable       bool `toml:"enable"`
	TimeDuration int  `toml:"timeDuration"`
	Count        int  `toml:"count"`
	Mode         int  `toml:"mode"`
}

type CcDefense struct {
	Enable       bool     `toml:"enable"`
	TimeDuration int      `toml:"timeDuration"`
	Count        int      `toml:"count"`
	BlackTime    int      `toml:"blackTime"`
	WhiteList    []string `toml:"whiteList"`
	BlackList    []string `toml:"blackList"`
}

type Log struct {
	LogLevel  string `toml:"logLevel"`
	WriteFile bool   `toml:"writeFile"`
	FileDir   string `toml:"fileDir"`
}

type ProxyInfo struct {
	Host       string   `toml:"host"`
	Target     []string `toml:"target"`
	ObtainMode int      `toml:"obtainMode"`
}

type SslBase struct {
	SessionTicket bool `toml:"sessionTicket"`
}

type SslCert struct {
	SslCertificate    string `toml:"ssl_certificate"`
	SslCertificateKey string `toml:"ssl_certificate_key"`
	//是否开启ocsp stapling，如果开启默认先去ssl证书平台拉取，失败再看本地是否有ocsp文件
	OcspStapling      bool   `toml:"ocsp_stapling"`
	OcspStaplingLocal bool   `toml:"ocsp_stapling_local"`
	OcspStaplingFile  string `toml:"ocsp_stapling_file"`
}

var (
	cfg  ProxyConfig
	once sync.Once
)

func Get() *ProxyConfig {
	once.Do(func() {
		cfg = ProxyConfig{}
		filePath := util.GetAbsolutePath("/bin/configs/conf.txt")
		if _, err := toml.DecodeFile(filePath, &cfg); err != nil {
			panic(err)
		}
		fmt.Printf("读取配置文件: %s\n", filePath)
	})
	return &cfg
}
