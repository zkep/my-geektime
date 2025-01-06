English | [中文](./README.md)

### my geektime
This is a tool to obtain the geektime  articles docs

---

### [Docs](https://zkep.github.io/mygeektime/)


### [Show Time](https://mygeektime.anyfun.tech)


#### Install

#### install with docker compose

```shell
git clone https://github.com/zkep/mygeektime.git

cd mygeektime/docker

# download default data
wget https://github.com/zkep/mygeektime/releases/download/v0.0.1/tasks.sql -O mysql/init/tasks.sql

docker-compose up -d
```

browser web url:  http://127.0.0.1:8090


#### thanks
* [gin](https://github.com/gin-gonic/gin)
* [amis](https://github.com/baidu/amis)
* [gorm](https://github.com/go-gorm/gorm)
* [FFmpeg](https://ffmpeg.org/download.html)

## Star History

<picture>
  <source
    media="(prefers-color-scheme: dark)"
    srcset="
      https://api.star-history.com/svg?repos=zkep/mygeektime&type=Date&theme=dark
    "
  />
  <source
    media="(prefers-color-scheme: light)"
    srcset="
      https://api.star-history.com/svg?repos=zkep/mygeektime&type=Date
    "
  />
  <img
    alt="Star History Chart"
    src="https://api.star-history.com/svg?repos=zkep/mygeektime&type=Date"
  />
</picture>