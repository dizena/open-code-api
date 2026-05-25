<template>
  <div>
    <div class="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3 mb-4 sm:mb-6">
      <h2 class="text-lg sm:text-xl font-semibold text-gray-800">Key 管理</h2>
      <el-button type="primary" @click="showCreate = true"><el-icon><Plus /></el-icon><span class="hidden sm:inline">创建 Key</span><span class="sm:hidden">创建</span></el-button>
    </div>

    <el-card shadow="never" class="border border-gray-100 mb-4">
      <el-form :inline="true" class="flex flex-wrap gap-2">
        <el-form-item label="用户账户" class="w-full sm:w-auto">
          <el-input v-model="account" placeholder="留空查询所有" class="w-full sm:w-48" />
        </el-form-item>
        <el-form-item label="开始日期" class="w-full sm:w-auto">
          <el-date-picker v-model="startDate" type="date" value-format="YYYY-MM-DD" class="w-full sm:w-40" />
        </el-form-item>
        <el-form-item label="结束日期" class="w-full sm:w-auto">
          <el-date-picker v-model="endDate" type="date" value-format="YYYY-MM-DD" class="w-full sm:w-40" />
        </el-form-item>
        <el-form-item class="w-full sm:w-auto">
          <el-button type="primary" @click="handleQuery" class="w-full sm:w-auto">查询</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card shadow="never" class="border border-gray-100">
      <div class="overflow-x-auto">
        <el-table :data="keys" v-loading="loading" stripe>
          <el-table-column prop="name" label="名称" width="150" />
          <el-table-column prop="key" label="Key" min-width="330">
            <template #default="{ row }">
              <div class="flex items-center gap-2">
                <span class="truncate">{{ row.key }}</span>
                <el-button size="small" text @click="copyKey(row.key)">
                  <el-icon><CopyDocument /></el-icon>
                </el-button>
              </div>
            </template>
          </el-table-column>
          <el-table-column label="状态" width="80">
            <template #default="{ row }">
              <el-tag :type="row.status === 1 ? 'success' : 'danger'" size="small">{{ row.status === 1 ? '正常' : '禁用' }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="createdAt" label="创建时间" width="210">
            <template #default="{ row }">{{ formatDate(row.createdAt) }}</template>
          </el-table-column>
          <el-table-column label="操作" width="210" fixed="right">
            <template #default="{ row }">
              <el-button size="small" text type="warning" @click="handleDisable(row.key)">禁用</el-button>
              <el-popconfirm title="确认删除？" @confirm="handleDelete(row.key)">
                <template #reference><el-button size="small" text type="danger">删除</el-button></template>
              </el-popconfirm>
            </template>
          </el-table-column>
        </el-table>
      </div>
      <div class="flex justify-center mt-4">
        <el-pagination v-model:current-page="page" :page-size="pageSize" :total="total" @current-change="fetchKeys" layout="prev, pager, next" />
      </div>
    </el-card>

    <el-dialog v-model="showCreate" title="为用户创建 Key" width="90%" :style="{ maxWidth: '400px' }">
      <el-form :model="createForm" label-width="80px">
        <el-form-item label="账户"><el-input v-model="createForm.account" /></el-form-item>
        <el-form-item label="名称"><el-input v-model="createForm.name" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCreate = false">取消</el-button>
        <el-button type="primary" :loading="creating" @click="handleCreate">创建</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import api from '@/scripts/api'
import { ElMessage } from 'element-plus'
import { Plus, CopyDocument } from '@element-plus/icons-vue'
import dayjs from 'dayjs'

const keys = ref<any[]>([])
const loading = ref(false)
const account = ref('')
const startDate = ref('')
const endDate = ref('')
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const showCreate = ref(false)
const creating = ref(false)
const createForm = ref({ account: '', name: '' })

function handleQuery() {
  page.value = 1
  fetchKeys()
}

async function fetchKeys() {
  loading.value = true
  try {
    const params: any = { page: page.value, pageSize: pageSize.value }
    if (account.value) params.account = account.value
    if (startDate.value) params.start = startDate.value
    if (endDate.value) params.end = endDate.value
    const { data } = await api.get('/v0/admin/keys', { params })
    keys.value = data.keys || []
    total.value = data.total || 0
  } catch { ElMessage.error('获取 Keys 失败') } finally { loading.value = false }
}

async function handleCreate() {
  creating.value = true
  try {
    const { data } = await api.post('/v0/admin/keys', createForm.value)
    ElMessage.success(`Key 创建成功: ${data.key}`)
    showCreate.value = false
    createForm.value = { account: '', name: '' }
    await fetchKeys()
  } catch { ElMessage.error('创建失败') } finally { creating.value = false }
}

async function handleDisable(key: string) {
  try {
    await api.put(`/v0/admin/keys/${key}/disable`)
    ElMessage.success('已禁用')
    await fetchKeys()
  } catch { ElMessage.error('操作失败') }
}

async function handleDelete(key: string) {
  try {
    await api.delete(`/v0/admin/keys/${key}`)
    ElMessage.success('删除成功')
    await fetchKeys()
  } catch { ElMessage.error('删除失败') }
}

async function copyKey(key: string) {
  try {
    await navigator.clipboard.writeText(key)
    ElMessage.success('已复制到剪贴板')
  } catch {
    ElMessage.error('复制失败')
  }
}

function formatDate(date: string) { return date ? dayjs(date).format('YYYY-MM-DD HH:mm') : '-' }

onMounted(fetchKeys)
</script>
