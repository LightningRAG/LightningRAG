export function AddSecret(secret) {
  if (!secret) {
    secret = ''
  }
  // 合肥云亿连
  global['lrag-secret'] = secret
}
