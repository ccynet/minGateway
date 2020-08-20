# minGateway

基于Go语言开发的一个小巧的http/https网关服务

## 优点：
- 部署简单，程序只有一个单独的文件，支持多个操作系统，即拷即用
- 性能优异，轻松实现每秒上万转发
- 可扩展性好，根据自身业务扩展起来方便（主要是因为代码量少）

## 功能：
- 负载均衡：支持三种目标路由方式可选，随机、轮询、哈希
- 多种转发规则：支持前逗号和后逗号泛路由转发
- 限流：设置最大连接数，实现限流功能
- CC攻击(Challenge Collapsar)防御：开启后记录频繁访问的IP，将其加入黑名单，具有一定的CC防御攻击能力
- 访问限制：开启后限制在一个时间范围内同IP访问同一URL请求次数，超过的会被抛弃
- HTTP(S)反向代理：由网关处理https加密连接，后端服务器只需提供非加密的http服务
- OCSP：可配置的OCSP Stapling功能
- 管理API：当前只有获取当前正在访问的数量，需要自行扩展

## 压测数据：
**内网环境：**
- 网关服务器(4核8g centos 7.x)，tomcat服务器(8核16g contos 7.x)
- tomcat提供GET的"/ping"请求，返回字符串"pong"
- 70微秒一次请求，约合每秒14285次请求，持续压测1分钟

**测试结果：**
- tomcat直压结果：共返回823842次成功结果，0次失败，约合每秒返回13730个成功结果
- 通过minGateway转发：共返回606926次成功结果，0次失败，约合每秒10115个成功结果

*另外，本次测试是在tomcat和minGateway的控制台都开着打印的情况，如果优化一下测试数据应该能再好一些。*

## 使用方法：

1. 配置go编译环境，本库在go 1.14中编译通过

2. 获取代码: `go get github.com/ccynet/minGateway`

3. 下载依赖: `cd minGateway && go mod tidy`

4. 编译程序: `go build`

5. 修改配置文件: 配置文件在minGateway的bin/configs/目录下

6. 如果需要设置TLS证书，证书文件保存在项目的bin/cert/目录下，并在配置文件中做相应设置

7. 运行编译出来的程序，请确保可执行文件放置在bin文件夹的同级目录下

## 附配置文件示例：

```toml
#服务器设置
[core]
#最大连接数，为0则不限流
limitMaxConn = 30000
#读超时,读完消息head和body的全部时间限制，为0则没有超时，单位秒
readTimeout = 5
#写超时，从读完消息开始到消息返回的用时限制，为0则没有超时，单位秒
writeTimeout = 30
#闲置超时，IdleTimeout是启用keep-alives状态后(默认启用)等待下一个请求的最长时间。
#如果IdleTimeout为零，则使用ReadTimeout的值。 如果两者均为零，则没有超时。单位秒
idleTimeout = 60
#最大头字节，为0则使用默认值1024k, 这里设置为131072=128k
maxHeaderBytes = 131072
#设置读取消息头中IP转发字段名称，按照数组里的顺序查找，如果为空则获取的是TCP连接的IP地址
#如果消息是通过前端代理服务器转发或者cdn转发，则需要从消息头中获取IP地址（注意确保IP的真实性）
ipForwardeds = ["Ali-Cdn-Real-Ip","X-Forwarded-For","X-Real-Ip","X-Real-IP"]

#CC防御(Challenge Collapsar)
#如设置为3000毫秒钟内访问100次，则把此IP加入黑名单3600秒
[ccDefense]
#是否开启 <value:false/true>
enable = false
#此时间的访问内做检查(单位:毫秒)
timeDuration = 3000
#允许访问的次数
count = 100
#放入黑名单时间(单位:秒)
blackTime = 3600
#IP白名单
whiteList = ["56.127.44.121","56.127.44.122"]
#IP黑名单
IPBlackList = []

#限制请求配置
#限制在一个时间范围内同IP访问同一URL请求次数，超过的会被抛弃
[limitReq]
#是否开启 <value:false/true>
enable = false
#此时间的访问内做检查(单位:毫秒)
timeDuration = 1000
#允许访问的次数
count = 1
#0:不包含参数 1:包含参数（为0时限制范围更广）
mode = 0

#log设置
[log]
#设置最低loglevel: debug,info,warn,error,fatal,panic
logLevel = "debug"
#开启写文件
writeFile = true
#存储日志文件的目录，不为空时是绝对路径，为空则是相对路径在程序的bin/logs/中
fileDir = ""

#代理设置
[proxyInfo]

#API服务器
[proxyInfo.shop]
#结尾有一个点的是泛解析，所有alogin开头的都走这里
host = "alogin."
#这里有2台服务器，用逗号间隔开
target = ["http://172.226.10.17:8080","http://172.226.10.19:8080"]
#服务器选择模式，1:随机方式  2:轮询方式  3:一致性哈希方式；如果未设置则使用随机方式
obtainMode = 3

#API服务器
[proxyInfo.alogin]
host = "alogin.obtc.com"
target = ["http://172.226.10.17:8080","http://172.226.10.19:8080"]
obtainMode = 3

#管理后台
[proxyInfo.admin]
host = "admin.obtc.com"
target = ["http://172.226.10.19:9150"]

#微信公众号
[proxyInfo.wxpay]
host = "wxpay."
target = ["http://172.226.10.17:9101"]

#开放平台,苹果拉微信
[proxyInfo.wxpaypay]
host = "wxpaypay.obtc.com"
target = ["http://172.226.10.17:9103"]

#简单网站和隐私协议
[proxyInfo.service]
host = "service.obtc.com"
target = ["http://172.226.10.17:9102"]

#都没找到走这里
[proxyInfo.default]
host = "default"
target = ["http://172.226.10.17:8080"]


[sslBase]
sessionTicket = true

#SSL证书文件
[sslCert]

[sslCert.wxpaypay_obtc_com]
ssl_certificate = "wxpaypay.obtc.com.crt"
ssl_certificate_key = "wxpaypay.obtc.com.key"
ocsp_stapling = true
ocsp_stapling_local = true
ocsp_stapling_file = "wxpaypay.obtc.com.ocsp"

[sslCert.wxpaypay_lidu_com]
ssl_certificate = "wxpaypay.lidu.com.crt"
ssl_certificate_key = "wxpaypay.lidu.com.key"

```

## Version
v0.1.0

## License
Licensed under the New BSD License.

## Author
Tom Chen (cgrencn@gmail.com)
