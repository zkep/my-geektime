 [English](./README_US.md) | 中文
 

## 我的极客时间
极客时间资源下载工具

---

### [文档](https://zkep.github.io/mygeektime/)

### [在线体验](https://mygeektime.anyfun.tech)

### 查看本地文档
```shell
git clone https://github.com/zkep/mygeektime.git

pip install mkdocs-material

mkdocs serve

```
浏览器访问:  http://127.0.0.1:8000/


### 安装

#### docker compose 方式, 该方式会启动mysql，redis等依赖服务

```shell
git clone https://github.com/zkep/mygeektime.git

cd mygeektime/docker

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

#### 扫码加入实战交流群

<img src="./web/public/wechat.jpg"  width="200" />
