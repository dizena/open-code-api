<template>
  <el-dialog v-model="visible" :title="mode === 'login' ? '登录' : '注册'" width="420px" @close="resetForm">
    <el-form ref="formRef" :model="form" :rules="rules" label-width="0">
      <el-form-item prop="account">
        <el-input v-model="form.account" placeholder="邮箱地址" prefix-icon="Message" />
      </el-form-item>

      <template v-if="step === 'password'">
        <el-form-item prop="password">
          <el-input v-model="form.password" type="password" placeholder="密码" prefix-icon="Lock" show-password />
        </el-form-item>
      </template>

      <template v-if="step === 'code'">
        <div class="flex gap-2">
          <el-form-item prop="code" class="flex-1 mb-0">
            <el-input v-model="form.code" placeholder="8位验证码" maxlength="8" />
          </el-form-item>
          <el-button :disabled="countdown > 0" @click="sendCode">
            {{ countdown > 0 ? `${countdown}s` : '发送验证码' }}
          </el-button>
        </div>
      </template>
    </el-form>

    <template #footer>
      <div class="flex flex-col gap-3 w-full">
        <el-button type="primary" :loading="loading" class="w-full" @click="handleSubmit">
          {{ mode === 'login' ? '登录' : '注册登录' }}
        </el-button>
        <el-button text class="w-full" @click="showForgotDialog = true">忘记密码？</el-button>
      </div>
    </template>

    <ForgotPasswordDialog v-model="showForgotDialog" />
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import api from '@/scripts/api'
import { useAuthStore } from '@/scripts/stores/auth'
import { ElMessage } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import ForgotPasswordDialog from './ForgotPasswordDialog.vue'

const props = defineProps<{ modelValue: boolean; mode: 'login' | 'register' }>()
const emit = defineEmits<{ 'update:modelValue': [val: boolean] }>()

const visible = computed({ get: () => props.modelValue, set: (v) => emit('update:modelValue', v) })
const authStore = useAuthStore()
const formRef = ref<FormInstance | null>(null)
const loading = ref(false)
const showForgotDialog = ref(false)
const step = ref<'check' | 'password' | 'code'>('check')
const countdown = ref(0)

const form = ref({ account: '', password: '', code: '' })

const rules: FormRules = {
  account: [
    { required: true, message: '请输入邮箱', trigger: 'blur' },
    { type: 'email', message: '请输入有效的邮箱地址', trigger: 'blur' },
  ],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }],
  code: [{ required: true, len: 8, message: '请输入8位验证码', trigger: 'blur' }],
}

watch(() => props.mode, () => { step.value = 'check' })

async function checkAccount() {
  try {
    const { data } = await api.get('/v0/auth/check-account', { params: { account: form.value.account } })
    return data.exists
  } catch {
    return false
  }
}

async function sendCode() {
  try {
    await api.post('/v0/auth/send-code', { account: form.value.account, purpose: props.mode === 'login' ? 'register' : 'reset_password' })
    ElMessage.success('验证码已发送')
    countdown.value = 60
    const timer = setInterval(() => {
      countdown.value--
      if (countdown.value <= 0) clearInterval(timer)
    }, 1000)
    step.value = 'code'
  } catch (err: any) {
    ElMessage.error(err.response?.data?.error || '发送失败')
  }
}

async function handleSubmit() {
  try {
    await formRef.value?.validate()
  } catch { return }

  loading.value = true
  try {
    if (step.value === 'check') {
      const exists = await checkAccount()
      if (exists && props.mode === 'register') {
        ElMessage.warning('该邮箱已注册，请直接登录')
        step.value = 'password'
        return
      }
      if (!exists && props.mode === 'login') {
        ElMessage.warning('该邮箱未注册，请先注册')
        step.value = 'code'
        await sendCode()
        return
      }
      if (exists) {
        step.value = 'password'
      } else {
        await sendCode()
      }
    } else if (step.value === 'password') {
      await authStore.login(form.value.account, form.value.password)
      ElMessage.success('登录成功')
      visible.value = false
      window.location.href = authStore.isAdmin ? '/admin' : '/user'
    } else if (step.value === 'code') {
      const { data } = await api.post('/v0/auth/verify-code', { account: form.value.account, code: form.value.code })
      authStore.token = data.token
      authStore.role = data.role
      authStore.account = data.account
      authStore.userId = data.userId
      authStore.balance = data.balance ?? 0
      localStorage.setItem('token', data.token)
      localStorage.setItem('role', data.role)
      localStorage.setItem('account', data.account)
      localStorage.setItem('userId', data.userId)
      localStorage.setItem('balance', String(data.balance ?? 0))
      ElMessage.success(props.mode === 'login' ? '登录成功' : '注册成功')
      visible.value = false
      window.location.href = authStore.isAdmin ? '/admin' : '/user'
    }
  } catch (err: any) {
    ElMessage.error(err.response?.data?.error || '操作失败')
  } finally {
    loading.value = false
  }
}

function resetForm() {
  step.value = 'check'
  form.value = { account: '', password: '', code: '' }
  countdown.value = 0
}
</script>
