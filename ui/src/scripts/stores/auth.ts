import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import api from '../api'

export const useAuthStore = defineStore('auth', () => {
  const token = ref(localStorage.getItem('token') || '')
  const role = ref(localStorage.getItem('role') || '')
  const account = ref(localStorage.getItem('account') || '')
  const userId = ref(localStorage.getItem('userId') || '')
  const balance = ref(Number(localStorage.getItem('balance')) || 0)

  const isLoggedIn = computed(() => !!token.value)
  const isAdmin = computed(() => role.value === 'admin')

  async function login(username: string, password: string) {
    const { data } = await api.post('/v0/auth/login', { account: username, password })
    token.value = data.token
    role.value = data.role
    account.value = data.account
    userId.value = data.userId
    balance.value = data.balance ?? 0
    localStorage.setItem('token', data.token)
    localStorage.setItem('role', data.role)
    localStorage.setItem('account', data.account)
    localStorage.setItem('userId', data.userId)
    localStorage.setItem('balance', String(data.balance ?? 0))
  }

  function logout() {
    token.value = ''
    role.value = ''
    account.value = ''
    userId.value = ''
    balance.value = 0
    localStorage.removeItem('token')
    localStorage.removeItem('role')
    localStorage.removeItem('account')
    localStorage.removeItem('userId')
    localStorage.removeItem('balance')
  }

  function updateBalance(newBalance: number) {
    balance.value = newBalance
    localStorage.setItem('balance', String(newBalance))
  }

  async function fetchBalance() {
    try {
      const { data } = await api.get('/v0/user/balance')
      balance.value = data.balance ?? 0
      localStorage.setItem('balance', String(data.balance ?? 0))
    } catch {
      // keep current balance on error
    }
  }

  async function changePassword(oldPassword: string, newPassword: string) {
    await api.post('/v0/user/change-password', { oldPassword, newPassword })
  }

  return { token, role, account, userId, balance, isLoggedIn, isAdmin, login, logout, updateBalance, fetchBalance, changePassword }
})
