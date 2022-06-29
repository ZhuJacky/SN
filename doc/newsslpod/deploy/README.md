# SSLPod 部署文档

服务部署文档步骤：

1. 准备基础服务
2. 准备数据库、redis
3. 修改services配置配置
4. 启动services服务



### 1. 准备基础服务 Node1

```
机器要求：4c/8g centos7
环境要求：无
外网需求：无
运行服务：echoip、etcd
```

**echoip**：执行 `./restart.sh`。验证启动`lsof -i :8089`。

**etcd**：执行 `./restart.sh`。验证启动 `lsof -i :2379`。

### 2. 准备数据库、redis

```
环境要求：MySQL: 5.7
 1. 创建数据库 sslpod
 2. 初始化数据库：执行 brand_sql.zip 内的 sql 脚本。

Redis: none
```



### 3. 更新services配置

根据运行环境来更改配置文件：

```
DEV -> dev.yml
TEST -> test.yml
PRE_PRODUCTION/PRODUCTION -> prod.yml
```

更改配置文件（这里用层级的方式来表示）:

```
1. 更新数据库地址
database.source: "<user>:<password>(<ip>:<port>)/sslpod?charset=utf8mb4&parseTime=True&loc=Local"

2. 更新redis地址与Auth
redis.address: <ip>:6379
redis.password: <auth>

3. 更新echo配置
etcd.endpoints:
    - http://<node1 ip>:2379
etcd.echourl: http://<node1 ip>:8089/latest/meta-data

4. 更新myssl配置
myssl.domain: myssl.com
myssl.id: sslpod
myssl.key: 5bc0c96c57186d42cdfd13709b401a2b3b7e5ea4
myssl.anaapi: https://myssl.com/eeapi/v1/analyze
myssl.detectioncount: 40
```



### 4. 启动services服务

```
机器要求：4c/8g centos7
外网需求：有
运行服务：backend、checker、notifier
```

根据运行环境来操作环境变量：

```
DEV -> export RUN_MODE=DEV
TEST -> export RUN_MODE=TEST
PRE_PRODUCTION/PRODUCTION -> export RUN_MODE=PRODUCTION
```

执行 `./restart.sh` 启动。验证启动1：脚本日志是否打印《启动成功》，验证启动2: `lsof -i :20000`、`lsof -i :20010`、`lsof -i :20020`
