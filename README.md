# jt809_server
   go部标809下级平台,目前支持2011,2013,2019版本的协议,目前消息来源是对接我们自己的业务系统,大家可以对接自己的业务系统,只要实现internal下的bu_service,后续抽象成接口
## 功能特性
- [X] 链路管理
- - [X] 主从链路登录
- - [X] 主从链路心跳
- [X] 车辆注册
- [X] 定位数据
- [X] 自动补报
- [X] 上报报警
- [X] 上报驾驶员信息
- [X] 报警督办
- [X] 车辆抓拍
- [X] 静态数据
- [X] 报文下发
- [X] 支持多用户

## 快速开始
> 这里假设用户已经安装好go,git 环境,如果没有安装好,请参考其它教程,go要求版本 >=1.19

1. 下载代码

    git clone https://github.com/Yordroid/jt809_server.git
2. 进入jt809_server 目录

   go mod tidy
3. 执行服务

   go run ./

 

## 目录说明
- apis  http请求接口
- config 应用的配置
- internal 内部服务处理
   - bu_service 业务服务消息管理服务
   - data_manage_service 数据管理
   - jt_service JT809部标服务管理
      - session_manage JT809链路管理
- models 数据结构定义
- routers http 路由
- util 基础工具
## 技术讨论
欢迎大家加入一起完善 QQ群:255797894