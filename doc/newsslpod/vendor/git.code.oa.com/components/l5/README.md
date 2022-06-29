# L5Api for go

不同于cgo版本，纯Go实现的l5api，基于l5api version:40007


# Api Doc

## 安装

```bash
go get -v -u git.code.oa.com/components/l5
```

## 创建实例(只需创建一次)

```go	
api := l5.NewDefaultApi()
```
or
```go
api := l5.NewApi(&Option{...})
```

## 根据l5名获取sid

```go
domain := api.Query("gol5test")
sid := fmt.Sprintf("%d:%d", domain.Mod, domain.Cmd)
```

## 根据sid获取服务器

```go
srv, err := api.GetServerBySid(mod, cmd)
```

## 根据sid获取全部服务器

*agent返回的全部srv，不确定是否包含权重为0的srv

```go
srvs, err := api.GetServersBySid(mod, cmd)
```

## 根据l5名获取服务器

```go
srv, err := api.GetServerByName("gol5test")
```

## 根据sid获取服务器

```go
srv, err := api.GetServerBySid(mod, cmd)
```

## 上报返回

```go
err = srv.StatUpdate(result, usetime)
```

# Benchmark

```
goos: linux
goarch: amd64
pkg: git.code.oa.com/components/l5
BenchmarkApi_GetServerBySid-6            2000000               675 ns/op
BenchmarkApi_GetServerByName-6           3000000               453 ns/op
BenchmarkServer_StatUpdate-6            20000000                82.1 ns/op
PASS
ok      git.code.oa.com/components/l5   5.709s
```

# Protocol
* QOS_CMD_QUERY_SNAME					根据名称查询mod+cmd
* QOS_CMD_BATCH_GET_ROUTE_WEIGHT			根据mod+cmd批量查询带权重的server
* QOS_CMD_GET_STAT						上报分配统计
* QOS_CMD_CALLER_UPDATE_BIT64			上报调用状态

# Tips

抓包（与agent通信）：

```bash
tcpdump -i lo udp and \(\(dst host 127.0.0.1 and dst port 8888\) or \(src host 127.0.0.1 and src port 8888\)\) -x -nn
```
