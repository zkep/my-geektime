# 快速开始 ⚡️

## 安装
* 目前支持三种安装方式, 首推docker compose 方式
* 如果，你的本地已经有redis，mysql等服务了，也可以考虑docker方式和二进制发行包方式
* 再如果，你也是技术爱好者，恰好懂golang和amis的话，也可以clone源码安装调试

## docker compose 安装
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

### docker-compose.yml

```yaml
name: mygeektime
networks:
  mygeektime:
    driver: bridge
services:
  mysql:
    image: mysql:latest
    hostname: "mysql"
    restart: always
    networks:
      - mygeektime
    environment:
      - MYSQL_ROOT_PASSWORD=123456
      - MYSQL_DATABASE=mygeektime
    volumes:
      - ./mysql/data:/var/lib/mysql
      - ./mysql/init/init.sql:/docker-entrypoint-initdb.d/init.sql
      - ./mysql/init/tasks.sql:/docker-entrypoint-initdb.d/tasks.sql
      - /etc/localtime:/etc/localtime:ro
    ports:
      - 33060:3306
  redis:
    image: redis:latest
    hostname: redis
    restart: always
    networks:
      - mygeektime
    volumes:
      - ./redis/data:/data
    command: redis-server --requirepass 123456
    ports:
      - 63790:6379
#  natter:
#    image: natter:latest
#    hostname: natter
#    restart: always
#    build: #启动服务时，先将build中指定的dockerfile打包成镜像，再运行该镜像
#      context: ./natter #指定上下文目录dockerfile所在目录[相对、绝对路径都可以]
#      dockerfile: Dockerfile.debian-amd64 #文件名称[在指定的context的目录下指定那个Dockerfile文件名称]
#    network_mode: host
#    volumes:
#      - ./natter:/data
#    command: -p 80
  server:
    hostname: mygeektime
    image: zkep/mygeektime:latest
#    image:  mygeektime:latest
#    build: #启动服务时，先将build中指定的dockerfile打包成镜像，再运行该镜像
#      context: ../ #指定上下文目录dockerfile所在目录[相对、绝对路径都可以]
#      dockerfile: Dockerfile #文件名称[在指定的context的目录下指定那个Dockerfile文件名称]
    privileged: true
    restart: always
    networks:
      - mygeektime
    command: server --config=config.yml
    ports:
      - 8090:8090
    environment:
      - GIN_MODE=test
    volumes:
      -  ./server/repo:/repo
      -  ./server/config.yml:/config.yml
      -  ./server/wechat.jpg:/wechat.jpg
    depends_on:
      - mysql
      - redis
  nginx:
    image: nginx:latest
    hostname: nginx
    restart: always
    networks:
      - mygeektime
    volumes:
      - ./nginx/nginx.conf:/etc/nginx/nginx.conf
      - ./nginx/html:/usr/share/nginx/html
      - ./nginx/conf.d:/etc/nginx/conf.d
      - ./nginx/logs:/var/logs/nginx
    ports:
      - 80:80
      - 443:443
    depends_on:
      - server
#      - natter

```


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

[配置项](./config.md) -->