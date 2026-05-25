<template>
  <div class="landing-page min-h-screen text-slate-100">
    <!-- Animated aurora blobs -->
    <div class="landing-aurora landing-aurora-left"></div>
    <div class="landing-aurora landing-aurora-right"></div>
    <div class="landing-aurora landing-aurora-center"></div>

    <!-- Floating particle dots -->
    <div class="landing-particles">
      <span v-for="n in 18" :key="n" class="landing-particle" :style="particleStyle(n)"></span>
    </div>

    <!-- Nav -->
    <nav class="fixed inset-x-0 top-0 z-50 px-4 pt-4 sm:px-6 lg:px-8">
      <div class="landing-nav-shell mx-auto flex h-16 max-w-7xl items-center justify-between rounded-2xl border border-white/[0.08] bg-slate-950/50 px-4 shadow-[0_24px_80px_rgba(15,23,42,0.35)] backdrop-blur-2xl sm:px-6">
        <div class="flex items-center gap-3">
          <div class="landing-logo-shell h-11 w-11">
            <img src="https://www.dizena.com/logo.png" alt="CodeAPI logo" class="landing-logo-image">
          </div>
          <div>
            <span class="landing-brand-text text-lg font-semibold sm:text-xl">CodeAPI</span>
          </div>
        </div>
        <div class="flex items-center gap-2 sm:gap-3">
          <el-button text class="landing-nav-btn" @click="showLoginDialog = true">登录</el-button>
          <el-button type="primary" class="landing-primary-btn" @click="showRegisterDialog = true">注册</el-button>
        </div>
      </div>
    </nav>

    <!-- Hero -->
    <section class="relative overflow-hidden px-4 pb-20 pt-32 sm:px-6 lg:px-8 lg:pb-24 lg:pt-36">
      <div class="mx-auto grid max-w-7xl items-center gap-12 lg:grid-cols-[minmax(0,1.15fr)_minmax(360px,0.85fr)]">
        <div class="relative z-10">
          <div class="landing-badge mb-6 inline-flex items-center gap-2 rounded-full border border-cyan-400/20 bg-cyan-400/[0.07] px-4 py-2 text-sm text-cyan-100">
            <span class="landing-pulse-dot"></span>
            统一接入多家 AI Provider 的轻量控制台
          </div>

          <h1 class="max-w-4xl text-4xl font-semibold leading-tight text-white sm:text-5xl lg:text-6xl">
            小而美的
            <span class="landing-gradient-text"> AI API 中枢</span>
          </h1>
          <p class="mt-6 max-w-3xl text-base leading-8 text-slate-300/90 sm:text-lg">
            为 AI Coding 工具提供统一的接入、管理和统计界面，将 Key 管理、额度控制、模型可用性和调用统计整合进一个清晰、现代且易维护的站点。
          </p>

          <div class="mt-8 flex flex-col gap-4 sm:flex-row">
            <el-button type="primary" size="large" class="landing-primary-btn landing-cta-btn" @click="showRegisterDialog = true">
              立即开始
              <el-icon class="ml-1"><Right /></el-icon>
            </el-button>
            <el-button size="large" class="landing-secondary-btn" @click="scrollToFeatures">了解核心能力</el-button>
          </div>

          <div class="mt-10 grid gap-4 sm:grid-cols-3">
            <div v-for="stat in stats" :key="stat.label" class="landing-stat-card">
              <p class="text-xs uppercase tracking-[0.24em] text-slate-400">{{ stat.label }}</p>
              <p class="mt-3 text-2xl font-semibold landing-stat-value">{{ stat.value }}</p>
              <p class="mt-2 text-sm text-slate-300/80">{{ stat.desc }}</p>
            </div>
          </div>
        </div>

        <!-- Hero panel -->
        <div class="relative z-10">
          <div class="landing-hero-panel">
            <div class="landing-panel-grid"></div>
            <div class="landing-scanline"></div>
            <div class="flex items-start justify-between gap-4">
              <div>
                <p class="text-xs uppercase tracking-[0.28em] text-cyan-200/60">Control Layer</p>
                <h2 class="mt-3 text-2xl font-semibold text-white">统一代理面板</h2>
              </div>
              <div class="landing-status-pill">
                <span class="landing-status-dot"></span>
                Stable Routing
              </div>
            </div>

            <div class="mt-8 space-y-4">
              <div v-for="signal in heroSignals" :key="signal.title" class="landing-signal-card">
                <div class="landing-signal-icon">
                  <el-icon :size="22" class="text-cyan-200"><component :is="signal.icon" /></el-icon>
                </div>
                <div class="min-w-0 flex-1">
                  <div class="flex items-center justify-between gap-3">
                    <h3 class="text-sm font-medium text-white">{{ signal.title }}</h3>
                    <span class="text-xs text-slate-400/70">{{ signal.meta }}</span>
                  </div>
                  <p class="mt-1 text-sm leading-6 text-slate-300/80">{{ signal.desc }}</p>
                </div>
              </div>
            </div>

            <div class="mt-8 landing-compat-section">
              <p class="text-sm font-medium text-slate-200">兼容接入</p>
              <div class="mt-4 grid grid-cols-2 gap-3 text-sm text-slate-300 sm:grid-cols-4">
                <div v-for="tool in toolBadges" :key="tool" class="landing-tool-badge">
                  {{ tool }}
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- Features -->
    <section ref="featuresRef" class="relative px-4 py-20 sm:px-6 lg:px-8">
      <div class="mx-auto max-w-7xl">
        <div class="mb-12 max-w-3xl">
          <p class="text-sm uppercase tracking-[0.28em] text-cyan-200/60">Capabilities</p>
          <h2 class="mt-4 text-3xl font-semibold text-white sm:text-4xl">为 AI 接入做减法，为稳定运营做加法</h2>
          <p class="mt-4 text-base leading-8 text-slate-300/85">
            页面不堆砌复杂概念，重点把日常最常用的能力放在顺手的位置，让接入、管理、统计都更直接。
          </p>
        </div>

        <div class="grid gap-6 md:grid-cols-2 xl:grid-cols-3">
          <article v-for="feature in features" :key="feature.title" class="landing-feature-card">
            <div class="landing-feature-icon-shell">
              <el-icon :size="24" class="text-cyan-100"><component :is="feature.icon" /></el-icon>
            </div>
            <h3 class="mt-6 text-xl font-semibold text-white">{{ feature.title }}</h3>
            <p class="mt-3 text-sm leading-7 text-slate-300/80">{{ feature.desc }}</p>
          </article>
        </div>
      </div>
    </section>

    <!-- Integrations -->
    <section class="relative px-4 pb-24 sm:px-6 lg:px-8">
      <div class="landing-integration-shell">
        <div class="mb-10 flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
          <div class="max-w-3xl">
            <p class="text-sm uppercase tracking-[0.28em] text-cyan-200/60">Quick Start</p>
            <h2 class="mt-4 text-3xl font-semibold text-white">保持原有调用习惯，几乎零成本切换</h2>
            <p class="mt-4 text-base leading-8 text-slate-300/85">
              你只需要替换 `base_url` 或环境变量，就能把已有工具链接到 CodeAPI 站点上，不必重写已有工作流。
            </p>
          </div>
          <div class="landing-compat-badge">
            兼容 CLI、SDK 与流式响应
          </div>
        </div>

        <div class="grid gap-6 lg:grid-cols-2">
          <article v-for="item in integrations" :key="item.name" class="landing-integration-card">
            <div class="flex items-start justify-between gap-4">
              <div>
                <h3 class="text-xl font-semibold text-white">{{ item.name }}</h3>
                <p class="mt-2 text-sm leading-7 text-slate-300/80">{{ item.desc }}</p>
              </div>
              <div class="landing-ready-pill">
                <span class="landing-ready-dot"></span>
                Ready
              </div>
            </div>
            <div class="landing-code-block">
              <div class="landing-code-dots">
                <span></span><span></span><span></span>
              </div>
              <pre class="font-mono text-sm leading-7 text-cyan-200">{{ item.code }}</pre>
            </div>
          </article>
        </div>
      </div>
    </section>

    <!-- Footer -->
    <footer class="landing-footer px-4 py-10 sm:px-6 lg:px-8">
      <div class="mx-auto flex max-w-7xl flex-col gap-5 sm:flex-row sm:items-center sm:justify-between">
        <div class="flex items-center gap-3">
          <div class="landing-logo-shell h-10 w-10">
            <img src="https://www.dizena.com/logo.png" alt="CodeAPI logo" class="landing-logo-image">
          </div>
          <div>
            <p class="text-sm font-medium text-white">CodeAPI</p>
            <p class="text-sm text-slate-400/80">小而美的 AI 接入站点</p>
          </div>
        </div>
        <div class="text-sm text-slate-400/70 sm:text-right">
          <p>© 2026 CodeAPI. All rights reserved.</p>
          <p class="mt-1">CodeAPI</p>
        </div>
      </div>
    </footer>

    <AuthDialog v-model="showLoginDialog" mode="login" />
    <AuthDialog v-model="showRegisterDialog" mode="register" />
  </div>
</template>

<script setup lang="ts">
import { markRaw, ref } from 'vue'
import { Promotion, Key, DataAnalysis, Connection, Document, Check, Right } from '@element-plus/icons-vue'
import AuthDialog from '@/views/components/AuthDialog.vue'

const showLoginDialog = ref(false)
const showRegisterDialog = ref(false)
const featuresRef = ref<HTMLElement | null>(null)

const stats = [
  { label: 'Unified Access', value: '4+', desc: '主流模型生态统一接入，减少客户端改造成本。' },
  { label: 'Fast Setup', value: '分钟级', desc: '保留原始调用方式，只需替换代理地址即可。' },
  { label: 'Visibility', value: '实时', desc: 'Key、模型、用量与额度状态清晰可查。' },
]

const heroSignals = [
  { icon: markRaw(Connection), title: '多 Provider 汇聚', meta: 'Proxy Layer', desc: '统一代理大模型入口，方便团队集中维护。' },
  { icon: markRaw(DataAnalysis), title: '调用可视化', meta: 'Insights', desc: '把请求记录、Token 消耗、费用变化整理成清晰数据视图。' },
  { icon: markRaw(Key), title: '安全的 Key 管理', meta: 'Security', desc: '创建、复制、删除与脱敏展示都集中在一个轻量控制面板中。' },
]

const toolBadges = ['Claude Code', 'Codex', 'OpenAI SDK', 'Gemini CLI']

const features = [
  { icon: markRaw(Key), title: 'API Key 管理', desc: '集中创建与维护访问密钥，支持多 Key 协同使用，减少分散管理带来的混乱。' },
  { icon: markRaw(DataAnalysis), title: '调用统计', desc: '按时间查看请求记录、输入输出 Token 与耗时，快速定位成本和稳定性变化。' },
  { icon: markRaw(Connection), title: '多模型支持', desc: '兼容 Claude Code、Codex 等常见 AI Coding 工具与主流模型接口。' },
  { icon: markRaw(Document), title: '额度感知', desc: '自动扣费、余额可见、异常更早暴露，让日常使用更可控。' },
  { icon: markRaw(Promotion), title: '高可用代理', desc: '支持多 provider 负载与故障切换，尽量把调用稳定性留在站点层处理。' },
  { icon: markRaw(Check), title: '安全可靠', desc: '敏感信息尽量少暴露，配合更清楚的状态反馈，让站点既轻又稳。' },
]

const URL_PREFIX = 'https://your-domain/v1'

const integrations = [
  {
    name: 'Codex CLI',
    desc: '替换 `CHATGPT_BASE_URL` 即可保持原命令使用方式，适合快速切到统一代理入口。',
    code: `export CHATGPT_BASE_URL="${URL_PREFIX}"
codex`,
  },
  {
    name: 'Claude Code',
    desc: '设置代理地址和 API Key 后即可使用，适合已有 Claude Code 工作流直接接入。',
    code: `export ANTHROPIC_BASE_URL="${URL_PREFIX}"
export ANTHROPIC_API_KEY="sk-your-key"
claude`,
  },
  {
    name: 'OpenAI SDK',
    desc: '代码结构不变，只替换 `base_url` 与密钥来源，尽量减少已有项目的迁移动作。',
    code: `from openai import OpenAI
client = OpenAI(
  base_url="${URL_PREFIX}",
  api_key="sk-your-key"
)`,
  },
  {
    name: 'Gemini CLI',
    desc: '通过配置代理地址接入站点能力，保留命令行侧的使用手感和响应方式。',
    code: `export GEMINI_BASE_URL="${URL_PREFIX}"
gemini`,
  },
]

function scrollToFeatures() {
  featuresRef.value?.scrollIntoView({ behavior: 'smooth', block: 'start' })
}

/** Stagger particle positions/delays deterministically */
function particleStyle(n: number) {
  const seed = n * 37
  const left = (seed * 7) % 100
  const top = (seed * 13) % 100
  const delay = ((seed * 3) % 8).toFixed(1)
  const duration = (6 + (seed % 6)).toFixed(1)
  const size = 2 + (n % 3)
  return {
    left: `${left}%`,
    top: `${top}%`,
    width: `${size}px`,
    height: `${size}px`,
    animationDelay: `${delay}s`,
    animationDuration: `${duration}s`,
  }
}
</script>

<style scoped>
/* ===== Page background ===== */
.landing-page {
  position: relative;
  overflow: hidden;
  background:
    radial-gradient(ellipse 80% 50% at 20% 10%, rgba(6, 182, 212, 0.18), transparent 55%),
    radial-gradient(ellipse 60% 45% at 85% 15%, rgba(99, 102, 241, 0.16), transparent 50%),
    radial-gradient(ellipse 50% 40% at 50% 80%, rgba(14, 165, 233, 0.08), transparent 50%),
    linear-gradient(180deg, #020617 0%, #0b1224 40%, #0f172a 100%);
  font-family: 'Avenir Next', 'Segoe UI Variable', 'PingFang SC', 'Hiragino Sans GB', 'Microsoft YaHei', sans-serif;
}

/* ===== Aurora blobs (animated) ===== */
.landing-aurora {
  pointer-events: none;
  position: absolute;
  border-radius: 9999px;
  filter: blur(100px);
  will-change: transform, opacity;
}

.landing-aurora-left {
  left: -10rem;
  top: 6rem;
  height: 22rem;
  width: 22rem;
  background: radial-gradient(circle, rgba(34, 211, 238, 0.22), rgba(6, 182, 212, 0.05) 70%, transparent);
  animation: aurora-drift-1 14s ease-in-out infinite alternate;
}

.landing-aurora-right {
  right: -8rem;
  top: 18rem;
  height: 20rem;
  width: 20rem;
  background: radial-gradient(circle, rgba(129, 140, 248, 0.18), rgba(99, 102, 241, 0.04) 70%, transparent);
  animation: aurora-drift-2 16s ease-in-out infinite alternate;
}

.landing-aurora-center {
  left: 30%;
  top: 50%;
  height: 18rem;
  width: 18rem;
  background: radial-gradient(circle, rgba(14, 165, 233, 0.10), transparent 70%);
  animation: aurora-drift-3 20s ease-in-out infinite alternate;
}

@keyframes aurora-drift-1 {
  0%   { transform: translate(0, 0) scale(1); opacity: 0.55; }
  100% { transform: translate(3rem, 2rem) scale(1.15); opacity: 0.75; }
}
@keyframes aurora-drift-2 {
  0%   { transform: translate(0, 0) scale(1); opacity: 0.5; }
  100% { transform: translate(-2rem, 3rem) scale(1.1); opacity: 0.68; }
}
@keyframes aurora-drift-3 {
  0%   { transform: translate(0, 0) scale(1); opacity: 0.3; }
  100% { transform: translate(2rem, -2rem) scale(1.2); opacity: 0.45; }
}

/* ===== Floating particles ===== */
.landing-particles {
  position: absolute;
  inset: 0;
  pointer-events: none;
  z-index: 0;
}

.landing-particle {
  position: absolute;
  border-radius: 50%;
  background: rgba(103, 232, 249, 0.35);
  box-shadow: 0 0 6px rgba(103, 232, 249, 0.25);
  animation: particle-float 8s ease-in-out infinite alternate;
  opacity: 0;
}

@keyframes particle-float {
  0%   { opacity: 0; transform: translateY(0); }
  20%  { opacity: 0.6; }
  80%  { opacity: 0.4; }
  100% { opacity: 0; transform: translateY(-30px); }
}

/* ===== Nav shell ===== */
.landing-nav-shell {
  background:
    linear-gradient(135deg, rgba(15, 23, 42, 0.65), rgba(15, 23, 42, 0.45));
  border-image: linear-gradient(135deg, rgba(34, 211, 238, 0.12), rgba(99, 102, 241, 0.08), rgba(255,255,255,0.06)) 1;
  border-image: none;
  border-color: rgba(255, 255, 255, 0.08);
  box-shadow:
    0 0 0 1px rgba(255, 255, 255, 0.03) inset,
    0 24px 80px rgba(15, 23, 42, 0.35);
}

.landing-brand-text {
  background: linear-gradient(135deg, #fff 40%, rgba(103, 232, 249, 0.85));
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

/* ===== Logo ===== */
.landing-logo-shell {
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  border-radius: 1rem;
  background: rgba(255, 255, 255, 0.96);
  box-shadow:
    0 0 20px rgba(34, 211, 238, 0.18),
    0 12px 30px rgba(14, 165, 233, 0.22);
  transition: box-shadow 0.4s ease;
}

.landing-logo-shell:hover {
  box-shadow:
    0 0 28px rgba(34, 211, 238, 0.28),
    0 14px 36px rgba(14, 165, 233, 0.3);
}

.landing-logo-image {
  height: 100%;
  width: 100%;
  object-fit: cover;
}

/* ===== Pulse dot in badge ===== */
.landing-pulse-dot {
  display: inline-block;
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: #67e8f9;
  box-shadow: 0 0 12px rgba(103, 232, 249, 0.8);
  animation: pulse-glow 2.4s ease-in-out infinite;
}

@keyframes pulse-glow {
  0%, 100% { box-shadow: 0 0 8px rgba(103, 232, 249, 0.6); opacity: 0.85; }
  50%      { box-shadow: 0 0 20px rgba(103, 232, 249, 1); opacity: 1; }
}

/* ===== Gradient heading text ===== */
.landing-gradient-text {
  background: linear-gradient(120deg, #67e8f9 0%, #7dd3fc 30%, #a5b4fc 65%, #c4b5fd 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

/* ===== Shared card glass style ===== */
.landing-hero-panel,
.landing-feature-card,
.landing-integration-card,
.landing-stat-card {
  position: relative;
  overflow: hidden;
  border: 1px solid rgba(255, 255, 255, 0.07);
  background: linear-gradient(160deg, rgba(15, 23, 42, 0.7), rgba(15, 23, 42, 0.45));
  box-shadow:
    0 0 0 1px rgba(255, 255, 255, 0.025) inset,
    0 30px 80px rgba(15, 23, 42, 0.32);
  backdrop-filter: blur(28px) saturate(1.2);
  transition: border-color 0.35s ease, box-shadow 0.35s ease, transform 0.35s ease;
}

.landing-hero-panel:hover,
.landing-feature-card:hover,
.landing-integration-card:hover,
.landing-stat-card:hover {
  border-color: rgba(103, 232, 249, 0.15);
  box-shadow:
    0 0 0 1px rgba(103, 232, 249, 0.06) inset,
    0 36px 90px rgba(15, 23, 42, 0.38),
    0 0 40px rgba(34, 211, 238, 0.04);
  transform: translateY(-2px);
}

/* Gradient shimmer overlay on cards */
.landing-feature-card::after,
.landing-integration-card::after,
.landing-stat-card::after,
.landing-hero-panel::after {
  content: '';
  position: absolute;
  inset: 0;
  background:
    linear-gradient(135deg, rgba(34, 211, 238, 0.06), transparent 35%, transparent 65%, rgba(129, 140, 248, 0.06));
  pointer-events: none;
  opacity: 0.7;
  transition: opacity 0.35s ease;
}

.landing-feature-card:hover::after,
.landing-integration-card:hover::after,
.landing-stat-card:hover::after,
.landing-hero-panel:hover::after {
  opacity: 1;
}

/* ===== Hero panel ===== */
.landing-hero-panel {
  border-radius: 32px;
  padding: 1.75rem;
}

.landing-panel-grid {
  position: absolute;
  inset: 0;
  background-image:
    linear-gradient(rgba(148, 163, 184, 0.045) 1px, transparent 1px),
    linear-gradient(90deg, rgba(148, 163, 184, 0.045) 1px, transparent 1px);
  background-size: 28px 28px;
  mask-image: linear-gradient(180deg, rgba(0, 0, 0, 0.7), transparent 85%);
}

/* Scanline sweep effect */
.landing-scanline {
  position: absolute;
  left: 0;
  right: 0;
  height: 80px;
  background: linear-gradient(180deg, transparent, rgba(34, 211, 238, 0.03), transparent);
  animation: scanline-sweep 6s linear infinite;
  pointer-events: none;
  z-index: 1;
}

@keyframes scanline-sweep {
  0%   { top: -80px; }
  100% { top: calc(100% + 80px); }
}

/* Status pill */
.landing-status-pill {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  border-radius: 9999px;
  border: 1px solid rgba(52, 211, 153, 0.2);
  background: rgba(52, 211, 153, 0.08);
  padding: 4px 12px;
  font-size: 0.75rem;
  color: #6ee7b7;
  letter-spacing: 0.02em;
}

.landing-status-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: #34d399;
  box-shadow: 0 0 8px rgba(52, 211, 153, 0.7);
  animation: pulse-glow-green 2.8s ease-in-out infinite;
}

@keyframes pulse-glow-green {
  0%, 100% { box-shadow: 0 0 6px rgba(52, 211, 153, 0.5); opacity: 0.8; }
  50%      { box-shadow: 0 0 14px rgba(52, 211, 153, 0.9); opacity: 1; }
}

/* Signal cards */
.landing-signal-card {
  display: flex;
  gap: 1rem;
  border-radius: 20px;
  border: 1px solid rgba(255, 255, 255, 0.06);
  background: rgba(255, 255, 255, 0.03);
  padding: 1rem;
  transition: background 0.3s ease, border-color 0.3s ease;
}

.landing-signal-card:hover {
  background: rgba(255, 255, 255, 0.05);
  border-color: rgba(103, 232, 249, 0.12);
}

.landing-signal-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 3rem;
  height: 3rem;
  border-radius: 1rem;
  background: linear-gradient(135deg, rgba(34, 211, 238, 0.12), rgba(99, 102, 241, 0.08));
  border: 1px solid rgba(255, 255, 255, 0.08);
  box-shadow: 0 0 20px rgba(34, 211, 238, 0.06);
  flex-shrink: 0;
}

/* Compat section */
.landing-compat-section {
  border-radius: 24px;
  border: 1px solid rgba(255, 255, 255, 0.06);
  background: rgba(2, 6, 23, 0.45);
  padding: 1.25rem;
}

.landing-tool-badge {
  border-radius: 14px;
  border: 1px solid rgba(255, 255, 255, 0.07);
  background: rgba(255, 255, 255, 0.04);
  padding: 0.6rem 0.75rem;
  text-align: center;
  font-size: 0.82rem;
  letter-spacing: 0.01em;
  transition: border-color 0.3s ease, background 0.3s ease, color 0.3s ease;
}

.landing-tool-badge:hover {
  border-color: rgba(103, 232, 249, 0.18);
  background: rgba(34, 211, 238, 0.06);
  color: #e0f2fe;
}

/* ===== Stat cards ===== */
.landing-stat-card {
  border-radius: 24px;
  padding: 1.25rem 1.5rem;
}

.landing-stat-value {
  background: linear-gradient(135deg, #fff 30%, rgba(103, 232, 249, 0.85));
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

/* ===== Feature cards ===== */
.landing-feature-card {
  border-radius: 28px;
  padding: 1.75rem;
}

.landing-feature-icon-shell {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 3.5rem;
  height: 3.5rem;
  border-radius: 1.15rem;
  background: linear-gradient(135deg, rgba(34, 211, 238, 0.2), rgba(99, 102, 241, 0.12));
  border: 1px solid rgba(103, 232, 249, 0.15);
  box-shadow:
    0 0 24px rgba(34, 211, 238, 0.08),
    0 8px 24px rgba(15, 23, 42, 0.25);
  transition: box-shadow 0.35s ease, transform 0.35s ease;
}

.landing-feature-card:hover .landing-feature-icon-shell {
  box-shadow:
    0 0 32px rgba(34, 211, 238, 0.14),
    0 10px 28px rgba(15, 23, 42, 0.3);
  transform: scale(1.05);
}

/* ===== Integration section shell ===== */
.landing-integration-shell {
  border-radius: 32px;
  border: 1px solid rgba(255, 255, 255, 0.07);
  background:
    linear-gradient(160deg, rgba(15, 23, 42, 0.6), rgba(15, 23, 42, 0.35));
  padding: 2.5rem 2rem;
  box-shadow:
    0 0 0 1px rgba(255, 255, 255, 0.025) inset,
    0 30px 80px rgba(15, 23, 42, 0.28);
  backdrop-filter: blur(24px) saturate(1.15);
}

.landing-compat-badge {
  display: inline-flex;
  align-items: center;
  border-radius: 1rem;
  border: 1px solid rgba(255, 255, 255, 0.08);
  background: rgba(15, 23, 42, 0.5);
  padding: 0.55rem 1rem;
  font-size: 0.82rem;
  color: #94a3b8;
  white-space: nowrap;
}

.landing-integration-card {
  border-radius: 28px;
  padding: 1.75rem;
}

/* Ready pill */
.landing-ready-pill {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  border-radius: 9999px;
  border: 1px solid rgba(34, 211, 238, 0.15);
  background: rgba(34, 211, 238, 0.07);
  padding: 4px 12px;
  font-size: 0.75rem;
  color: #a5f3fc;
  white-space: nowrap;
  flex-shrink: 0;
}

.landing-ready-dot {
  width: 5px;
  height: 5px;
  border-radius: 50%;
  background: #22d3ee;
  box-shadow: 0 0 6px rgba(34, 211, 238, 0.6);
}

/* Code block with terminal dots */
.landing-code-block {
  position: relative;
  margin-top: 1.25rem;
  border-radius: 18px;
  border: 1px solid rgba(255, 255, 255, 0.06);
  background: rgba(2, 6, 23, 0.8);
  padding: 1.25rem 1.25rem 1.25rem 1.25rem;
  box-shadow:
    0 2px 16px rgba(0, 0, 0, 0.2) inset,
    0 0 24px rgba(34, 211, 238, 0.03);
  overflow-x: auto;
}

.landing-code-dots {
  display: flex;
  gap: 6px;
  margin-bottom: 12px;
}

.landing-code-dots span {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  background: rgba(148, 163, 184, 0.2);
}

.landing-code-dots span:nth-child(1) { background: rgba(248, 113, 113, 0.55); }
.landing-code-dots span:nth-child(2) { background: rgba(251, 191, 36, 0.5); }
.landing-code-dots span:nth-child(3) { background: rgba(52, 211, 153, 0.5); }

.landing-code-block pre {
  color: #67e8f9;
  font-family: 'JetBrains Mono', 'Fira Code', 'SF Mono', ui-monospace, monospace;
}

/* ===== Buttons ===== */
:deep(.landing-primary-btn.el-button) {
  border: 1px solid rgba(125, 211, 252, 0.25);
  background: linear-gradient(135deg, #0891b2, #2563eb);
  box-shadow:
    0 0 20px rgba(37, 99, 235, 0.18),
    0 16px 36px rgba(37, 99, 235, 0.25);
  transition: all 0.35s ease;
}

:deep(.landing-primary-btn.el-button:hover) {
  border-color: rgba(186, 230, 253, 0.4);
  background: linear-gradient(135deg, #06b6d4, #3b82f6);
  box-shadow:
    0 0 32px rgba(37, 99, 235, 0.28),
    0 20px 44px rgba(37, 99, 235, 0.32);
  transform: translateY(-1px);
}

:deep(.landing-secondary-btn.el-button),
:deep(.landing-nav-btn.el-button) {
  border: 1px solid rgba(255, 255, 255, 0.08);
  background: rgba(15, 23, 42, 0.4);
  color: #cbd5e1;
  transition: all 0.3s ease;
}

:deep(.landing-secondary-btn.el-button:hover),
:deep(.landing-nav-btn.el-button:hover) {
  color: #fff;
  border-color: rgba(103, 232, 249, 0.2);
  background: rgba(30, 41, 59, 0.65);
  box-shadow: 0 0 20px rgba(34, 211, 238, 0.06);
}

:deep(.landing-cta-btn.el-button) {
  min-width: 148px;
}

/* ===== Footer ===== */
.landing-footer {
  border-top: 1px solid rgba(255, 255, 255, 0.06);
}

/* ===== Responsive ===== */
@media (max-width: 767px) {
  .landing-hero-panel,
  .landing-feature-card,
  .landing-integration-card,
  .landing-stat-card {
    border-radius: 22px;
  }

  .landing-hero-panel {
    padding: 1.25rem;
  }

  .landing-integration-shell {
    padding: 1.5rem 1rem;
    border-radius: 24px;
  }
}

/* ===== Reduced motion ===== */
@media (prefers-reduced-motion: reduce) {
  .landing-aurora,
  .landing-particle,
  .landing-scanline,
  .landing-pulse-dot,
  .landing-status-dot {
    animation: none !important;
  }
}
</style>
