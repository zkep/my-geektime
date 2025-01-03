# 使用默认数据库资源

**默认数据库，缓存了一些VIP课程，可以直接使用mygeektime进行在线观看**


## [releases下载sql文件导入mysql](https://github.com/zkep/mygeektime/releases/v0.0.1)


* 方式1: docker compose 在启动前，将tasks.sql 放入 docker/mysql/init 目录即可
```shell
git clone https://github.com/zkep/mygeektime

wget https://github.com/zkep/mygeektime/releases/download/v0.0.1/tasks.sql -O docker/mysql/init/tasks.sql

# 进入docker-compose目录
cd docker
# 启动
docker-compose up  -d
```
* 方式2: 将tasks.sql导入到mysql
```shell
mysqldump -uroot -P3306 -p123456 mygeektime < tasks.sql
```

