三个服务均会使用 `conf` 下的配置文件.如需要单独部署，请拷贝二进制及 `conf` 文件夹。



启动服务之前，请设置机器的运行模式：

```
# DEV 会使用配置文件 dev.yml
export RUN_MODE=DEV

# TEST 会使用配置文件 test.yml
export RUN_MODE=TEST

# PRE_PRODUCTION 会使用配置文件 prod.yml
export RUN_MODE=PRE_PRODUCTION

# PRODUCTION 会使用配置文件 prod.yml
export RUN_MODE=PRODUCTION
```

项目绝对路径为：`/data/release/sslpod/`
