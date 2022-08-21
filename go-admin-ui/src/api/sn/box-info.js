import request from '@/utils/request'

// 查询列表
export function listPost(query) {
  return request({
    url: '/api/v1/box-info',
    method: 'get',
    params: query
  })
}
// 修改状态
export function updatePost(data, id) {
  return request({
    url: '/api/v1/box-info/' + id,
    method: 'put',
    data: data
  })
}

// 查询出库箱号列表
export function listExWarehouseBox(query) {
  return request({
    url: '/api/v1/ex-warehouse/ex-warehouse-box',
    method: 'get',
    params: query
  })
}

//执行出库
export function doExWarehouseBox(data, id) {
  return request({
    url: '/api/v1/ex-warehouse/do-ex-warehouse',
    method: 'POST',
    data: data
  })
}
