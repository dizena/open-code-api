<template>
  <div class="page-shell space-y-6">
    <section class="page-hero">
      <div>
        <p class="page-kicker">Usage</p>
        <h2 class="page-title">用量统计</h2>
        <p class="page-desc">按日期范围查看请求记录、Token 消耗、延迟和费用，保留原有查询与分页逻辑。</p>
      </div>
      <div class="summary-pill">
        <span class="summary-pill-label">总记录数</span>
        <span class="summary-pill-value">{{ total }}</span>
      </div>
    </section>

    <section class="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
      <article class="summary-card">
        <p class="summary-label">当前页请求</p>
        <p class="summary-value">{{ records.length }}</p>
        <p class="summary-desc">当前分页内拉取到的调用记录数。</p>
      </article>
      <article class="summary-card">
        <p class="summary-label">Token 总量</p>
        <p class="summary-value summary-value-accent">{{ totalTokens }}</p>
        <p class="summary-desc">输入、输出与推理 Token 的合计值。</p>
      </article>
      <article class="summary-card">
        <p class="summary-label">当前页费用</p>
        <p class="summary-value summary-value-cost">{{ totalCost }} <span class="text-sm font-normal">积分</span></p>
        <p class="summary-desc">按当前页数据汇总后的扣费金额。</p>
      </article>
      <article class="summary-card">
        <p class="summary-label">平均耗时</p>
        <p class="summary-value summary-value-time">{{ avgLatencySeconds }}</p>
        <p class="summary-desc">基于当前页记录估算的平均响应时间。</p>
      </article>
    </section>

    <el-card shadow="never" class="panel-card">
      <div class="mb-5">
        <h3 class="text-lg font-semibold text-slate-900">筛选条件</h3>
        <p class="mt-1 text-sm text-slate-500">按开始与结束日期过滤数据，查询行为保持原样。</p>
      </div>

      <el-form :inline="true" class="flex flex-wrap gap-2">
        <el-form-item label="开始日期" class="w-full sm:w-auto">
          <el-date-picker v-model="startDate" type="date" value-format="YYYY-MM-DD" placeholder="选择日期" class="w-full sm:w-44" />
        </el-form-item>
        <el-form-item label="结束日期" class="w-full sm:w-auto">
          <el-date-picker v-model="endDate" type="date" value-format="YYYY-MM-DD" placeholder="选择日期" class="w-full sm:w-44" />
        </el-form-item>
        <el-form-item class="w-full sm:w-auto">
          <el-button type="primary" class="page-primary-btn w-full sm:w-auto" @click="fetchUsage">查询</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card shadow="never" class="panel-card">
      <div class="mb-5 flex items-center justify-between gap-4">
        <div>
          <h3 class="text-lg font-semibold text-slate-900">调用明细</h3>
          <p class="mt-1 text-sm text-slate-500">表格字段和分页规则与原有功能保持一致。</p>
        </div>
      </div>

      <div class="overflow-x-auto">
        <el-table :data="records" v-loading="loading" stripe class="data-table">
          <el-table-column prop="apiKey" label="API Key" min-width="120" />
          <el-table-column prop="alias" label="模型" min-width="130" />
          <el-table-column prop="inputTokens" label="输入" width="90" />
          <el-table-column prop="outputTokens" label="输出" width="90" />
          <el-table-column prop="reasoningTokens" label="推理" width="90" />
          <el-table-column prop="cost" label="扣费" width="100">
            <template #default="{ row }">
              <span class="cost-cell">{{ row.cost }} 积分</span>
            </template>
          </el-table-column>
          <el-table-column prop="latencyMs" label="用时(秒)" width="100">
            <template #default="{ row }">
              {{ (row.latencyMs * 0.001).toFixed(2) }}
            </template>
          </el-table-column>
          <el-table-column prop="requestedAt" label="请求时间" width="210">
            <template #default="{ row }">
              {{ formatDate(row.requestedAt) }}
            </template>
          </el-table-column>
        </el-table>
      </div>

      <div class="mt-5 flex justify-center">
        <el-pagination
          v-model:current-page="page"
          :page-size="pageSize"
          :total="total"
          @current-change="fetchUsage"
          layout="prev, pager, next"
          class="usage-pagination"
        />
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import api from '@/scripts/api'
import { ElMessage } from 'element-plus'
import dayjs from 'dayjs'

interface Record {
  apiKey: string
  alias: string
  inputTokens: number
  outputTokens: number
  reasoningTokens: number
  cost: number
  latencyMs: number
  requestedAt: string
}

const records = ref<Record[]>([])
const loading = ref(false)
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const startDate = ref(dayjs().format('YYYY-MM-DD'))
const endDate = ref(dayjs().format('YYYY-MM-DD'))

const totalTokens = computed(() =>
  records.value.reduce((sum, item) => sum + item.inputTokens + item.outputTokens + item.reasoningTokens, 0),
)

const totalCost = computed(() =>
  records.value.reduce((sum, item) => sum + item.cost, 0),
)

const avgLatencySeconds = computed(() => {
  if (!records.value.length) return '0.00s'
  const totalLatency = records.value.reduce((sum, item) => sum + item.latencyMs, 0)
  return `${(totalLatency / records.value.length / 1000).toFixed(2)}s`
})

async function fetchUsage() {
  loading.value = true
  try {
    const params: any = { page: page.value, pageSize: pageSize.value }
    if (startDate.value) params.start = startDate.value
    if (endDate.value) params.end = endDate.value
    const { data } = await api.get('/v0/user/usage', { params })
    records.value = data.records || []
    total.value = data.total || 0
  } catch {
    ElMessage.error('获取用量失败')
  } finally {
    loading.value = false
  }
}

function formatDate(date: string) {
  return date ? dayjs(date).format('YYYY-MM-DD HH:mm:ss') : '-'
}

onMounted(fetchUsage)
</script>

<style scoped>
.page-shell {
  font-family: 'Avenir Next', 'Segoe UI Variable', 'PingFang SC', 'Hiragino Sans GB', 'Microsoft YaHei', sans-serif;
}

.page-hero,
.summary-card,
.summary-pill,
.panel-card {
  border: 1px solid rgba(148, 163, 184, 0.14);
  background: rgba(255, 255, 255, 0.92);
  box-shadow:
    0 0 0 1px rgba(255, 255, 255, 0.5) inset,
    0 20px 50px rgba(15, 23, 42, 0.06);
  backdrop-filter: blur(20px) saturate(1.15);
  transition: all 0.35s ease;
}

.page-hero:hover,
.summary-card:hover,
.summary-pill:hover,
.panel-card:hover {
  border-color: rgba(14, 165, 233, 0.18);
  box-shadow:
    0 0 0 1px rgba(255, 255, 255, 0.5) inset,
    0 24px 60px rgba(15, 23, 42, 0.08),
    0 0 30px rgba(14, 165, 233, 0.04);
}

.summary-card:hover,
.summary-pill:hover {
  transform: translateY(-2px);
}

.page-hero {
  display: flex;
  flex-direction: column;
  gap: 1rem;
  border-radius: 28px;
  padding: 1.4rem 1.3rem;
}

.page-kicker {
  margin: 0;
  font-size: 0.72rem;
  letter-spacing: 0.28em;
  text-transform: uppercase;
  color: #0891b2;
}

.page-title {
  margin-top: 0.5rem;
  margin-bottom: 0;
  font-size: 1.8rem;
  font-weight: 600;
  color: #0f172a;
}

.page-desc {
  margin-top: 0.85rem;
  margin-bottom: 0;
  max-width: 42rem;
  color: #475569;
  line-height: 1.85;
}

.summary-pill {
  display: inline-flex;
  align-self: flex-start;
  flex-direction: column;
  gap: 0.3rem;
  border-radius: 20px;
  padding: 0.9rem 1rem;
}

.summary-pill-label,
.summary-label {
  font-size: 0.76rem;
  letter-spacing: 0.18em;
  text-transform: uppercase;
  color: #64748b;
}

.summary-pill-value,
.summary-value {
  color: #0f172a;
  font-weight: 600;
}

.summary-pill-value {
  font-size: 1.8rem;
  background: linear-gradient(135deg, #0f172a 40%, #0891b2);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.summary-card {
  border-radius: 24px;
  padding: 1.35rem;
}

.summary-label {
  margin: 0;
}

.summary-value {
  margin: 0.6rem 0 0;
  font-size: 2rem;
  background: linear-gradient(135deg, #0f172a 40%, #0891b2);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.summary-value-accent {
  background: linear-gradient(135deg, #0e7490 40%, #06b6d4);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.summary-value-cost {
  background: linear-gradient(135deg, #b45309 40%, #f59e0b);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.summary-value-time {
  background: linear-gradient(135deg, #4338ca 40%, #6366f1);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.summary-desc {
  margin: 0.65rem 0 0;
  color: #64748b;
  line-height: 1.75;
}

:deep(.panel-card.el-card) {
  border-radius: 28px;
}

:deep(.panel-card .el-card__body) {
  padding: 1.25rem;
}

:deep(.page-primary-btn.el-button) {
  border: 1px solid rgba(125, 211, 252, 0.25);
  background: linear-gradient(135deg, #0891b2, #2563eb);
  box-shadow:
    0 0 18px rgba(37, 99, 235, 0.14),
    0 14px 32px rgba(37, 99, 235, 0.2);
  transition: all 0.35s ease;
}

:deep(.page-primary-btn.el-button:hover) {
  border-color: rgba(186, 230, 253, 0.35);
  background: linear-gradient(135deg, #06b6d4, #3b82f6);
  box-shadow:
    0 0 28px rgba(37, 99, 235, 0.22),
    0 18px 40px rgba(37, 99, 235, 0.28);
  transform: translateY(-1px);
}

:deep(.data-table.el-table) {
  --el-table-header-bg-color: rgba(15, 23, 42, 0.03);
  --el-table-row-hover-bg-color: rgba(14, 165, 233, 0.06);
  --el-table-border-color: rgba(148, 163, 184, 0.12);
  border-radius: 18px;
}

.cost-cell {
  color: #d97706;
  font-weight: 500;
}

:deep(.usage-pagination.el-pagination) {
  --el-pagination-bg-color: transparent;
  --el-pagination-text-color: #64748b;
  --el-pagination-button-bg-color: rgba(255, 255, 255, 0.9);
  --el-pagination-button-color: #475569;
  --el-pagination-hover-color: #0891b2;
}

:deep(.usage-pagination .el-pager li) {
  border-radius: 12px;
  transition: all 0.3s ease;
}

:deep(.usage-pagination .el-pager li:hover) {
  background: rgba(14, 165, 233, 0.08);
  color: #0891b2;
}

:deep(.usage-pagination .el-pager li.is-active) {
  background: linear-gradient(135deg, #0891b2, #06b6d4);
  color: #fff;
  box-shadow: 0 4px 16px rgba(6, 182, 212, 0.3);
}

@media (min-width: 640px) {
  .page-hero {
    flex-direction: row;
    align-items: center;
    justify-content: space-between;
    padding: 1.7rem 1.8rem;
  }

  .summary-pill {
    align-self: center;
    min-width: 160px;
  }

  :deep(.panel-card .el-card__body) {
    padding: 1.5rem;
  }
}
</style>
