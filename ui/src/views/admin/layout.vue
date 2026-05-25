<template>
  <el-container class="min-h-screen bg-gray-50">
    <el-drawer v-model="drawerVisible" :size="drawerSize" direction="ltr" class="md:hidden">
      <template #header>
        <div class="flex items-center gap-3">
          <div class="w-9 h-9 rounded-lg bg-gradient-to-br from-rose-400 to-orange-500 flex items-center justify-center">
            <el-icon :size="20" color="#fff"><Setting /></el-icon>
          </div>
          <span class="font-semibold text-gray-800">管理面板</span>
        </div>
      </template>
      <el-menu :default-active="route.path" router class="border-0" :ellipsis="false" @select="drawerVisible = false">
        <el-menu-item index="/admin/models">
          <el-icon><Grid /></el-icon>
          <span>模型管理</span>
        </el-menu-item>
         <el-menu-item index="/admin/alias">
          <el-icon><Menu /></el-icon>
          <span>别名管理</span>
        </el-menu-item>
        <el-menu-item index="/admin/users">
          <el-icon><UserFilled /></el-icon>
          <span>用户管理</span>
        </el-menu-item>
        <el-menu-item index="/admin/keys">
          <el-icon><Key /></el-icon>
          <span>凭证管理</span>
        </el-menu-item>
        <el-menu-item index="/admin/usage">
          <el-icon><DataAnalysis /></el-icon>
          <span>用量查询</span>
        </el-menu-item>

        <el-divider class="my-2" />
      </el-menu>
    </el-drawer>

    <el-aside :width="asideWidth" class="bg-white shadow-sm border-r border-gray-100 hidden md:block">
      <div class="p-5 border-b border-gray-100">
        <div class="flex items-center gap-3">
          <div class="w-9 h-9 rounded-lg bg-gradient-to-br from-rose-400 to-orange-500 flex items-center justify-center">
            <el-icon :size="20" color="#fff"><Setting /></el-icon>
          </div>
          <span class="font-semibold text-gray-800 hidden lg:block">管理面板</span>
        </div>
      </div>
      <el-menu :default-active="route.path" router class="border-0" :ellipsis="false">
        <el-menu-item index="/admin/models">
          <el-icon><Grid /></el-icon>
          <span>模型管理</span>
        </el-menu-item>
        <el-menu-item index="/admin/alias">
          <el-icon><Menu /></el-icon>
          <span>别名管理</span>
        </el-menu-item>  
        <el-menu-item index="/admin/users">
          <el-icon><UserFilled /></el-icon>
          <span>用户管理</span>
        </el-menu-item>              
        <el-menu-item index="/admin/keys">
          <el-icon><Key /></el-icon>
          <span>凭证管理</span>
        </el-menu-item>
        <el-menu-item index="/admin/usage">
          <el-icon><DataAnalysis /></el-icon>
          <span>用量查询</span>
        </el-menu-item>
         <el-divider class="my-2" />
      </el-menu>
    </el-aside>

    <el-container>
      <el-header class="bg-white border-b border-gray-100 flex items-center justify-between px-4 md:px-6" height="60px">
        <div class="flex items-center gap-3">
          <el-button class="md:hidden" text @click="drawerVisible = true">
            <el-icon :size="20"><Setting /></el-icon>
          </el-button>
          <span class="text-gray-500 text-sm">管理员：{{ authStore.account }}</span>
        </div>
        <div class="flex items-center gap-2 md:gap-3">
          <el-button size="small" text @click="handleSwitchToUser">
            <el-icon><Switch /></el-icon>
            <span class="hidden sm:inline">切换用户</span>
          </el-button>
          <el-button size="small" text @click="handleLogout">
            <el-icon><SwitchButton /></el-icon>
            <span class="hidden sm:inline">退出</span>
          </el-button>
        </div>
      </el-header>

      <el-main class="p-3 sm:p-4 md:p-6">
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup lang="ts">
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/scripts/stores/auth'
import { UserFilled, Grid, Key, DataAnalysis, Switch, SwitchButton, Setting, Menu } from '@element-plus/icons-vue'
import { ref, computed } from 'vue'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const drawerVisible = ref(false)

const asideWidth = computed(() => {
  if (typeof window !== 'undefined' && window.innerWidth >= 1280) return '240px'
  return '200px'
})

const drawerSize = computed(() => {
  if (typeof window !== 'undefined' && window.innerWidth < 640) return '75%'
  return '280px'
})

function handleSwitchToUser() {
  router.push('/user')
}

function handleLogout() {
  authStore.logout()
  window.location.href = '/'
}
</script>
