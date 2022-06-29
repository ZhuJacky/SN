#!/bin/sh

chmod +x echoip

restart(){
  echo "正在停止服务，2s..."
  # backend
  pid=`pidof echoip`
  if [ $pid ]; then
    kill -9 $pid
    if [ $? -eq 0 ]; then
      echo "echoip 停止成功！"
    else
      echo "echoip 停止失败，请检查！"
    fi
  fi

  echo "正在启动服务..."
  nohup ./echoip > echoip.log 2>&1 &
  sleep 2
  if pidof echoip >/dev/null 2>&1; then
    echo "echoip 启动成功！"
  else
    echo "echoip 启动失败，查看 echoip.log！"
    exit 1
  fi
}

restart "$@"
