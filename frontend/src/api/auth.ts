import api from './index'

export interface LoginRequest {
  username: string
  password: string
}

export interface LoginResponse {
  token: string
  username: string
}

export const authApi = {
  login(data: LoginRequest) {
    return api.post<LoginResponse>('/auth/login', data)
  },
  logout() {
    return api.post('/auth/logout')
  },
  changePassword(oldPassword: string, newPassword: string) {
    return api.put('/auth/password', {
      old_password: oldPassword,
      new_password: newPassword,
    })
  },
}
