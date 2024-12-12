 [English](./README.md) | 中文
 

### 我的极客时间
极客时间资源下载工具

---
#### 安装

```shell
go install github.com/zkep/mygeektime@latest
```

#### 查看帮助
```shell
mygeektime -help
```

#### 依赖项

[FFmpeg](https://ffmpeg.org/download.html)

MacOS
```shell
brew install ffmpeg        
```
Linux
```shell
git clone https://github.com/FFmpeg/FFmpeg.git ffmpeg

cd ffmpeg

./configure --enable-ffplay --enable-ffserver

make && make install
```

[ChromeDriver](https://developer.chrome.google.cn/docs/chromedriver/get-started)
* linux64：https://storage.googleapis.com/chrome-for-testing-public/${version}/linux64/chromedriver-linux64.zip
* mac-arm64：https://storage.googleapis.com/chrome-for-testing-public/${version}/mac-arm64/chromedriver-mac-arm64.zip
* mac-x64：https://storage.googleapis.com/chrome-for-testing-public/${version}/mac-x64/chromedriver-mac-x64.zip
* win32：https://storage.googleapis.com/chrome-for-testing-public/${version}/win32/chromedriver-win32.zip
* win64：https://storage.googleapis.com/chrome-for-testing-public/${version}/win64/chromedriver-win64.zip

查看 Chrome 版本号, 在谷歌浏览器的地址栏输入：  
```shell
chrome://version
```
获取到chrome版本号后，可以根据自己的系统替换上面的链接里的 ${version} 下载对应系统的版本
解压后将 chromedriver 文件放在程序执行目录


#### 查看帮助输出
```shell
My GeekTime CLI 0.0.1

Available commands:

   server   This is http server 
   cli      This is Command 

Flags:

  -help
        Get help on the 'mygeektime' command.

```

#### web方式

```shell
mygeektime server
```

#### 感谢
* [gin](https://github.com/gin-gonic/gin)
* [amis](https://github.com/baidu/amis)
* [tebeka/selenium](https://github.com/tebeka/selenium)
* [gorm](https://github.com/go-gorm/gorm)
* [FFmpeg](https://ffmpeg.org/download.html) 
* [ChromeDriver](https://mirrors.huaweicloud.com/chromedriver/)