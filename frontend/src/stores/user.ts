import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { authApi, type LoginRequest } from '@/api/auth'

export const useUserStore = defineStore('user', () => {
  const token = ref(localStorage.getItem('token') || '')
  const username = ref(localStorage.getItem('username') || '')

  const isLoggedIn = computed(() => !!token.value)

  async function login(data: LoginRequest) {
    const response = await authApi.login(data)
    token.value = response.data.token
    username.value = response.data.username
    localStorage.setItem('token', response.data.token)
    localStorage.setItem('username', response.data.username)
  }

  async function logout() {
    try {
      await authApi.logout()
    } finally {
      token.value = ''
      username.value = ''
      localStorage.removeItem('token')
      localStorage.removeItem('username')
    }
  }

  return {
    token,
    username,
    isLoggedIn,
    login,
    logout,
  }
})
