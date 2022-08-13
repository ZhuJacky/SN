import request from '@/utils/request'

// 查询SN列表
export function listPost(query) {
  return request({
    url: '/api/v1/box-relation-info',
    method: 'get',
    params: query
  })
}
