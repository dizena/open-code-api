<template>
  <el-dialog v-model="visible" title="忘记密码" width="420px" @close="resetForm">
    <el-form ref="formRef" :model="form" :rules="rules" label-width="0">
      <el-form-item prop="account">
        <el-input v-model="form.account" placeholder="邮箱地址" prefix-icon="Message" />
      </el-form-item>

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
      <el-button type="primary" :loading="loading" class="w-full" @click="handleSubmit">
        {{ step === 'check' ? '发送验证码' : '验证并重置密码' }}
      </el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import api from '@/scripts/api'
import { ElMessage } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'

const props = defineProps<{ modelValue: boolean }>()
const emit = defineEmits<{ 'update:modelValue': [val: boolean] }>()

const visible = computed({ get: () => props.modelValue, set: (v) => emit('update:modelValue', v) })
const formRef = ref<FormInstance | null>(null)
const loading = ref(false)
const step = ref<'check' | 'code'>('check')
const countdown = ref(0)

const form = ref({ account: '', code: '' })

const rules: FormRules = {
  account: [
    { required: true, message: '请输入邮箱', trigger: 'blur' },
    { type: 'email', message: '请输入有效的邮箱地址', trigger: 'blur' },
  ],
  code: [{ required: true, len: 8, message: '请输入8位验证码', trigger: 'blur' }],
}

async function sendCode() {
  try {
    await api.post('/v0/auth/send-code', { account: form.value.account, purpose: 'reset_password' })
    ElMessage.success('验证码已发送到邮箱')
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
      await sendCode()
    } else {
      const { data } = await api.post('/v0/auth/verify-code', { account: form.value.account, code: form.value.code })
      ElMessage.success('密码已重置，请使用新密码登录')
      visible.value = false
    }
  } catch (err: any) {
    ElMessage.error(err.response?.data?.error || '操作失败')
  } finally {
    loading.value = false
  }
}

function resetForm() {
  step.value = 'check'
  form.value = { account: '', code: '' }
  countdown.value = 0
}
</script>
