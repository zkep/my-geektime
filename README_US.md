English | [中文](./README.md)

### my geektime
This is a tool to obtain the geektime video or articles with you geektime account

---

### [Docs](https://zkep.github.io/mygeektime/)


### [Show Time](https://mygeektime.anyfun.tech)

### Local Docs
```shell
git clone https://github.com/zkep/mygeektime.git

pip install mkdocs-material

mkdocs serve
```
browser web url:  http://127.0.0.1:8000/

#### Install

#### install with docker compose

```shell
git clone https://github.com/zkep/mygeektime.git

cd mygeektime/docker

docker-compose up -d
```

browser web url:  http://127.0.0.1

#### install with  docker 
```shell
docker run  -d  --restart always \
--name mygeektime  \
-p 8090:8090 \
-v repo:/repo \
-v config.yml:/config.yml \
zkep/mygeektime:latest  server --config=config.yml
```

##### docker with specify download directory
replace ${directory} with you local directory
```shell
docker run  -d  --restart always \
-p 8090:8090 \
--name mygeektime \
-v config.yml:/config.yml \
-v ${directory}:/repo  \
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
# Use custom configuration templates
mygeektime server --config=config.yml
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

#### Join our communication group

<img src="./web/public/wechat.jpg"  width="200" />