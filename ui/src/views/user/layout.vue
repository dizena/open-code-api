<template>
  <div class="user-shell min-h-screen">
    <div class="shell-orb shell-orb-cyan"></div>
    <div class="shell-orb shell-orb-indigo"></div>
    <div class="shell-orb shell-orb-center"></div>

    <el-container class="relative min-h-screen bg-transparent">
      <el-drawer v-model="drawerVisible" :size="drawerSize" direction="ltr" class="user-drawer md:hidden">
        <template #header>
          <div class="flex items-center gap-3">
            <div class="brand-logo-shell h-11 w-11">
              <img src="https://www.dizena.com/logo.png" alt="CodeAPI logo" class="brand-logo-image">
            </div>
            <div>
              <p class="text-[11px] uppercase tracking-[0.28em] text-cyan-200/70">User Space</p>
              <span class="brand-text text-lg font-semibold">CodeAPI</span>
            </div>
          </div>
        </template>
        <el-menu :default-active="route.path" router class="user-menu border-0" :ellipsis="false" @select="drawerVisible = false">
          <el-menu-item index="/user/models">
            <el-icon><Collection /></el-icon>
            <span>可用模型</span>
          </el-menu-item>
          <el-menu-item index="/user/keys">
            <el-icon><Key /></el-icon>
            <span>API Keys</span>
          </el-menu-item>
          <el-menu-item index="/user/usage">
            <el-icon><DataAnalysis /></el-icon>
            <span>用量统计</span>
          </el-menu-item>
        </el-menu>
      </el-drawer>

      <el-aside :width="asideWidth" class="hidden border-0 bg-transparent px-5 py-5 md:block">
        <aside class="user-sidebar">
          <div class="sidebar-header">
            <div class="flex items-center gap-3">
              <div class="brand-logo-shell h-11 w-11">
                <img src="https://www.dizena.com/logo.png" alt="CodeAPI logo" class="brand-logo-image">
              </div>
              <div class="min-w-0">
                <p class="text-[11px] uppercase tracking-[0.28em] text-cyan-200/70">User Space</p>
                <span class="brand-text truncate text-lg font-semibold">CodeAPI</span>
              </div>
            </div>
          </div>

          <div class="px-3 py-4">
            <el-menu :default-active="route.path" router class="user-menu border-0" :ellipsis="false">
              <el-menu-item index="/user/models">
                <el-icon><Collection /></el-icon>
                <span>可用模型</span>
              </el-menu-item>
              <el-menu-item index="/user/keys">
                <el-icon><Key /></el-icon>
                <span>API Keys</span>
              </el-menu-item>
              <el-menu-item index="/user/usage">
                <el-icon><DataAnalysis /></el-icon>
                <span>用量统计</span>
              </el-menu-item>
            </el-menu>
          </div>

          <div class="mt-auto p-4">
            <div class="sidebar-runtime-card">
              <div class="runtime-status-dot"></div>
              <p class="text-xs uppercase tracking-[0.26em] text-slate-400">Runtime</p>
              <p class="mt-3 text-sm leading-7 text-slate-200/85">轻量控制台已就绪，Key、模型和用量信息会统一汇聚到这里。</p>
            </div>
          </div>
        </aside>
      </el-aside>

      <el-container class="bg-transparent">
        <el-header class="border-0 bg-transparent px-3 pt-3 sm:px-4 md:h-auto md:px-0 md:pt-5" height="auto">
          <div class="mx-auto max-w-7xl px-0 md:px-6">
            <div class="topbar-panel">
              <div class="flex items-start gap-3 sm:items-center">
                <el-button class="md:hidden" text @click="drawerVisible = true">
                  <el-icon :size="20" class="text-slate-100"><Promotion /></el-icon>
                </el-button>
                <div>
                  <p class="text-xs uppercase tracking-[0.28em] text-cyan-200/70">Workspace</p>
                  <h1 class="mt-1 text-lg font-semibold text-white sm:text-xl">欢迎回来，{{ authStore.account }}</h1>
                </div>
              </div>

              <div class="flex flex-col items-stretch gap-3 sm:flex-row sm:items-center">
                <div class="balance-chip">
                  <span class="text-xs uppercase tracking-[0.22em] text-slate-400">Balance</span>
                  <span class="balance-value text-lg font-semibold">{{ authStore.balance }} <span class="text-sm font-normal text-slate-300">积分</span></span>
                </div>
                <div class="flex items-center justify-end gap-2">
                  <el-button size="small" type="primary" class="user-primary-btn" @click="showRechargeDialog = true">
                    <el-icon><Wallet /></el-icon>
                    <span class="hidden sm:inline">充值</span>
                  </el-button>
                  <el-button v-if="authStore.isAdmin" size="small" class="user-secondary-btn" @click="handleSwitchToAdmin">
                    <el-icon><Switch /></el-icon>
                    <span class="hidden sm:inline">切换管理</span>
                  </el-button>
                  <el-button size="small" class="user-secondary-btn" @click="showPasswordDialog = true">
                    <el-icon><Lock /></el-icon>
                    <span class="hidden sm:inline">修改密码</span>
                  </el-button>
                  <el-button size="small" class="user-secondary-btn" @click="handleLogout">
                    <el-icon><SwitchButton /></el-icon>
                    <span class="hidden sm:inline">退出</span>
                  </el-button>
                </div>
              </div>
            </div>
          </div>
        </el-header>

        <el-main class="bg-transparent px-3 pb-4 pt-3 sm:px-4 md:px-0 md:pb-6 md:pt-4">
          <div class="mx-auto max-w-7xl md:px-6">
            <router-view />
          </div>
        </el-main>
      </el-container>

      <el-dialog v-model="showRechargeDialog" title="充值" width="400px" class="user-dialog">
        <div class="py-3 text-center">
          <p class="mb-4 text-gray-600">请接入您的支付系统</p>
        </div>
      </el-dialog>

      <el-dialog v-model="showPasswordDialog" title="修改密码" width="400px" class="user-dialog" @closed="resetPasswordForm">
        <el-form ref="passwordFormRef" :model="passwordForm" :rules="passwordRules" label-position="top" @submit.prevent="handleChangePassword">
          <el-form-item label="旧密码" prop="oldPassword">
            <el-input v-model="passwordForm.oldPassword" type="password" show-password placeholder="请输入旧密码" />
          </el-form-item>
          <el-form-item label="新密码" prop="newPassword">
            <el-input v-model="passwordForm.newPassword" type="password" show-password placeholder="请输入新密码（至少6位）" />
          </el-form-item>
          <el-form-item label="确认新密码" prop="confirmPassword">
            <el-input v-model="passwordForm.confirmPassword" type="password" show-password placeholder="请再次输入新密码" />
          </el-form-item>
        </el-form>
        <template #footer>
          <el-button @click="showPasswordDialog = false">取消</el-button>
          <el-button type="primary" :loading="passwordLoading" @click="handleChangePassword">确认修改</el-button>
        </template>
      </el-dialog>
    </el-container>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/scripts/stores/auth'
import { Key, DataAnalysis, SwitchButton, Promotion, Collection, Switch, Wallet, Lock } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const drawerVisible = ref(false)
const showRechargeDialog = ref(false)

const asideWidth = computed(() => {
  if (typeof window !== 'undefined' && window.innerWidth >= 1280) return '280px'
  return '236px'
})

const drawerSize = computed(() => {
  if (typeof window !== 'undefined' && window.innerWidth < 640) return '84%'
  return '300px'
})

const showPasswordDialog = ref(false)
const passwordLoading = ref(false)
const passwordFormRef = ref()
const passwordForm = reactive({
  oldPassword: '',
  newPassword: '',
  confirmPassword: '',
})

const validateConfirmPassword = (_rule: unknown, value: string, callback: (err?: Error) => void) => {
  if (value !== passwordForm.newPassword) {
    callback(new Error('两次输入的新密码不一致'))
  } else {
    callback()
  }
}

const passwordRules = {
  oldPassword: [{ required: true, message: '请输入旧密码', trigger: 'blur' }],
  newPassword: [
    { required: true, message: '请输入新密码', trigger: 'blur' },
    { min: 6, message: '新密码至少6位', trigger: 'blur' },
  ],
  confirmPassword: [
    { required: true, message: '请再次输入新密码', trigger: 'blur' },
    { validator: validateConfirmPassword, trigger: 'blur' },
  ],
}

function resetPasswordForm() {
  passwordForm.oldPassword = ''
  passwordForm.newPassword = ''
  passwordForm.confirmPassword = ''
  passwordFormRef.value?.resetFields()
}

async function handleChangePassword() {
  const valid = await passwordFormRef.value?.validate().catch(() => false)
  if (!valid) return

  passwordLoading.value = true
  try {
    await authStore.changePassword(passwordForm.oldPassword, passwordForm.newPassword)
    ElMessage.success('密码修改成功')
    showPasswordDialog.value = false
  } catch (err: unknown) {
    const msg = (err as { response?: { data?: { error?: string } } })?.response?.data?.error || '修改失败'
    ElMessage.error(msg)
  } finally {
    passwordLoading.value = false
  }
}

onMounted(() => {
  authStore.fetchBalance()
})

function handleSwitchToAdmin() {
  router.push('/admin')
}

function handleLogout() {
  authStore.logout()
  window.location.href = '/'
}
</script>

<style scoped>
.user-shell {
  position: relative;
  overflow: hidden;
  background:
    radial-gradient(ellipse 80% 50% at 12% 8%, rgba(34, 211, 238, 0.16), transparent 50%),
    radial-gradient(ellipse 60% 45% at 85% 5%, rgba(99, 102, 241, 0.14), transparent 48%),
    linear-gradient(180deg, #020617 0%, #0b1224 38%, #e2e8f0 38.5%, #f8fafc 100%);
  font-family: 'Avenir Next', 'Segoe UI Variable', 'PingFang SC', 'Hiragino Sans GB', 'Microsoft YaHei', sans-serif;
}

/* ===== Aurora orbs (animated) ===== */
.shell-orb {
  pointer-events: none;
  position: absolute;
  border-radius: 9999px;
  filter: blur(100px);
  will-change: transform, opacity;
}

.shell-orb-cyan {
  left: -6rem;
  top: 4rem;
  height: 20rem;
  width: 20rem;
  background: radial-gradient(circle, rgba(34, 211, 238, 0.2), rgba(6, 182, 212, 0.04) 70%, transparent);
  animation: shell-drift-1 14s ease-in-out infinite alternate;
}

.shell-orb-indigo {
  right: -2rem;
  top: 8rem;
  height: 20rem;
  width: 20rem;
  background: radial-gradient(circle, rgba(129, 140, 248, 0.16), rgba(99, 102, 241, 0.04) 70%, transparent);
  animation: shell-drift-2 16s ease-in-out infinite alternate;
}

.shell-orb-center {
  left: 35%;
  top: 12rem;
  height: 14rem;
  width: 14rem;
  background: radial-gradient(circle, rgba(14, 165, 233, 0.1), transparent 70%);
  animation: shell-drift-3 20s ease-in-out infinite alternate;
}

@keyframes shell-drift-1 {
  0%   { transform: translate(0, 0) scale(1); opacity: 0.5; }
  100% { transform: translate(2.5rem, 1.5rem) scale(1.12); opacity: 0.7; }
}
@keyframes shell-drift-2 {
  0%   { transform: translate(0, 0) scale(1); opacity: 0.45; }
  100% { transform: translate(-2rem, 2rem) scale(1.08); opacity: 0.6; }
}
@keyframes shell-drift-3 {
  0%   { transform: translate(0, 0) scale(1); opacity: 0.25; }
  100% { transform: translate(1.5rem, -1.5rem) scale(1.18); opacity: 0.4; }
}

/* ===== Sidebar ===== */
.user-sidebar,
.topbar-panel {
  border: 1px solid rgba(255, 255, 255, 0.07);
  background: linear-gradient(165deg, rgba(15, 23, 42, 0.78), rgba(15, 23, 42, 0.58));
  box-shadow:
    0 0 0 1px rgba(255, 255, 255, 0.025) inset,
    0 30px 80px rgba(15, 23, 42, 0.26);
  backdrop-filter: blur(28px) saturate(1.2);
}

.user-sidebar {
  display: flex;
  min-height: calc(100vh - 2.5rem);
  flex-direction: column;
  overflow: hidden;
  border-radius: 28px;
}

.sidebar-header {
  border-bottom: 1px solid rgba(255, 255, 255, 0.07);
  padding: 1.25rem;
}

.brand-text {
  background: linear-gradient(135deg, #fff 40%, rgba(103, 232, 249, 0.85));
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

/* Runtime card at bottom of sidebar */
.sidebar-runtime-card {
  position: relative;
  border-radius: 20px;
  border: 1px solid rgba(255, 255, 255, 0.07);
  background: rgba(255, 255, 255, 0.04);
  padding: 1rem;
}

.runtime-status-dot {
  position: absolute;
  top: 1rem;
  right: 1rem;
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: #34d399;
  box-shadow: 0 0 8px rgba(52, 211, 153, 0.6);
  animation: runtime-pulse 2.8s ease-in-out infinite;
}

@keyframes runtime-pulse {
  0%, 100% { box-shadow: 0 0 6px rgba(52, 211, 153, 0.45); opacity: 0.75; }
  50%      { box-shadow: 0 0 14px rgba(52, 211, 153, 0.85); opacity: 1; }
}

/* ===== Topbar ===== */
.topbar-panel {
  display: flex;
  flex-direction: column;
  gap: 1rem;
  border-radius: 24px;
  padding: 1.1rem 1.2rem;
}

.balance-chip {
  display: inline-flex;
  min-width: 132px;
  flex-direction: column;
  gap: 0.25rem;
  border-radius: 18px;
  border: 1px solid rgba(255, 255, 255, 0.07);
  background: rgba(255, 255, 255, 0.035);
  padding: 0.75rem 1rem;
  transition: border-color 0.3s ease, background 0.3s ease;
}

.balance-chip:hover {
  border-color: rgba(103, 232, 249, 0.14);
  background: rgba(255, 255, 255, 0.05);
}

.balance-value {
  background: linear-gradient(135deg, #fff 35%, rgba(103, 232, 249, 0.85));
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

/* ===== Logo ===== */
.brand-logo-shell {
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  border-radius: 0.9rem;
  background: rgba(255, 255, 255, 0.98);
  box-shadow:
    0 0 18px rgba(34, 211, 238, 0.14),
    0 12px 30px rgba(14, 165, 233, 0.18);
  transition: box-shadow 0.4s ease;
}

.brand-logo-shell:hover {
  box-shadow:
    0 0 26px rgba(34, 211, 238, 0.24),
    0 14px 36px rgba(14, 165, 233, 0.26);
}

.brand-logo-image {
  height: 100%;
  width: 100%;
  object-fit: cover;
}

/* ===== Menu ===== */
:deep(.user-menu.el-menu) {
  background: transparent;
}

:deep(.user-menu .el-menu-item) {
  margin-bottom: 0.35rem;
  border-radius: 16px;
  color: #cbd5e1;
  transition: all 0.3s ease;
}

:deep(.user-menu .el-menu-item:hover) {
  background: rgba(255, 255, 255, 0.06);
  color: #fff;
}

:deep(.user-menu .el-menu-item.is-active) {
  background: linear-gradient(135deg, rgba(6, 182, 212, 0.2), rgba(59, 130, 246, 0.18));
  color: #fff;
  box-shadow: 0 0 20px rgba(6, 182, 212, 0.06);
}

:deep(.user-menu .el-menu-item.is-active::after) {
  display: none;
}

/* ===== Drawer ===== */
:deep(.user-drawer .el-drawer) {
  background: linear-gradient(180deg, #020617 0%, #0f172a 100%);
}

:deep(.user-drawer .el-drawer__header) {
  margin-bottom: 0;
  border-bottom: 1px solid rgba(255, 255, 255, 0.08);
  padding-bottom: 1rem;
}

:deep(.user-drawer .el-drawer__body) {
  padding: 1rem;
}

/* ===== Buttons ===== */
:deep(.user-primary-btn.el-button) {
  border: 1px solid rgba(125, 211, 252, 0.25);
  background: linear-gradient(135deg, #0891b2, #2563eb);
  box-shadow:
    0 0 18px rgba(37, 99, 235, 0.14),
    0 14px 32px rgba(37, 99, 235, 0.2);
  transition: all 0.35s ease;
}

:deep(.user-primary-btn.el-button:hover) {
  border-color: rgba(186, 230, 253, 0.35);
  background: linear-gradient(135deg, #06b6d4, #3b82f6);
  box-shadow:
    0 0 28px rgba(37, 99, 235, 0.22),
    0 18px 40px rgba(37, 99, 235, 0.28);
  transform: translateY(-1px);
}

:deep(.user-secondary-btn.el-button) {
  border: 1px solid rgba(255, 255, 255, 0.08);
  background: rgba(255, 255, 255, 0.04);
  color: #cbd5e1;
  transition: all 0.3s ease;
}

:deep(.user-secondary-btn.el-button:hover) {
  border-color: rgba(103, 232, 249, 0.18);
  background: rgba(255, 255, 255, 0.08);
  color: #fff;
  box-shadow: 0 0 16px rgba(34, 211, 238, 0.05);
}

/* ===== Dialog ===== */
:deep(.user-dialog .el-dialog) {
  border-radius: 20px;
}

/* ===== Responsive ===== */
@media (min-width: 640px) {
  .topbar-panel {
    flex-direction: row;
    align-items: center;
    justify-content: space-between;
    padding: 1.2rem 1.4rem;
  }
}

@media (prefers-reduced-motion: reduce) {
  .shell-orb,
  .runtime-status-dot {
    animation: none !important;
  }
}
</style>
