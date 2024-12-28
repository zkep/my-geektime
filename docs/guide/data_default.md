# 使用默认数据库资源

**默认数据库，缓存了一些VIP课程，可以直接使用mygeektime进行在线观看**


## 方式一: [releases下载sql文件导入mysql](https://github.com/zkep/mygeektime/releases)


## 方式二: git-lfs下载默认数据库表导入mysql

### 安装 git-lfs

```shell
sudo apt-get install git-lfs
```

### 拉取仓库大文件
```shell
git clone https://github.com/zkep/mygeektime.git

cd mygeektime

git lfs pull
```

## 将tasks.sql导入到mysql

```shell
mysqldump -uroot -P3306 -p123456 mygeektime < docker/mysql/init/tasks.sql
```
