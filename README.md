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
--name mygeektime \
-v ${directory}:/repo  \
zkep/mygeektime:latest  server   
```
browser web url:  http://127.0.0.1:8090  

#### docker compose

```shell
git clone https://github.com/zkep/mygeektime.git

cd mygeektime

docker-compose up -d
```

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
  open_browser: true # Automatically open browser after service startup
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

> Method 1: The browser developer tool retrieves a valid Geektime cookie


#### thanks
* [gin](https://github.com/gin-gonic/gin)
* [amis](https://github.com/baidu/amis)
* [gorm](https://github.com/go-gorm/gorm)
* [FFmpeg](https://ffmpeg.org/download.html)
