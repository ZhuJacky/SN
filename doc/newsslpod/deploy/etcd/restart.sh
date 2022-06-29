#!/bin/sh

chmod +x etcd

restart(){
  echo "正在停止服务，2s..."
  # backend
  pid=`pidof etcd`
  if [ $pid ]; then
    kill -9 $pid
    if [ $? -eq 0 ]; then
      echo "etcd 停止成功！"
    else
      echo "etcd 停止失败，请检查！"
    fi
  fi

  echo "正在启动服务..."
  nohup ./etcd --data-dir /etcd-data \
    --name node1 \
    --listen-peer-urls http://0.0.0.0:2380 \
    --listen-client-urls http://0.0.0.0:2379 \
    --advertise-client-urls http://0.0.0.0:2379 \
    --initial-advertise-peer-urls http://0.0.0.0:2380 \
    --initial-cluster node1=http://0.0.0.0:2380 > etcd.log 2>&1 &
  sleep 2
  if pidof etcd >/dev/null 2>&1; then
    echo "etcd 启动成功！"
  else
    echo "etcd 启动失败，查看 etcd.log！"
    exit 1
  fi
}

restart "$@"
