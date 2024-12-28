# 使用默认数据库资源

**默认数据库，缓存了一些VIP课程，可以直接使用mygeektime进行在线观看**


## [releases下载sql文件导入mysql](https://github.com/zkep/mygeektime/releases)


## 将tasks.sql导入到mysql

```shell
mysqldump -uroot -P3306 -p123456 mygeektime < tasks.sql
```
