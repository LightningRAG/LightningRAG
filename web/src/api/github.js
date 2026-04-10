import axios from 'axios'

const service = axios.create()

export function Commits(page) {
  return service({
    url:
      'https://api.github.com/repos/LightningRAG/LightningRAG/commits?page=' +
      page,
    method: 'get'
  })
}

export function Members() {
  return service({
    url: 'https://api.github.com/orgs/LightningRAG/members',
    method: 'get'
  })
}
