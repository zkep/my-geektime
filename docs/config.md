# 配置项 ⚙️

默认配置项内容如下： 
```yaml
server:
  app_name: My Geek Time  # 服务名
  run_mode: debug
  http_addr: 0.0.0.0      # ip
  http_port: 8090         # http 端口
jwt:                      # jwt 权限配置
  secret: my-geektime-secret 
  expires: 86400
i18n:                     # 国际化配置
  directory: i18n
  default_lang: zh-CN
database:                 # 数据库配置，默认是sqlite，可以自定义为mysql，postgres
  driver:  mysql         # mysql|postgres|sqlite
  # source:  mygeektime.db 
  source: root:123456@tcp(127.0.0.1:3306)/mygeektime?charset=utf8&parseTime=True&loc=Local&timeout=1000ms
  # source: host=127.0.0.1 user=postgres password=postgres dbname=mygeektime port=5432 sslmode=disable TimeZone=Asia/Shanghai
  max_idle_conns: 10
  max_open_conns: 100
storage:                  # 音视频资源下载目录，
  driver: local           # 目前仅支持存在本地，但是留了扩展，后面可以支持多种存储方式
  directory: repo         # 本地目录
  bucket: object          # 访问链接前缀，没有特殊需求，可以不用修改
  host: http://127.0.0.1:8090  # 如果是本地服务，端口需要和上面的http_port保持一致，如果配置了域名请换成自己的域名
browser:
  open_browser: true       # 默认启动会自动打开浏览器访问，docker部署无视改参数
site:                      # 站点配置
  download: true           # 是否下载音视频，默认是
  login:                   # 登录配置，默认用户名登录，与注册方式相同
    type: name             # name
    guest:                 # 是否开启访客模式，填写默认name，passwrod视为开启，同时数据库users表应该有该记录
      name:                # 登录名
      password:            # 密码
  play:                    # 播放配置
    type: origin  #  origin | local
    # 使用源站播放，如果site.download 设置为false，默认是不会下载音视频（如果你的磁盘有限），播放时会直接用极客时间的播放链接
    # 如果发现播放的时候没有下载权限，请配置proxy_url，则会重写header的orgin代理下载分片
    proxy_url:  
      - https://res001.geekbang.org
  proxy:   # 使用服务端代理，解决geektime源图片不显示配置
    cache: true
    proxy_url: http://127.0.0.1:8090/v2/file/proxy?url={url} 
    urls: # 需要被代理请求的极客时间链接前缀
      - https://static001.geekbang.org/resource/image
      - https://static001.geekbang.org/account/avatar

```
