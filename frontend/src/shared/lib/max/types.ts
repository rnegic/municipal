export interface MaxWebAppUser {
  id: number
  first_name?: string
  last_name?: string
  username?: string
  photo_url?: string
}

export interface MaxWebAppInitData {
  user?: MaxWebAppUser
  start_param?: string
  auth_date?: number
  hash?: string
}

export interface MaxWebAppApi {
  initData: string
  initDataUnsafe?: MaxWebAppInitData
  platform?: string
  version?: string
  ready?: () => void
  expand?: () => void
}
