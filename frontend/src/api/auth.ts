import api from './index'

export interface LoginRequest {
  username: string
  password: string
}

export interface LoginResponse {
  token: string
  username: string
}

export interface ChangePasswordRequest {
  old_password: string
  new_password: string
}

export const authApi = {
  login(data: LoginRequest) {
    return api.post<LoginResponse>('/auth/login', data)
  },
  logout() {
    return api.post('/auth/logout')
  },
  changePassword(data: ChangePasswordRequest) {
    return api.put('/auth/password', data)
  },
}
