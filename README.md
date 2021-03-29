# 整合xray和fofa的自动挖掘机:Snowball

经常采用xray开监听，然后在浏览器中设置代理来对目标扫描，感觉比较麻烦。最近在学习golang，所以尝试用golang来写这个手工处理过程。
golang给人的感觉很强大，其实就是强大，目前还没有学到家，写的代码非常之不成熟，现在能实现简单功能，以后再一边学语言一边完善。

## 原来的功能需求

- 交互式命令行操作

- 支持多平台搜索

- 跨平台运行

- 扫描结果可视会

现在只实现了基本的命令行交互操作，后续还有很多功能待完善。

## 使用

[!help](https://github.com/rootklt/snowball/blob/main/images/help.png)

1. 如果要使用xray扫描漏洞，则进入交互后启动xray代理

```bash
    snowball>>xray start
```

[!start xray](https://github.com/rootklt/snowball/blob/main/images/start_xray.png)

[!stop_xray](https://github.com/rootklt/snowball/blob/main/images/xray_stop.png)

只是查询返回url=>title。

2. 搜索命令

```bash
    snowball>>search fofa -h    (查看帮助)
```
[!fofa](https://github.com/rootklt/snowball/blob/main/images/fofa.png)

开启了xray代理扫描后会将搜索到的目标去重、通过代理访问目标，流量交给xray处理

3. xray使用--webhook-output模式，因此在进入命令行交互模式时已经开启webhook服务，将会当前执行状态和漏洞结果显示在标准输出上，看起来有些碍事。

4. 扫描到漏洞的话，一是将日志信息写到console，同时也写入日志文件（文件位置可以在config/logger.json配置），方便查询

   [!scan](https://github.com/rootklt/snowball/blob/main/images/scanning.png)

## 声明

仅用于学习目的，学习之外的不由本项目负责。