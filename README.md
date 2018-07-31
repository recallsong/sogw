# Sogw
闲着也是闲着，于是就写个 Api Gateway  -_-!

Sogw是一个Golang实现的 Api Gateway 服务程序，轻量级，高性能。具有如下特色：

* 基于fasthttp实现，且支持同时监听https和http
* 支持设置Header、Cookie等
* 支持URL Rewrite
* 支持通配符路由、支持路径变量
* 支持数据校验
* 参考echo实现高效路由、且支持API条件匹配路由
* 支持域名路由、支持域名虚拟主机
* 支持负载均衡，有RoundRobin 和 IPHash等策略
* 可以从文件中读取路由、Api、服务等信息
* 可以通过etcd等K/V存储实现服务发现
* 支持路由、API等信息热更新，无需重启
* 支持Http Basic等认证方式
* 易于扩展Filter、易于扩展后台任务Job
* 支持从Swagger文件导入Api等信息
* 提供Restful接口管理 Api Gateway

本项目还在开发完善中...

# Topology

![Topology](https://github.com/recallsong/sogw/raw/master/docs/img/topology.png)

# Download

    go get github.com/recallsong/sogw

# Config
Edit store url in sogw/conf/sogw.yml
## Run With File Store

    store:
        url: "file://./conf/meta.yml"
        watch: true

[Store File Example](https://xxx)

## Run With Etcd Store

    store:
       url: "etcd://localhost:2379/test"
       watch: true
     
# Run

    cd sogw
    make run

## Build To Docker Image
    
    make docker-build

## Run In Docker Container

    make docker-run

# Publish Apis

    cd sogw-ctl
    // edit sogw-ctl.yml and swagger file
    sogw-ctl pub
   
## Show Information From Store
  
    sogw-ctl show

# Run Backend Server
  
    cd sogw-backend
    go run main.go --addr=7001

# Request Test 

![Request Test](https://github.com/recallsong/sogw/raw/master/docs/img/example.png)

# TODO List
* 以后台任务方式实现Health Check
* 添加监控统计
* 添加OAuth认证
* 添加Lambda表达式
* 规范代码、添加注释、添加测试
* 全面测试

# License
[MIT](https://github.com/recallsong/sogw/blob/master/LICENSE)
