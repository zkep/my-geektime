English | [中文](./README.md)

### my geektime
This is a tool to obtain the geektime  articles docs

---

### [Docs](https://zkep.github.io/mygeektime/) | [Show Time](https://mygeektime.anyfun.tech)


#### Install

#### install with docker compose

```shell
git clone https://github.com/zkep/mygeektime.git

cd mygeektime/docker

# Apple M1 , M2 modify docker-compose.yml 35 line use zkep/mygeektime:mac-m

# download default data
wget https://github.com/zkep/mygeektime/releases/download/v0.0.1/tasks.sql -O mysql/init/tasks.sql

docker-compose up -d
```

browser web url:  http://127.0.0.1:8090


#### thanks
* [gin](https://github.com/gin-gonic/gin)
* [amis](https://github.com/baidu/amis)
* [FFmpeg](https://ffmpeg.org/download.html)



#### WeChat Sponsor

<picture>
  <img
    alt="sponsor"
    src="docs/images/sponsor.jpg"
    width="256px"
  />
</picture>


