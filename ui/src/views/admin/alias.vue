<template>
  <div>
    <div class="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3 mb-4 sm:mb-6">
      <h2 class="text-lg sm:text-xl font-semibold text-gray-800">别名管理</h2>
    </div>

    <el-card shadow="never" class="border border-gray-100">
      <div class="overflow-x-auto">
        <el-table :data="aliases" v-loading="loading" stripe>
          <el-table-column prop="alias" label="别名" width="260" />
          <el-table-column prop="name" label="模型名称" width="260" />
          <el-table-column prop="rate" label="倍率" width="100">
            <template #default="{ row }">{{ row.rate.toFixed(2) }}</template>
          </el-table-column>
           <el-table-column prop="provider" label="提供商" width="210" />
          <el-table-column prop="url" label="Base URL" min-width="211" />
          <el-table-column label="操作" width="120" fixed="right">
            <template #default="{ row }">
              <el-button size="small" text type="primary" @click="openEditDialog(row)">编辑</el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>
    </el-card>

    <el-dialog v-model="showDialog" title="编辑别名" width="95%" :style="{ maxWidth: '500px' }">
      <el-form :model="form" label-width="80px">
        <el-form-item label="提供商">
          <el-input v-model="form.provider" disabled />
        </el-form-item>
        <el-form-item label="模型名称">
          <el-input v-model="form.name" disabled />
        </el-form-item>
        <el-form-item label="别名">
          <el-input v-model="form.alias" placeholder="请输入别名" />
        </el-form-item>
        <el-form-item label="倍率">
          <el-input-number v-model="form.rate" :min="0.1" :step="0.1" class="w-full" />
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
import { ref, onMounted } from 'vue'
import api from '@/scripts/api'
import { ElMessage } from 'element-plus'

interface AliasItem {
  docId: string
  provider: string
  url: string
  name: string
  alias: string
  rate: number
  index: number
}

const aliases = ref<AliasItem[]>([])
const loading = ref(false)
const saving = ref(false)
const showDialog = ref(false)

const form = ref({
  docId: '',
  provider: '',
  name: '',
  alias: '',
  rate: 1.0,
  index: 0,
})

async function fetchAliases() {
  loading.value = true
  try {
    const { data } = await api.get('/v0/admin/aliases')
    const fetched = (data.aliases || []) as AliasItem[]
    aliases.value = fetched.sort((a: AliasItem, b: AliasItem) =>
      (a.alias || "").localeCompare(b.alias || "")
    )
  } catch {
    ElMessage.error('获取别名失败')
  } finally {
    loading.value = false
  }
}

function openEditDialog(row: AliasItem) {
  form.value = {
    docId: row.docId,
    provider: row.provider,
    name: row.name,
    alias: row.alias,
    rate: row.rate,
    index: row.index,
  }
  showDialog.value = true
}

async function handleSave() {
  saving.value = true
  try {
    await api.put(`/v0/admin/models/${form.value.docId}/alias`, {
      index: form.value.index,
      alias: form.value.alias.trim(),
      rate: form.value.rate,
    })
    ElMessage.success('更新成功')
    showDialog.value = false
    await fetchAliases()
  } catch {
    ElMessage.error('保存失败')
  } finally {
    saving.value = false
  }
}

onMounted(fetchAliases)
</script>
