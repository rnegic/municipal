export interface MaxWebAppUser {
  id: number
  first_name?: string
  last_name?: string
  username?: string
  photo_url?: string
}

export interface MaxWebAppInitData {
  query_id?: string
  ip?: string
  user?: MaxWebAppUser
  chat?: MaxWebAppChat
  start_param?: string
  auth_date?: number
  hash?: string
}

export interface MaxWebAppChat {
  id: number
  type: 'DIALOG' | 'CHAT' | 'CHANNEL'
}

export interface MaxWebAppShareContentParams {
  text?: string
  link?: string
}

export interface MaxWebAppApi {
  initData: string
  initDataUnsafe?: MaxWebAppInitData
  platform?: string
  version?: string
  ready?: () => void
  expand?: () => void
  openMaxLink?: (url: string) => void
  shareMaxContent?: (params: MaxWebAppShareContentParams) => void | Promise<unknown>
}
