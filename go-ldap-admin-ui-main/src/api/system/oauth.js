import request from '@/utils/request'

// 登录页连接器列表
export function publicConnectors() {
  return request({
    url: '/api/oauth/connectors/public',
    method: 'get'
  })
}

// 管理端连接器列表
export function listConnectors(params) {
  return request({
    url: '/api/oauth/connectors',
    method: 'get',
    params
  })
}

export function createConnector(data) {
  return request({
    url: '/api/oauth/connectors',
    method: 'post',
    data
  })
}

export function updateConnector(data) {
  return request({
    url: '/api/oauth/connectors/update',
    method: 'post',
    data
  })
}

export function deleteConnector(data) {
  return request({
    url: '/api/oauth/connectors/delete',
    method: 'post',
    data
  })
}

// 启动登录授权
export function startOAuth(data) {
  return request({
    url: '/api/oauth/start',
    method: 'post',
    data
  })
}

// 启动密码校验授权
export function startPasswordOAuth(data) {
  return request({
    url: '/api/oauth/start/password',
    method: 'post',
    data
  })
}
