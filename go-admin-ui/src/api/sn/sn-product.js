import request from '@/utils/request'

// 查询产品列表
export function listProduct(query) {
  return request({
    url: '/api/v1/sn-product',
    method: 'get',
    params: query
  })
}