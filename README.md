 [English](./README_US.md) | 中文
 

## 我的极客时间
极客时间课程在线文档，不仅仅是下载器，还是在线文档，支持部署为在线服务，分享给你爱学习的小伙伴

---
特点：
 * 支持极客时间VIP账号一次缓存数据，永久观看
 * 支持一键发布为在线文档
 * 支持下载音视频资源到本地目录
 * 支持用户管理，轻松搭建共享服务

### [项目文档](https://zkep.github.io/mygeektime/)

### [在线体验](https://mygeektime.anyfun.tech)


### 安装

#### docker compose 方式, 该方式会启动mysql，redis等依赖服务

```shell
git clone https://github.com/zkep/mygeektime.git

cd mygeektime/docker

# 没有极客时间VIP的用户，执行下面命令，可以导入一些默认课程数据
# wget https://github.com/zkep/mygeektime/releases/download/v0.0.1/tasks.sql -O mysql/init/tasks.sql

docker-compose up -d
```
浏览器访问:  http://127.0.0.1:8090


#### docker 方式

##### docker 使用默认配置启动
```shell
docker run  -d  --restart always \
--name mygeektime  \
-p 8090:8090 \
zkep/mygeektime:latest  server
```
浏览器访问:  http://127.0.0.1:8090

##### docker 自定义配置启动
```shell
docker run  -d  --restart always \
--name mygeektime  \
-p 8090:8090 \
-v config.yml:/config.yml \
zkep/mygeektime:latest  server --config=config.yml
```
浏览器访问:  http://127.0.0.1:8090

##### docker 挂载下载目录启动
使用宿主机目录替换下面的 ${directory}
```shell
docker run -d --restart always \
-p 8090:8090 \
--name mygeektime \
-v ${directory}:/repo  \
-v config.yml:/config.yml \
zkep/mygeektime:latest  server --config=config.yml
```
浏览器访问:  http://127.0.0.1:8090



#### golang 方式
```shell
go install github.com/zkep/mygeektime@latest
```

#### 启动web服务

##### 默认配置启动http服务
```shell
# 使用默认配置启动
mygeektime server
```

##### 自定义配置启动http服务
```shell
# 使用自定义配置启动服务
mygeektime server --config=config.yml
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

./configure --enable-gpl --enable-libx264

make && make install
```

#### 感谢
* [gin](https://github.com/gin-gonic/gin)
* [amis](https://github.com/baidu/amis)
* [gorm](https://github.com/go-gorm/gorm)
* [FFmpeg](https://ffmpeg.org/download.html)

