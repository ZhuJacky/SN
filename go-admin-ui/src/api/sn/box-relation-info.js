import request from '@/utils/request'

// 查询SN列表
export function listPost(query) {
  return request({
    url: '/api/v1/box-relation-info',
    method: 'get',
    params: query
  })
}

//执行装箱操作
export function packBox(data, id) {

  //alert(id)
  return request({
    url: '/api/v1/box-relation-info/' + id,
    method: 'put',
    data: data
  })
}
