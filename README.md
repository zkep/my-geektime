English | [中文](./README_CN.md)

### my geektime
This is a tool to obtain the geektime video or articles with you geektime account

---
#### Install

#### install with  docker 
```shell
docker run -p 8090:8090 -d --name mygeektime --restart always zkep/mygeektime:latest  server  
```

##### docker with specify download directory
replace ${directory} with you local directory
```shell
docker run  -d  --restart always \
-p 8090:8090 \
-v  ${directory}:/repo \ 
--name mygeektime \
zkep/mygeektime:latest  server  
```
browser web url:  http://127.0.0.1:8090  


#### install with golang
```shell
go install github.com/zkep/mygeektime@latest
```
#### Start web service

##### Default configuration to start HTTP service
```shell
mygeektime server
```

##### Customize configuration to start HTTP service
```shell
# Generate a configuration template, and then customize the template content
mygeektime cli config --config=config_templete.yml

# Use custom configuration templates
mygeektime server --config=config_templete.yml
```

##### Default configuration file
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
storage: # mp4 or mp3 save folder
  driver: local
  directory: repo  # Customize download folder, default to execute repo directory under the directory
  bucket: object
  host: http://127.0.0.1:8090  # Keep the port consistent with the http_port in the server
browser:
  driver_path: chromedriver # If there is no cookie file, chromedriver will be used by default to simulate login and obtain cookies
  open_browser: true # Automatically open browser after service startup
geektime:
  auto_sync: true # Automatically sync geektime api data to db
  auth_validate: true # check geektime account auth
```

Commond help:
```shell
mygeektime -help
```

#### [FFmpeg](https://ffmpeg.org/download.html)

MacOS
```shell
brew install ffmpeg        
```
Linux
```shell
git clone https://github.com/FFmpeg/FFmpeg.git ffmpeg

cd ffmpeg

./configure --enable-gpl --enable-libx264

make && make install
```

### Simulate user login：

> Method 1: The browser developer tool retrieves a valid Geektime cookie and stores it in the project execution directory's cookie.txt file

> Method : [ChromeDriver](https://googlechromelabs.github.io/chrome-for-testing/#stable)
> Check the Chrome version number and enter it in the address bar of Google Chrome:
>```shell
>chrome://version
>```
> After obtaining the Chrome version number, you can replace the ${version} in the link below to download the corresponding system version according to your own system
After decompression, place the chromedriver file in the program execution directory
>* linux64：https://storage.googleapis.com/chrome-for-testing-public/${version}/linux64/chromedriver-linux64.zip
>* mac-arm64：https://storage.googleapis.com/chrome-for-testing-public/${version}/mac-arm64/chromedriver-mac-arm64.zip
>* mac-x64：https://storage.googleapis.com/chrome-for-testing-public/${version}/mac-x64/chromedriver-mac-x64.zip
>* win32：https://storage.googleapis.com/chrome-for-testing-public/${version}/win32/chromedriver-win32.zip
>* win64：https://storage.googleapis.com/chrome-for-testing-public/${version}/win64/chromedriver-win64.zip


#### thanks
* [gin](https://github.com/gin-gonic/gin)
* [amis](https://github.com/baidu/amis)
* [tebeka/selenium](https://github.com/tebeka/selenium)
* [gorm](https://github.com/go-gorm/gorm)
* [FFmpeg](https://ffmpeg.org/download.html)
* [ChromeDriver](https://developer.chrome.google.cn/docs/chromedriver/get-started)