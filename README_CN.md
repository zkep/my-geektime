 [English](./README.md) | 中文
 

## 我的极客时间
极客时间资源下载工具

---
### 安装
#### docker
```shell
docker run -p 8090:8090 -d --name mygeektime --restart always zkep/mygeektime:latest  server  
```

##### docker 挂载下载目录
使用宿主机目录替换下面的 ${directory}
```shell
docker run  -d  --restart always \
-p 8090:8090 \
-v  ${directory}:/repo \ 
--name mygeektime \
zkep/mygeektime:latest  server  
```
#### golang
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
# 生成配置模版后, 可以自定义配置
mygeektime cli config --config=config_templete.yml

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
  driver_path: chromedriver # 如果没有cookie文件，默认使用chromedriver模拟登录获取cookie
  open_browser: true # 服务启动后自动打开浏览器
geektime:
  auto_sync: true # 建议开启，默认将geektime的接口数据缓存
  auth_validate: true # 设置为false，不在登录时验证geektime的账号信息,可以读取已经缓存的下载任务
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

> 方式1: 浏览器开发者工具获取geektime有效cookie，存入项目执行目录 cookie.txt文件
    
> 方式2: [ChromeDriver](https://googlechromelabs.github.io/chrome-for-testing/#stable)
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


#### 感谢
* [gin](https://github.com/gin-gonic/gin)
* [amis](https://github.com/baidu/amis)
* [tebeka/selenium](https://github.com/tebeka/selenium)
* [gorm](https://github.com/go-gorm/gorm)
* [FFmpeg](https://ffmpeg.org/download.html) 
* [ChromeDriver](https://developer.chrome.google.cn/docs/chromedriver/get-started)