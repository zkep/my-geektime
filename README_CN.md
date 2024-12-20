 [English](./README.md) | 中文
 

## 我的极客时间
极客时间资源下载工具

---
### 安装
#### docker 方式
```shell
docker run -p 8090:8090 -d --name mygeektime --restart always zkep/mygeektime:latest  server  
```

##### docker 挂载下载目录
使用宿主机目录替换下面的 ${directory}
```shell
docker run -d --restart always \
-p 8090:8090 \
--name mygeektime \
-v ${directory}:/repo  \
zkep/mygeektime:latest server   
```
浏览器访问:  http://127.0.0.1:8090

#### docker compose

```shell
git clone https://github.com/zkep/mygeektime.git

cd mygeektime

docker-compose up -d
```
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
mygeektime server --config=config_templete.yml
```

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
  open_browser: true # 服务启动后自动打开浏览器

```


### 查看帮助
```shell
mygeektime -help
```
#### 查看帮助输出
```shell
My GeekTime CLI 0.0.1

Available commands:

   server   This is http server 
   cli      This is command 

Flags:

  -help
        Get help on the 'mygeektime' command.
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

### 模拟用户登录：

> 方式1: 浏览器开发者工具获取geektime有效cookie

#### 感谢
* [gin](https://github.com/gin-gonic/gin)
* [amis](https://github.com/baidu/amis)
* [gorm](https://github.com/go-gorm/gorm)
* [FFmpeg](https://ffmpeg.org/download.html)