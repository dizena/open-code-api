<template>
  <div>
    <h2 class="text-lg sm:text-xl font-semibold text-gray-800 mb-4 sm:mb-6">用量查询</h2>

    <el-card shadow="never" class="border border-gray-100 mb-4">
      <el-form :inline="true" class="flex flex-wrap gap-2">
        <el-form-item label="用户账户" class="w-full sm:w-auto">
          <el-input v-model="account" placeholder="输入账户" class="w-full sm:w-48" />
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
        <el-table :data="records" v-loading="loading" stripe>
          <el-table-column prop="apiKey" label="API Key" width="150" />
          <el-table-column prop="provider" label="提供商" width="120" />
          <el-table-column prop="alias" label="模型" width="260" />
          <el-table-column prop="model" label="Model" width="260" />
          <el-table-column prop="inputTokens" label="输入" width="80" />
          <el-table-column prop="outputTokens" label="输出" width="80" />
          <el-table-column prop="reasoningTokens" label="推理" width="80" />
          <el-table-column prop="totalTokens" label="总计" min-width="100">
            <template #default="{ row }">
              {{ (row.inputTokens || 0) + (row.outputTokens || 0) + (row.reasoningTokens || 0) }}
            </template>
          </el-table-column>
          <el-table-column prop="cost" label="扣费" min-width="100" >
            <template #default="{ row }">
              {{ row.cost }} 积分 
            </template>
          </el-table-column>
          <el-table-column prop="latencyMs" label="用时(秒)" width="90" >
            <template #default="{ row }">
              {{ (row.latencyMs * 0.001).toFixed(2) }}
            </template>
          </el-table-column>
          <el-table-column label="状态" width="80">
            <template #default="{ row }">
              <el-tag :type="row.failed ? 'danger' : 'success'" size="small">{{ row.failed ? '失败' : '成功' }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="requestedAt" label="请求时间" width="210">
            <template #default="{ row }">{{ formatDate(row.requestedAt) }}</template>
          </el-table-column>
        </el-table>
      </div>

      <div class="flex justify-center mt-4">
        <el-pagination v-model:current-page="page" :page-size="pageSize" :total="total" @current-change="fetchUsage" layout="prev, pager, next" />
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import api from '@/scripts/api'
import { ElMessage } from 'element-plus'
import dayjs from 'dayjs'

const records = ref<any[]>([])
const loading = ref(false)
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const account = ref('')
const startDate = ref(dayjs().format('YYYY-MM-DD'))
const endDate = ref(dayjs().format('YYYY-MM-DD'))

function handleQuery() {
  page.value = 1
  fetchUsage()
}

async function fetchUsage() {
  loading.value = true
  try {
    const params: any = { page: page.value, pageSize: pageSize.value }
    if (account.value) params.account = account.value
    if (startDate.value) params.start = startDate.value
    if (endDate.value) params.end = endDate.value
    const { data } = await api.get('/v0/admin/usage', { params })
    records.value = data.records || []
    total.value = data.total || 0
  } catch { ElMessage.error('获取用量失败') } finally { loading.value = false }
}

function formatDate(date: string) { return date ? dayjs(date).format('YYYY-MM-DD HH:mm:ss') : '-' }

onMounted(fetchUsage)
</script>
