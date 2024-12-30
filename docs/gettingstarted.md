# 快速开始 ⚡️

## 安装
* 目前支持三种安装方式, 首推docker compose 方式
* 如果，你的本地已经有redis，mysql等服务了，也可以考虑docker方式和二进制发行包方式
* 再如果，你也是技术爱好者，恰好懂golang和amis的话，也可以clone源码安装调试

#### docker 方式

##### docker 使用默认配置启动,默认使用sqlite数据库，如果使用其他数据库，使用下面的自定义配置启动
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
zkep/mygeektime:latest server   
```
浏览器访问:  http://127.0.0.1:8090

#### docker compose 方式, 该方式会启动 nginx，mysql，redis等依赖服务
```shell
# 下载本项目
git clone https://github.com/zkep/mygeektime.git

# gitee 也会同步更新，没有网络环境的小伙伴可以用这个方式
# git clone https://gitee.com/zkep/mygeektime.git

# 切入docker compose 文件目录
cd mygeektime/docker

# 后台启动服务
docker-compose up -d
```
服务启动后浏览器访问:  http://127.0.0.1


## docker 安装
${directory} 是宿主机的音视频下载目录，替换成你自己的文件目录即可
```shell
docker run -d --restart always \
-p 8090:8090 \
--name mygeektime \
-v config.yml:/config.yml \
-v ${directory}:/repo  \
zkep/mygeektime:latest server   
```
服务启动后浏览器访问:  http://127.0.0.1:8090

## [二进制包安装](https://github.com/zkep/mygeektime/releases)
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