import service from '@/utils/request'

export const oauthGlobalGet = () => {
  return service({
    url: '/sysOAuthSetting/getOAuthSetting',
    method: 'get'
  })
}

export const oauthGlobalUpdate = (data) => {
  return service({
    url: '/sysOAuthSetting/updateOAuthSetting',
    method: 'put',
    data
  })
}

export const oauthRegisteredKinds = () => {
  return service({
    url: '/sysOAuthProvider/getRegisteredOAuthKinds',
    method: 'get'
  })
}

export const oauthProviderList = (params) => {
  return service({
    url: '/sysOAuthProvider/getOAuthProviderList',
    method: 'get',
    params
  })
}

export const oauthProviderFind = (ID) => {
  return service({
    url: '/sysOAuthProvider/findOAuthProvider',
    method: 'get',
    params: { ID }
  })
}

export const oauthProviderCreate = (data) => {
  return service({
    url: '/sysOAuthProvider/createOAuthProvider',
    method: 'post',
    data
  })
}

export const oauthProviderUpdate = (data) => {
  return service({
    url: '/sysOAuthProvider/updateOAuthProvider',
    method: 'put',
    data
  })
}

export const oauthProviderDelete = (ID) => {
  return service({
    url: '/sysOAuthProvider/deleteOAuthProvider',
    method: 'delete',
    params: { ID }
  })
}
