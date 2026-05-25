<template>
  <div>
    <div class="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3 mb-4 sm:mb-6">
      <h2 class="text-lg sm:text-xl font-semibold text-gray-800">用户管理</h2>
      <el-button type="primary" @click="showCreate = true">
        <el-icon><Plus /></el-icon>
        <span class="hidden sm:inline">创建用户</span>
        <span class="sm:hidden">创建</span>
      </el-button>
    </div>

    <el-card shadow="never" class="border border-gray-100">
      <div class="overflow-x-auto">
        <el-table :data="users" v-loading="loading" stripe>
          <el-table-column prop="account" label="账户" min-width="120" />
          <el-table-column label="角色" width="100">
            <template #default="{ row }">
              <el-tag :type="row.role === 'admin' ? 'danger' : 'success'" size="small">
                {{ row.role === 'admin' ? '管理' : '用户' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="balance" label="积分" width="100" />
          <el-table-column label="状态" width="80">
            <template #default="{ row }">
              <el-tag :type="row.status === 1 ? 'success' : 'danger'" size="small">
                {{ row.status === 1 ? '正常' : '禁用' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="createAt" label="创建时间" width="180">
            <template #default="{ row }">
              {{ formatCreateTime(row.createAt) }}
            </template>
          </el-table-column>
          <el-table-column label="操作" width="211" fixed="right">
            <template #default="{ row }">
              <el-button size="small" text type="primary" @click="openEditDialog(row)">编辑</el-button>
              <el-button size="small" text type="primary" @click="handleToggleStatus(row)">
                {{ row.status === 1 ? '禁用' : '启用' }}
              </el-button>
              <el-popconfirm title="确认删除？" @confirm="handleDelete(row.account)">
                <template #reference>
                  <el-button size="small" text type="danger">删除</el-button>
                </template>
              </el-popconfirm>
            </template>
          </el-table-column>
        </el-table>
      </div>

      <div class="flex justify-center mt-4">
        <el-pagination v-model:current-page="page" :page-size="pageSize" :total="total" @current-change="fetchUsers" layout="prev, pager, next" />
      </div>
    </el-card>

    <el-dialog v-model="showCreate" title="创建用户" width="90%" :style="{ maxWidth: '400px' }">
      <el-form :model="createForm" label-width="80px">
        <el-form-item label="账户"><el-input v-model="createForm.account" /></el-form-item>
        <el-form-item label="密码"><el-input v-model="createForm.password" type="password" /></el-form-item>
        <el-form-item label="角色">
          <el-select v-model="createForm.role" class="w-full"><el-option label="用户" value="user" /><el-option label="管理员" value="admin" /></el-select>
        </el-form-item>
        <el-form-item label="积分"><el-input-number v-model="createForm.balance" :min="0" class="w-full" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCreate = false">取消</el-button>
        <el-button type="primary" :loading="creating" @click="handleCreate">创建</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="showEdit" title="编辑用户" width="90%" :style="{ maxWidth: '400px' }">
      <el-form :model="editForm" label-width="80px">
        <el-form-item label="账户"><el-input v-model="editForm.account" disabled /></el-form-item>
        <el-form-item label="角色">
          <el-select v-model="editForm.role" class="w-full"><el-option label="用户" value="user" /><el-option label="管理员" value="admin" /></el-select>
        </el-form-item>
        <el-form-item label="积分"><el-input-number v-model="editForm.balance" :min="0" class="w-full" /></el-form-item>
        <el-form-item label="状态">
          <el-select v-model="editForm.status" class="w-full"><el-option label="正常" :value="1" /><el-option label="禁用" :value="0" /></el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showEdit = false">取消</el-button>
        <el-button type="primary" :loading="editing" @click="handleEdit">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import api from '@/scripts/api'
import { ElMessage } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import dayjs from 'dayjs'

interface User { id: string; account: string; role: string; balance: number; status: number; createAt: string }

const users = ref<User[]>([])
const loading = ref(false)
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const showCreate = ref(false)
const creating = ref(false)
const createForm = ref({ account: '', password: '', role: 'user', balance: 0 })
const showEdit = ref(false)
const editing = ref(false)
const editForm = ref({ account: '', role: 'user', balance: 0, status: 1 })

async function fetchUsers() {
  loading.value = true
  try {
    const { data } = await api.get('/v0/admin/users', { params: { page: page.value, pageSize: pageSize.value } })
    users.value = data.users || []
    total.value = data.total || 0
  } catch { ElMessage.error('获取用户列表失败') } finally { loading.value = false }
}

async function handleCreate() {
  if (!createForm.value.account || !createForm.value.password) return
  creating.value = true
  try {
    await api.post('/v0/admin/users', createForm.value)
    ElMessage.success('创建成功')
    showCreate.value = false
    createForm.value = { account: '', password: '', role: 'user', balance: 0 }
    await fetchUsers()
  } catch { ElMessage.error('创建失败') } finally { creating.value = false }
}

function openEditDialog(row: User) {
  editForm.value = { account: row.account, role: row.role, balance: row.balance, status: row.status }
  showEdit.value = true
}

async function handleEdit() {
  editing.value = true
  try {
    await api.put(`/v0/admin/users/${editForm.value.account}`, {
      role: editForm.value.role,
      balance: editForm.value.balance,
      status: editForm.value.status
    })
    ElMessage.success('更新成功')
    showEdit.value = false
    await fetchUsers()
  } catch { ElMessage.error('更新失败') } finally { editing.value = false }
}

async function handleToggleStatus(user: User) {
  try {
    await api.put(`/v0/admin/users/${user.account}`, { status: user.status === 1 ? 0 : 1 })
    ElMessage.success('操作成功')
    await fetchUsers()
  } catch { ElMessage.error('操作失败') }
}

async function handleDelete(account: string) {
  try {
    await api.delete(`/v0/admin/users/${account}`)
    ElMessage.success('删除成功')
    await fetchUsers()
  } catch { ElMessage.error('删除失败') }
}

function formatCreateTime(date: string) {
  if (!date || date.startsWith('0001-') || date.startsWith('0000-')) return '-'
  return dayjs(date).format('YYYY-MM-DD HH:mm')
}

function formatDate(date: string) { return date ? dayjs(date).format('YYYY-MM-DD HH:mm') : '-' }

onMounted(fetchUsers)
</script>
