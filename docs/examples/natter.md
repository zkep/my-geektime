# 使用[natter](https://github.com/MikeWang000000/Natter)将服务暴露在公网环境


## 快速开始

```bash

git clone https://github.com/zkep/mygeektime.git

cd docker natter 
# 使用内置转发，对外开放本机 8090 端口：

python3 natter.py -p 8090
```

使用 iptables 内核转发（需要 root 权限），对外开放本机 8090 端口：

```bash
sudo python3 natter.py -m iptables -p 8090
```

或者, 使用 Docker:

```bash
docker run --net=host nattertool/natter -p 8090
```

```
2023-11-01 01:00:08 [I] Natter
2023-11-01 01:00:08 [I] Tips: Use `--help` to see help messages
2023-11-01 01:00:12 [I]
2023-11-01 01:00:12 [I] tcp://192.168.1.100:13483 <--Natter--> tcp://203.0.113.10:14500
2023-11-01 01:00:12 [I]
2023-11-01 01:00:12 [I] Test mode in on.
2023-11-01 01:00:12 [I] Please check [ http://203.0.113.10:14500 ]
2023-11-01 01:00:12 [I]
2023-11-01 01:00:12 [I] LAN > 192.168.1.100:13483   [ OPEN ]
2023-11-01 01:00:12 [I] LAN > 192.168.1.100:13483   [ OPEN ]
2023-11-01 01:00:12 [I] LAN > 203.0.113.10:14500    [ OPEN ]
2023-11-01 01:00:13 [I] WAN > 203.0.113.10:14500    [ OPEN ]
2023-11-01 01:00:13 [I]
```

上述例子中, `203.0.113.10` 是您 NAT 1 外部的公网 IP 地址。Natter 打开了 TCP 端口 `203.0.113.10:14500` 以供测试。

在局域网外访问 `http://203.0.113.10:14500` ，您可以看到如下网页:

```
It works!

--------
Natter
```




