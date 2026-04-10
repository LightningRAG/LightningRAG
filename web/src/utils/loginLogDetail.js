/**
 * 登录日志 error_message：库中可能为服务端 i18n 键、或历史各语言句子。
 * 展示时统一走前端 vue-i18n（英语文案以 en 为准）。
 */

const SERVER_KEY_TO_I18N = {
  'sys.login.captcha_wrong': 'tools.loginLog.detailCaptchaWrong',
  'sys.login.bad_credentials': 'tools.loginLog.detailBadCredentials',
  'sys.login.user_disabled': 'tools.loginLog.detailUserDisabled',
  'sys.login.token_failed': 'tools.loginLog.detailTokenFailed',
  'sys.login.session_setup_failed': 'tools.loginLog.detailSessionFailed',
  'sys.login.jwt_revoke_failed': 'tools.loginLog.detailJwtRevokeFailed',
  'common.login_success': 'tools.loginLog.loginOkDetail'
}

/** 历史记录：各语言已写入库的整句 → 同上 i18n 键 */
const LEGACY_TEXT_TO_I18N = {
  // sys.login.captcha_wrong
  验证码错误: 'tools.loginLog.detailCaptchaWrong',
  驗證碼錯誤: 'tools.loginLog.detailCaptchaWrong',
  'Invalid captcha': 'tools.loginLog.detailCaptchaWrong',
  // sys.login.bad_credentials
  用户名不存在或者密码错误: 'tools.loginLog.detailBadCredentials',
  使用者名稱不存在或密碼錯誤: 'tools.loginLog.detailBadCredentials',
  'Invalid username or password': 'tools.loginLog.detailBadCredentials',
  密码错误: 'tools.loginLog.detailBadCredentials',
  // sys.login.user_disabled
  用户被禁止登录: 'tools.loginLog.detailUserDisabled',
  此帳戶已被禁止登入: 'tools.loginLog.detailUserDisabled',
  'This account is disabled': 'tools.loginLog.detailUserDisabled',
  // sys.login.token_failed
  获取token失败: 'tools.loginLog.detailTokenFailed',
  取得權杖失敗: 'tools.loginLog.detailTokenFailed',
  'Failed to issue token': 'tools.loginLog.detailTokenFailed',
  // sys.login.session_setup_failed
  设置登录状态失败: 'tools.loginLog.detailSessionFailed',
  設定登入狀態失敗: 'tools.loginLog.detailSessionFailed',
  'Failed to set login session': 'tools.loginLog.detailSessionFailed',
  // sys.login.jwt_revoke_failed
  jwt作废失败: 'tools.loginLog.detailJwtRevokeFailed',
  '作廢 JWT 失敗': 'tools.loginLog.detailJwtRevokeFailed',
  'Failed to revoke the previous token': 'tools.loginLog.detailJwtRevokeFailed',
  // common.login_success
  登录成功: 'tools.loginLog.loginOkDetail',
  登入成功: 'tools.loginLog.loginOkDetail',
  'Signed in successfully': 'tools.loginLog.loginOkDetail'
}

function translateByPath(t, path) {
  const out = t(path)
  return out === path ? '' : out
}

/**
 * @param {string|null|undefined} raw 库里的 error_message
 * @param {(key: string) => string} t vue-i18n t
 * @returns {string}
 */
export function resolveLoginLogDetailMessage(raw, t) {
  if (raw == null) return ''
  const s = String(raw).trim()
  if (!s) return ''

  const path = SERVER_KEY_TO_I18N[s] || LEGACY_TEXT_TO_I18N[s]
  if (path) return translateByPath(t, path)

  return s
}

/**
 * @param {boolean} statusOk 是否登录成功
 */
export function formatLoginLogDetailCell(statusOk, errorMessage, t) {
  if (statusOk) {
    const r = resolveLoginLogDetailMessage(errorMessage, t)
    return r || t('tools.loginLog.loginOkDetail')
  }
  const r = resolveLoginLogDetailMessage(errorMessage, t)
  return r || errorMessage || t('common.loginStatusFail')
}
