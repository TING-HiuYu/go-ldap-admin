import request from '@/utils/request'
// 用户登录（已完成）
export function login(data) {
  return request({
    url: '/api/base/login',
    method: 'post',
    data
  })
}

export function refreshToken() {
  return request({
    url: '/api/base/refreshToken',
    method: 'post'
  })
}
// 用户退出接口（已完成）
export function logout() {
  return request({
    url: '/api/base/logout',
    method: 'post'
  })
}

// 邮箱验证码登录：发送验证码
export function sendLoginOtp(data) {
  return request({
    url: '/api/base/otp/send',
    method: 'post',
    data
  })
}

// 邮箱验证码登录：提交验证码换取 token
export function otpLogin(data) {
  return request({
    url: '/api/base/otp/login',
    method: 'post',
    data
  })
}
// 获取配置信息
export function getConfig() {
  return request({
    url: '/api/base/config',
    method: 'get'
  })
}
// 获取版本信息
export function getVersion() {
  return request({
    url: '/api/base/version',
    method: 'get'
  })
}
