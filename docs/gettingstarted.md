# 快速开始 ⚡️

## 安装
* 目前支持三种安装方式, 首推 docker compose 方式
* 如果，你的本地已经有redis，mysql等服务了，也可以考虑docker方式和二进制发行包方式
* 再如果，你也是技术爱好者，恰好懂golang和amis的话，也可以clone源码安装调试

#### docker compose 方式，推荐指数 🌟🌟🌟🌟🌟
```shell
# 下载本项目
git clone https://github.com/zkep/mygeektime.git

# 切入docker compose 文件目录
cd mygeektime/docker

# 下载默认数据表
wget https://github.com/zkep/mygeektime/releases/download/v0.0.1/tasks.sql -O mysql/init/tasks.sql

# 后台启动服务
docker-compose up -d
```
服务启动后浏览器访问:  http://127.0.0.1:8090

#### docker 方式，推荐指数 🌟🌟🌟
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


#### [二进制包安装](https://github.com/zkep/mygeektime/releases) ，推荐指数 🌟🌟
下载对应操作系统的二进制包，下面以MacOS为例
```shell
# 下载
wget https://github.com/zkep/mygeektime/releases/download/v0.0.1/mygeektime_Darwin_arm64.tar.gz
# 解压
tar -zxvf mygeektime_Darwin_arm64.tar.gz

# 切入解压目录
cd mygeektime_Darwin_arm64

# 默认配置启动服务
./mygeektime server 

# 执行生成自定义配置模版命令，会生成 config_templete.yml文件，自行修改配置内容 
./mygeektime cli config

# 自定义配置启动服务
./mygeektime server --config=config_templete.yml

```
二进制方式，缺少一些依赖项，比如最重要的是ffmpeg，用于音视频合成的，需要自行安装后，加入环境变量中
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

[配置项](./config.md)  👉