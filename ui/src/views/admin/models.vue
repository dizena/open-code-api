<template>
  <div>
    <div class="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3 mb-4 sm:mb-6">
      <h2 class="text-lg sm:text-xl font-semibold text-gray-800">模型管理</h2>
      <el-button type="primary" @click="openCreateDialog"><el-icon><Plus /></el-icon><span class="hidden sm:inline">添加模型</span><span class="sm:hidden">添加</span></el-button>
    </div>

    <el-card shadow="never" class="border border-gray-100">
      <div class="overflow-x-auto">
        <el-table :data="models" v-loading="loading" stripe>
          <el-table-column prop="type" label="类型" width="168" />
          <el-table-column prop="name" label="名称" width="150" />
          <el-table-column prop="priority" label="优先级" width="80" />
          <el-table-column prop="baseUrl" label="Base URL" min-width="211" />
          <el-table-column label="模型数" width="100">
            <template #default="{ row }">{{ row.models?.length || 0 }}</template>
          </el-table-column>
          <el-table-column label="操作" width="210" fixed="right">
            <template #default="{ row }">
              <el-button size="small" text type="primary" @click="openEditDialog(row)">编辑</el-button>
              <el-button size="small" text type="danger" @click="handleDelete(row._id || row.id)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>
      <div class="flex justify-center mt-4">
        <el-pagination v-model:current-page="page" :page-size="pageSize" :total="total" @current-change="fetchModels" layout="prev, pager, next" />
      </div>
    </el-card>

    <el-dialog v-model="showDialog" :title="isEdit ? '编辑模型配置' : '添加模型配置'" width="95%" :style="{ maxWidth: '700px' }">
      <el-form :model="form" label-width="100px" class="max-h-[70vh] overflow-y-auto pr-2">
        <el-form-item label="类型">
          <el-select v-model="form.type" placeholder="请选择类型" >
            <el-option label="openai-compatibility" value="openai-compatibility" />
            <el-option label="gemini-api-key" value="gemini-api-key" />
            <el-option label="codex-api-key" value="codex-api-key" />
            <el-option label="claude-api-key" value="claude-api-key" />
          </el-select>
        </el-form-item>
        <el-form-item label="名称"><el-input v-model="form.name"/></el-form-item>
        <el-form-item label="优先级"><el-input-number v-model="form.priority" :min="1" /></el-form-item>
        <el-form-item label="Base URL"><el-input v-model="form.baseUrl" /></el-form-item>
        <el-form-item v-if="showApiKeys" label="API Keys">
          <div class="w-full">
            <el-tag v-for="(key, idx) in form.apiKey" :key="idx" closable @close="removeApiKey(idx)" class="mr-2 mb-1">{{ key }}</el-tag>
            <el-input v-if="inputVisible" v-model="inputValue" size="small" class="w-48" @blur="confirmInput" @keyup.enter="confirmInput" />
            <el-button v-else size="small" @click="showInput">+ 添加 Key</el-button>
          </div>
        </el-form-item>
        <el-form-item v-if="showHeaders" label="Headers">
          <div class="w-full">
            <div v-for="(h, idx) in form.headerList" :key="idx" class="flex flex-col sm:flex-row gap-2 mb-2">
              <el-input v-model="h.key" placeholder="Key" class="w-40" />
              <el-input v-model="h.value" placeholder="Value" class="flex-1" />
              <el-button size="small" type="danger" text @click="removeHeader(idx)">删除</el-button>
            </div>
            <el-button size="small" @click="addHeader">+ 添加 Header</el-button>
          </div>
        </el-form-item>
        <el-form-item label="模型列表">
          <div class="w-full">
            <div v-for="(m, idx) in form.models" :key="idx" class="flex flex-col sm:flex-row gap-2 mb-2">
              <el-input v-model="m.name" placeholder="Name" class="flex-1" />
              <el-input v-model="m.alias" placeholder="Alias" class="flex-1" />
              <el-input-number v-model="m.rate" :min="0.1" :step="0.1" class="w-28" />
              <el-button size="small" type="danger" text @click="removeModel(idx)">删除</el-button>
            </div>
            <el-button size="small" @click="addModel">+ 添加模型</el-button>
          </div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showDialog = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="handleSave">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, nextTick, computed } from 'vue'
import api from '@/scripts/api'
import { ElMessage } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'

const models = ref<any[]>([])
const loading = ref(false)
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const showDialog = ref(false)
const isEdit = ref(false)
const saving = ref(false)
const editingId = ref('')

const form = ref({
  type: 'openai-compatibility',
  name: '',
  priority: 1,
  baseUrl: '',
  apiKey: [] as string[],
  headerList: [] as { key: string; value: string }[],
  models: [] as { name: string; alias: string; rate: number }[]
})

const inputVisible = ref(false)
const inputValue = ref('')

const showApiKeys = computed(() => form.value.type !== 'openai-compatibility')
const showHeaders = computed(() => form.value.type === 'openai-compatibility')

async function fetchModels() {
  loading.value = true
  try {
    const { data } = await api.get('/v0/admin/models', { params: { page: page.value, pageSize: pageSize.value } })
    models.value = data.models || []
    total.value = data.total || 0
  } catch { ElMessage.error('获取模型失败') } finally { loading.value = false }
}

function openCreateDialog() {
  isEdit.value = false
  editingId.value = ''
  form.value = {
    type: 'openai-compatibility',
    name: '',
    priority: 1,
    baseUrl: '',
    apiKey: [],
    headerList: [{ key: 'Content-Type', value: 'application/json' }],
    models: []
  }
  showDialog.value = true
}

function openEditDialog(row: any) {
  isEdit.value = true
  editingId.value = row._id || row.id
  form.value = {
    type: row.type || '',
    name: row.name || '',
    priority: row.priority ?? 1,
    baseUrl: row.baseUrl || '',
    apiKey: [...(row.apiKey || [])],
    headerList: Object.entries(row.headers || {}).map(([key, value]) => ({ key, value: value as string })),
    models: (row.models || []).map((m: any) => ({ name: m.name || '', alias: m.alias || '', rate: m.rate ?? 1 }))
  }
  showDialog.value = true
}

function showInput() {
  inputVisible.value = true
  nextTick(() => { inputValue.value = '' })
}

function confirmInput() {
  const v = inputValue.value.trim()
  if (v && !form.value.apiKey.includes(v)) {
    form.value.apiKey.push(v)
  }
  inputVisible.value = false
  inputValue.value = ''
}

function removeApiKey(idx: number) {
  form.value.apiKey.splice(idx, 1)
}

function addHeader() {
  form.value.headerList.push({ key: '', value: '' })
}

function removeHeader(idx: number) {
  form.value.headerList.splice(idx, 1)
}

function addModel() {
  form.value.models.push({ name: '', alias: '', rate: 1 })
}

function removeModel(idx: number) {
  form.value.models.splice(idx, 1)
}

async function handleSave() {
  saving.value = true
  try {
    const headers: Record<string, string> = {}
    form.value.headerList.forEach(h => { if (h.key) headers[h.key] = h.value })

    const payload: any = {
      type: form.value.type,
      name: form.value.name,
      priority: form.value.priority,
      baseUrl: form.value.baseUrl,
      models: form.value.models
    }

    if (showApiKeys.value) {
      payload.apiKey = form.value.apiKey
    } else {
      payload.headers = headers
    }

    if (isEdit.value) {
      await api.put(`/v0/admin/models/${editingId.value}`, payload)
      ElMessage.success('更新成功')
    } else {
      await api.post('/v0/admin/models', payload)
      ElMessage.success('创建成功')
    }
    showDialog.value = false
    await fetchModels()
  } catch { ElMessage.error('保存失败') } finally { saving.value = false }
}

async function handleDelete(id: string) {
  try {
    await api.delete(`/v0/admin/models/${id}`)
    ElMessage.success('删除成功')
    await fetchModels()
  } catch { ElMessage.error('删除失败') }
}

onMounted(fetchModels)
</script>
