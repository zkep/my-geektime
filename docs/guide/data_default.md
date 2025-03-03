# 使用默认数据库资源

**默认数据库，缓存了330+的VIP体系课程，可以直接使用my-geektime进行在线观看，音视频缓存**

微信赞赏并留言 <b>邮箱账号</b>，将回赠<b> 数据库 </b>，无需VIP，即刻畅享VIP课程

<picture>
  <img
    alt="sponsor"
    src="../../images/sponsor.jpg"
    width="256px"
  />
</picture>

#### 赞赏后邮箱收到的附件目录
```shell
my-geektime
├── docker  # docker 目录
│   ├── docker-compose.yml  # docker-compose 配置文件
│   └── mysql  # mysql 目录
│       └── init    # 初始化数据目录
│           ├── article_comment_discussions.sql # 文章评论讨论表
│           ├── article_comments.sql # 文章评论表
│           ├── init.sql  # 创建数据库
│           └── tasks.sql # 课程任务表
└── README.md    # 操作指南
```


###  默认数据库表导入

* 方式1:  docker compose 在启动前，将tasks.sql 放入 docker/mysql/init 目录即可

* 方式2:  cp 到 docker mysql镜像中
```shell
docker cp tasks.sql mysql:/

docker exec -it mysql bash

mysql -uroot -p123456

use mygeektime;

source tasks.sql;
```

* 方式3: mysqldump 将tasks.sql导入到mysql
```shell
mysqldump -uroot -P3306 -p123456 mygeektime < tasks.sql
```

step: 1
```shell
git clone https://github.com/zkep/my-geektime

# 进入docker-compose目录
cd docker

# 如果 docker/mysql/data 存在，需要删除，就可以完全使用赞赏后的数据库

```
step2: docker-compose.yml 添加mysql默认数据表
```yaml
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
      - ./mysql/init/article_comments.sql:/docker-entrypoint-initdb.d/article_comments.sql
      - ./mysql/init/article_comment_discussions.sql:/docker-entrypoint-initdb.d/
```
step3: 启动
```shell
docker-compose up  -d
```



