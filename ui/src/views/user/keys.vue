<template>
  <div class="page-shell space-y-6">
    <section class="page-hero">
      <div>
        <p class="page-kicker">Keys</p>
        <h2 class="page-title">API Keys</h2>
        <p class="page-desc">集中管理访问密钥，快速查看状态、创建时间和最近一次使用情况。</p>
      </div>
      <el-button type="primary" class="page-primary-btn" @click="showCreate = true">
        <el-icon><Plus /></el-icon>
        <span class="hidden sm:inline">创建 Key</span>
        <span class="sm:hidden">创建</span>
      </el-button>
    </section>

    <section class="grid gap-4 sm:grid-cols-3">
      <article class="summary-card">
        <p class="summary-label">总 Key 数</p>
        <p class="summary-value">{{ totalKeys }}</p>
        <p class="summary-desc">当前账号下已创建的所有访问凭证。</p>
      </article>
      <article class="summary-card">
        <p class="summary-label">正常状态</p>
        <p class="summary-value summary-value-success">{{ activeKeys }}</p>
        <p class="summary-desc">可正常使用中的 Key 数量。</p>
      </article>
      <article class="summary-card">
        <p class="summary-label">未使用</p>
        <p class="summary-value summary-value-muted">{{ unusedKeys }}</p>
        <p class="summary-desc">创建后还未发生调用的 Key。</p>
      </article>
    </section>

    <el-card shadow="never" class="panel-card">
      <div class="mb-5 flex items-center justify-between gap-4">
        <div>
          <h3 class="text-lg font-semibold text-slate-900">密钥列表</h3>
          <p class="mt-1 text-sm text-slate-500">支持复制和删除，列表信息保持与当前业务逻辑一致。</p>
        </div>
      </div>

      <div class="overflow-x-auto">
        <el-table :data="keys" v-loading="loading" stripe class="data-table">
          <el-table-column prop="name" label="名称" width="160" />
          <el-table-column prop="key" label="Key" min-width="240">
            <template #default="{ row }">
              <div class="flex items-center gap-2">
                <span class="truncate font-mono text-[13px]">{{ row.key }}</span>
                <el-button size="small" text class="copy-btn" @click="copyKey(row.key)">
                  <el-icon><CopyDocument /></el-icon>
                </el-button>
              </div>
            </template>
          </el-table-column>
          <el-table-column label="状态" width="110">
            <template #default="{ row }">
              <el-tag :type="row.status === 1 ? 'success' : 'danger'" size="small" effect="dark" class="status-tag">
                {{ row.status === 1 ? '正常' : '禁用' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="createdAt" label="创建时间" width="210">
            <template #default="{ row }">
              {{ formatDate(row.createdAt) }}
            </template>
          </el-table-column>
          <el-table-column prop="lastUsedAt" label="最后使用" width="210">
            <template #default="{ row }">
              {{ row.lastUsedAt ? formatDate(row.lastUsedAt) : '未使用' }}
            </template>
          </el-table-column>
          <el-table-column label="操作" width="120" fixed="right">
            <template #default="{ row }">
              <el-popconfirm title="确认删除此 Key？" @confirm="handleDelete(row.key)">
                <template #reference>
                  <el-button type="danger" size="small" text class="delete-btn">删除</el-button>
                </template>
              </el-popconfirm>
            </template>
          </el-table-column>
        </el-table>
      </div>
    </el-card>

    <el-dialog v-model="showCreate" title="创建 API Key" width="90%" :style="{ maxWidth: '420px' }" class="user-dialog">
      <el-form :model="createForm" label-width="60px">
        <el-form-item label="名称">
          <el-input v-model="createForm.name" placeholder="输入 Key 名称" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCreate = false">取消</el-button>
        <el-button type="primary" :loading="creating" class="page-primary-btn" @click="handleCreate">创建</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import api from '@/scripts/api'
import { ElMessage } from 'element-plus'
import { Plus, CopyDocument } from '@element-plus/icons-vue'
import dayjs from 'dayjs'

interface Key {
  id: string
  name: string
  key: string
  status: number
  createdAt: string
  lastUsedAt: string
}

const keys = ref<Key[]>([])
const loading = ref(false)
const showCreate = ref(false)
const creating = ref(false)
const createForm = ref({ name: '' })

const totalKeys = computed(() => keys.value.length)
const activeKeys = computed(() => keys.value.filter(item => item.status === 1).length)
const unusedKeys = computed(() => keys.value.filter(item => !item.lastUsedAt || item.lastUsedAt === '0001-01-01T00:00:00Z').length)

async function fetchKeys() {
  loading.value = true
  try {
    const { data } = await api.get('/v0/user/keys')
    keys.value = data.keys || []
  } catch {
    ElMessage.error('获取 Keys 失败')
  } finally {
    loading.value = false
  }
}

async function handleCreate() {
  if (!createForm.value.name) return
  creating.value = true
  try {
    const { data } = await api.post('/v0/user/keys', createForm.value)
    ElMessage.success(`Key 创建成功: ${data.key}`)
    await navigator.clipboard.writeText(data.key).catch(() => {})
    showCreate.value = false
    createForm.value.name = ''
    await fetchKeys()
  } catch {
    ElMessage.error('创建失败')
  } finally {
    creating.value = false
  }
}

async function handleDelete(key: string) {
  try {
    await api.delete(`/v0/user/keys/${key}`)
    ElMessage.success('删除成功')
    await fetchKeys()
  } catch {
    ElMessage.error('删除失败')
  }
}

async function copyKey(key: string) {
  try {
    await navigator.clipboard.writeText(key)
    ElMessage.success('已复制到剪贴板')
  } catch {
    ElMessage.error('复制失败')
  }
}

function formatDate(date: string) {
  if (!date) return '-'
  if (date === '0001-01-01T00:00:00Z') return '-'
  return dayjs(date).format('YYYY-MM-DD HH:mm')
}

onMounted(fetchKeys)
</script>

<style scoped>
.page-shell {
  font-family: 'Avenir Next', 'Segoe UI Variable', 'PingFang SC', 'Hiragino Sans GB', 'Microsoft YaHei', sans-serif;
}

.page-hero,
.summary-card,
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
.panel-card:hover {
  border-color: rgba(14, 165, 233, 0.18);
  box-shadow:
    0 0 0 1px rgba(255, 255, 255, 0.5) inset,
    0 24px 60px rgba(15, 23, 42, 0.08),
    0 0 30px rgba(14, 165, 233, 0.04);
}

.summary-card:hover {
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
  max-width: 40rem;
  color: #475569;
  line-height: 1.85;
}

.summary-card {
  border-radius: 24px;
  padding: 1.35rem;
}

.summary-label {
  margin: 0;
  font-size: 0.8rem;
  letter-spacing: 0.18em;
  text-transform: uppercase;
  color: #64748b;
}

.summary-value {
  margin: 0.6rem 0 0;
  font-size: 2rem;
  font-weight: 600;
  color: #0f172a;
  background: linear-gradient(135deg, #0f172a 40%, #0891b2);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.summary-value-success {
  background: linear-gradient(135deg, #047857 40%, #10b981);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.summary-value-muted {
  background: linear-gradient(135deg, #475569 40%, #94a3b8);
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
  align-self: flex-start;
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

.copy-btn {
  color: #64748b;
  transition: color 0.25s ease, transform 0.25s ease;
}

.copy-btn:hover {
  color: #0891b2;
  transform: scale(1.1);
}

.delete-btn {
  transition: opacity 0.25s ease;
}

.delete-btn:hover {
  opacity: 0.85;
}

.status-tag {
  border-radius: 10px;
  letter-spacing: 0.02em;
}

:deep(.user-dialog .el-dialog) {
  border-radius: 20px;
}

@media (min-width: 640px) {
  .page-hero {
    flex-direction: row;
    align-items: center;
    justify-content: space-between;
    padding: 1.7rem 1.8rem;
  }

  :deep(.panel-card .el-card__body) {
    padding: 1.5rem;
  }
}
</style>
