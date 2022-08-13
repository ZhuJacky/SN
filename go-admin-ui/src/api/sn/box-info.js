import request from '@/utils/request'

// 查询SN列表
export function listPost(query) {
  return request({
    url: '/api/v1/box-info',
    method: 'get',
    params: query
  })
}
// 修改SN状态
export function updatePost(data, id) {
  return request({
    url: '/api/v1/box-info/' + id,
    method: 'put',
    data: data
  })
}
