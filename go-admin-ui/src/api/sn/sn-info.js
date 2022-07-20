import request from '@/utils/request'

// 查询岗位列表
export function listPost(query) {
  return request({
    url: '/api/v1/sn-info',
    method: 'get',
    params: query
  })
}


