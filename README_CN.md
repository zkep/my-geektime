 [English](./README.md) | 中文
 

## 我的极客时间
极客时间资源下载工具

---
### 安装

```shell
go install github.com/zkep/mygeektime@latest
```

### 查看帮助
```shell
mygeektime -help
```

### 依赖项

#### [FFmpeg 处理视频](https://ffmpeg.org/download.html)

MacOS
```shell
brew install ffmpeg        
```
Linux
```shell
git clone https://github.com/FFmpeg/FFmpeg.git ffmpeg

cd ffmpeg

./configure --enable-ffplay --enable-ffserver

make && make install
```

### 模拟用户登录：

> 方式1: 浏览器开发者工具获取geektime有效cookie，存入项目执行目录 cookie.txt文件
    
> 方式2: [ChromeDriver](https://developer.chrome.google.cn/docs/chromedriver/get-started)
> 查看 Chrome 版本号, 在谷歌浏览器的地址栏输入：  
>```shell
>chrome://version
>```
> 获取到chrome版本号后，可以根据自己的系统替换下面的链接里的 ${version} 下载对应系统的版本
解压后将 chromedriver 文件放在程序执行目录
>* linux64：https://storage.googleapis.com/chrome-for-testing-public/${version}/linux64/chromedriver-linux64.zip
>* mac-arm64：https://storage.googleapis.com/chrome-for-testing-public/${version}/mac-arm64/chromedriver-mac-arm64.zip
>* mac-x64：https://storage.googleapis.com/chrome-for-testing-public/${version}/mac-x64/chromedriver-mac-x64.zip
>* win32：https://storage.googleapis.com/chrome-for-testing-public/${version}/win32/chromedriver-win32.zip
>* win64：https://storage.googleapis.com/chrome-for-testing-public/${version}/win64/chromedriver-win64.zip


#### 查看帮助输出
```shell
My GeekTime CLI 0.0.1

Available commands:

   server   This is http server 
   cli      This is Command 

Flags:

  -help
        Get help on the 'mygeektime' command.

```

#### 启动web服务

##### 默认配置文件
```yaml
server:
  app_name: My Geek Time
  run_mode: debug
  http_addr: 0.0.0.0
  http_port: 8090
jwt:
  secret: mygeektime-secret
  expires: 7200
database:
# driver: mysql
# source: root:123456@tcp(127.0.0.1:3306)/mygeektime?charset=utf8&parseTime=True&loc=Local&timeout=1000ms
# driver: postgres
# source: host=127.0.0.1 user=postgres password=123456 dbname=mygeektime port=5432 sslmode=disable TimeZone=Asia/Shanghai
  driver:  sqlite   # mysql|postgres|sqlite
  source:  mygeektime.db
  max_idle_conns: 10
  max_open_conns: 10
storage: # mp4 或 mp3 存储目录
  directory: repo # 自定义下载文件夹，默认执行目录下的repo目录
  driver: local
  bucket: object
  host: http://127.0.0.1:8090 # 端口与server中的 http_port 保持一致
browser:  # 
  driver_path: ./chromedriver
  cookie_path: ./cookie.txt
  open_browser: true
geektime:
  auto_sync: true # 建议开启，默认将geektime的接口数据缓存
```
##### 启动http服务
```shell
mygeektime server --config=config.yml
```

#### 感谢
* [gin](https://github.com/gin-gonic/gin)
* [amis](https://github.com/baidu/amis)
* [tebeka/selenium](https://github.com/tebeka/selenium)
* [gorm](https://github.com/go-gorm/gorm)
* [FFmpeg](https://ffmpeg.org/download.html) 
* [ChromeDriver](https://developer.chrome.google.cn/docs/chromedriver/get-started)