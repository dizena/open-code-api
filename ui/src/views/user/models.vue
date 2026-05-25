<template>
  <div class="page-shell space-y-6">
    <section class="page-hero">
      <div>
        <p class="page-kicker">Models</p>
        <h2 class="page-title">可用模型</h2>
        <p class="page-desc">当前账号可调用的模型列表。展示方式更轻盈，但数据来源和业务能力保持不变。</p>
      </div>
      <div class="summary-pill">
        <span class="summary-pill-label">可用数量</span>
        <span class="summary-pill-value">{{ models.length }}</span>
      </div>
    </section>

    <el-card shadow="never" class="panel-card">
      <div class="mb-6 flex items-center justify-between gap-4">
        <div>
          <h3 class="text-lg font-semibold text-slate-900">模型目录</h3>
          <p class="mt-1 text-sm text-slate-500">按照当前接口返回结果实时展示。</p>
        </div>
      </div>

      <div v-loading="loading">
        <div v-if="models.length" class="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
          <article v-for="model in models" :key="model.alias" class="model-card">
            <p class="model-label">Model Alias</p>
            <h4 class="mt-3 break-all font-mono text-base text-slate-950">{{ model.alias }}</h4>
            <div class="mt-3 flex items-center justify-end">
              <el-button size="small" text class="copy-btn" @click="copyModel(model.alias)">
                <el-icon :size="14"><CopyDocument /></el-icon>
                <span class="ml-1 text-xs">复制</span>
              </el-button>
            </div>
          </article>
        </div>
        <div v-else class="empty-state">
          <p class="text-base font-medium text-slate-700">暂无可用模型</p>
          <p class="mt-2 text-sm text-slate-500">当接口返回模型列表后，这里会自动更新。</p>
        </div>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import api from '@/scripts/api'
import { ElMessage } from 'element-plus'
import { CopyDocument } from '@element-plus/icons-vue'

interface ModelItem {
  alias: string
}

const models = ref<ModelItem[]>([])
const loading = ref(false)

async function fetchModels() {
  loading.value = true
  try {
    const { data } = await api.get('/v0/user/models')
    models.value = (data.models || []).map((m: string) => ({ alias: m }))
  } catch {
    ElMessage.error('获取可用模型失败')
  } finally {
    loading.value = false
  }
}

async function copyModel(alias: string) {
  try {
    await navigator.clipboard.writeText(alias)
    ElMessage.success('已复制到剪贴板')
  } catch {
    ElMessage.error('复制失败')
  }
}

onMounted(fetchModels)
</script>

<style scoped>
.page-shell {
  font-family: 'Avenir Next', 'Segoe UI Variable', 'PingFang SC', 'Hiragino Sans GB', 'Microsoft YaHei', sans-serif;
}

.page-hero,
.panel-card,
.model-card,
.summary-pill {
  border: 1px solid rgba(148, 163, 184, 0.14);
  background: rgba(255, 255, 255, 0.92);
  box-shadow:
    0 0 0 1px rgba(255, 255, 255, 0.5) inset,
    0 20px 50px rgba(15, 23, 42, 0.06);
  backdrop-filter: blur(20px) saturate(1.15);
  transition: all 0.35s ease;
}

.page-hero:hover,
.panel-card:hover,
.model-card:hover,
.summary-pill:hover {
  border-color: rgba(14, 165, 233, 0.18);
  box-shadow:
    0 0 0 1px rgba(255, 255, 255, 0.5) inset,
    0 24px 60px rgba(15, 23, 42, 0.08),
    0 0 30px rgba(14, 165, 233, 0.04);
}

.model-card:hover {
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

.summary-pill-label {
  font-size: 0.76rem;
  letter-spacing: 0.18em;
  text-transform: uppercase;
  color: #64748b;
}

.summary-pill-value {
  font-size: 1.8rem;
  font-weight: 600;
  background: linear-gradient(135deg, #0f172a 40%, #0891b2);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

:deep(.panel-card.el-card) {
  border-radius: 28px;
}

:deep(.panel-card .el-card__body) {
  padding: 1.25rem;
}

.model-card {
  border-radius: 22px;
  padding: 1.15rem;
}

.model-label {
  margin: 0;
  font-size: 0.78rem;
  letter-spacing: 0.2em;
  text-transform: uppercase;
  color: #64748b;
}

.copy-btn {
  color: #64748b;
  transition: color 0.25s ease;
}

.copy-btn:hover {
  color: #0891b2;
}

.empty-state {
  border-radius: 24px;
  border: 1px dashed rgba(148, 163, 184, 0.35);
  background: rgba(248, 250, 252, 0.95);
  padding: 3rem 1.5rem;
  text-align: center;
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
