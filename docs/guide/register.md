# 注册账号

默认**用户名密码**注册，第一个注册的用户默认是**管理员权限**

接下来注册的用户是**普通用户权限**

### 管理员权限：

  * 我的课程列表，缓存极客时间的课程
  * 极客时间课程，挑选课程进行缓存
  * 用户管理，开放给其他用户登录系统

<img src="../../images/admin_home.png" />

### 普通用户权限：

 * 我的课程列表

<img src="../../images/user_home.png" />

### 访客模式
填写guest访客账户名密码，即可开启访客模式，根据注册方式，配置登录的方式

guest访客账户名密码需要真实存在于系统中，也就是可以注册一个非管理账号填写guest配置信息

```yaml
site:
  login:
    type: name # name
    guest:
      name: 
      password: 
```
### 关闭注册
修改配置文件中，site.register.type 为 **none** 既可关闭注册页面
```yaml
site:
  download: true
  register:
    type: none #  name | none
```
<img src="../../images/only_login.png" />