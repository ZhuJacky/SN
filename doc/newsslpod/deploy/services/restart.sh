#!/bin/sh

if [ "$USER" != "user_00" ]; then
  echo "Sorry,Only User User_00 can exec this shell"
  exit 1
fi

chmod +x /data/release/sslpod/backend \
  /data/release/sslpod/checker \
  /data/release/sslpod/notifier

restart(){
  echo "正在停止服务，2s..."
  # backend
  pid=`pidof backend`
  if [ $pid ]; then
    kill -9 $pid
    if [ $? -eq 0 ]; then
      echo "backend 停止成功！"
    else
      echo "backend 停止失败，请检查！"
    fi
  fi
  # checker
  pid=`pidof checker`
  if [ $pid ]; then
    kill -9 $pid
    if [ $? -eq 0 ]; then
      echo "checker 停止成功！"
    else
      echo "checker 停止失败，请检查！"
    fi
  fi
  # notifier
  pid=`pidof notifier`
  if [ $pid ]; then
    kill -9 $pid
    if [ $? -eq 0 ]; then
      echo "notifier 停止成功！"
    else
      echo "notifier 停止失败，请检查！"
    fi
  fi

  echo "正在启动服务..."
  nohup /data/release/sslpod/backend >backend2.log &
  sleep 2
  if pidof backend >/dev/null 2>&1; then
    echo "backend 启动成功！"
  else
    echo "backend 启动失败，查看 backend2.log！"
    exit 1
  fi
  nohup /data/release/sslpod/checker >checker2.log &
  sleep 2
  if pidof checker >/dev/null 2>&1; then
    echo "checker 启动成功！"
  else
    echo "checker 启动失败，查看 checker2.log！"
    exit 1
  fi
  nohup /data/release/sslpod/notifier >notifier2.log &
  sleep 2
  if pidof notifier >/dev/null 2>&1; then
    echo "notifier 启动成功！"
  else
    echo "notifier 启动失败，查看 notifier2.log！"
    exit 1
  fi
}

restart "$@"
