<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { chat, listProjects, createProject, getBlackboard, writeBlackboard } from './api/athena'

interface Project {
  id: string
  name: string
  status: string
  original_requirement: string
  created_at: string
}

interface BlackboardEntry {
  id: string
  category: string
  content: string
  certainty: string
  author: string
  created_at: string
}

const projects = ref<Project[]>([])
const messages = ref<{ role: string; content: string }[]>([])
const inputMessage = ref('')
const selectedProject = ref<string | null>(null)
const blackboardEntries = ref<BlackboardEntry[]>([])
const loading = ref(false)

onMounted(() => {
  fetchProjects()
})

async function fetchProjects() {
  try {
    const res = await listProjects()
    projects.value = res.data.projects || []
  } catch (e) {
    console.error('Failed to fetch projects:', e)
  }
}

async function sendMessage() {
  if (!inputMessage.value.trim()) return

  const msg = inputMessage.value.trim()
  messages.value.push({ role: 'user', content: msg })
  inputMessage.value = ''
  loading.value = true

  try {
    const res = await chat(msg)
    messages.value.push({ role: 'assistant', content: res.data.response })
    // Refresh projects list after chat
    await fetchProjects()
  } catch (e: any) {
    messages.value.push({ role: 'assistant', content: `Error: ${e.message}` })
  } finally {
    loading.value = false
  }
}

async function selectProject(id: string) {
  selectedProject.value = id
  try {
    const res = await getBlackboard(id)
    blackboardEntries.value = res.data.entries || []
  } catch (e) {
    blackboardEntries.value = []
  }
}

function certaintyColor(c: string) {
  switch (c) {
    case 'certain': return '#22c55e'
    case 'conjecture': return '#f59e0b'
    case 'pending_verification': return '#94a3b8'
    default: return '#6b7280'
  }
}

function categoryLabel(c: string) {
  const labels: Record<string, string> = {
    goal: '目标',
    fact: '事实',
    discovery: '发现',
    decision: '决策',
    progress: '进展',
    resolution: '决议',
    auxiliary: '辅助',
  }
  return labels[c] || c
}
</script>

<template>
  <div class="app">
    <!-- Header -->
    <header class="header">
      <h1>Athena</h1>
      <span class="subtitle">AI Agent 公司化编排系统</span>
    </header>

    <div class="layout">
      <!-- Sidebar: Projects -->
      <aside class="sidebar">
        <h3>项目列表</h3>
        <div
          v-for="p in projects"
          :key="p.id"
          class="project-item"
          :class="{ active: selectedProject === p.id }"
          @click="selectProject(p.id)"
        >
          <div class="project-name">{{ p.name }}</div>
          <div class="project-status" :class="p.status">{{ p.status }}</div>
        </div>
        <div v-if="projects.length === 0" class="empty">
          暂无项目，在下方输入需求创建
        </div>
      </aside>

      <!-- Main content -->
      <main class="main">
        <!-- Chat area -->
        <div class="chat-area">
          <div class="messages" ref="messagesEl">
            <div v-if="messages.length === 0" class="welcome">
              <p>输入你的项目需求，Athena 会自动创建项目、招聘 Agent、分配任务。</p>
              <p>例如："帮我开发一个图书管理系统，支持用户登录、书籍搜索和借阅"</p>
            </div>
            <div
              v-for="(m, i) in messages"
              :key="i"
              class="message"
              :class="m.role"
            >
              <div class="message-label">{{ m.role === 'user' ? 'CEO' : 'Athena' }}</div>
              <div class="message-content">{{ m.content }}</div>
            </div>
            <div v-if="loading" class="message assistant">
              <div class="message-content loading">思考中...</div>
            </div>
          </div>

          <!-- Input -->
          <div class="input-area">
            <input
              v-model="inputMessage"
              @keydown.enter="sendMessage"
              placeholder="输入项目需求或指令..."
              :disabled="loading"
            />
            <button @click="sendMessage" :disabled="loading || !inputMessage.trim()">
              发送
            </button>
          </div>
        </div>

        <!-- Blackboard panel (shown when project selected) -->
        <div v-if="selectedProject" class="blackboard-panel">
          <h3>黑板 - {{ selectedProject }}</h3>
          <div class="entries">
            <div v-for="e in blackboardEntries" :key="e.id" class="entry">
              <span class="entry-category" :style="{ borderColor: certaintyColor(e.certainty) }">
                {{ categoryLabel(e.category) }}
              </span>
              <span class="entry-certainty" :style="{ color: certaintyColor(e.certainty) }">
                {{ e.certainty }}
              </span>
              <div class="entry-content">{{ e.content }}</div>
              <div class="entry-meta">
                {{ e.author }} · {{ new Date(e.created_at).toLocaleString() }}
              </div>
            </div>
            <div v-if="blackboardEntries.length === 0" class="empty">
              黑板暂无条目
            </div>
          </div>
        </div>
      </main>
    </div>
  </div>
</template>

<style scoped>
.app {
  height: 100vh;
  display: flex;
  flex-direction: column;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
}

.header {
  padding: 12px 20px;
  background: #1a1a2e;
  color: white;
  display: flex;
  align-items: baseline;
  gap: 12px;
}
.header h1 { margin: 0; font-size: 20px; }
.subtitle { color: #94a3b8; font-size: 13px; }

.layout {
  flex: 1;
  display: flex;
  overflow: hidden;
}

.sidebar {
  width: 240px;
  background: #f8fafc;
  border-right: 1px solid #e2e8f0;
  padding: 16px;
  overflow-y: auto;
}
.sidebar h3 { margin: 0 0 12px; font-size: 14px; color: #64748b; }

.project-item {
  padding: 10px 12px;
  border-radius: 6px;
  cursor: pointer;
  margin-bottom: 4px;
}
.project-item:hover { background: #e2e8f0; }
.project-item.active { background: #dbeafe; border-left: 3px solid #3b82f6; }
.project-name { font-size: 13px; font-weight: 500; }
.project-status {
  font-size: 11px;
  margin-top: 2px;
  color: #94a3b8;
}
.project-status.active { color: #22c55e; }

.main {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.chat-area {
  flex: 1;
  display: flex;
  flex-direction: column;
}

.messages {
  flex: 1;
  overflow-y: auto;
  padding: 16px;
}

.welcome {
  text-align: center;
  color: #94a3b8;
  margin-top: 40px;
}
.welcome p { margin: 8px 0; font-size: 14px; }

.message {
  margin-bottom: 12px;
  max-width: 80%;
}
.message.user {
  margin-left: auto;
}
.message.user .message-label { text-align: right; color: #3b82f6; }
.message.assistant .message-label { color: #64748b; }

.message-label {
  font-size: 11px;
  margin-bottom: 2px;
}

.message-content {
  padding: 10px 14px;
  border-radius: 10px;
  font-size: 14px;
  line-height: 1.5;
  white-space: pre-wrap;
}
.message.user .message-content {
  background: #3b82f6;
  color: white;
}
.message.assistant .message-content {
  background: #f1f5f9;
  color: #1e293b;
}

.loading { color: #94a3b8; font-style: italic; }

.input-area {
  padding: 12px 16px;
  border-top: 1px solid #e2e8f0;
  display: flex;
  gap: 8px;
}
.input-area input {
  flex: 1;
  padding: 10px 14px;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  font-size: 14px;
  outline: none;
}
.input-area input:focus { border-color: #3b82f6; }
.input-area button {
  padding: 10px 20px;
  background: #3b82f6;
  color: white;
  border: none;
  border-radius: 8px;
  cursor: pointer;
  font-size: 14px;
}
.input-area button:disabled { opacity: 0.5; cursor: not-allowed; }

.blackboard-panel {
  border-top: 1px solid #e2e8f0;
  max-height: 300px;
  overflow-y: auto;
  padding: 16px;
  background: #f8fafc;
}
.blackboard-panel h3 {
  margin: 0 0 12px;
  font-size: 14px;
  color: #64748b;
}

.entry {
  padding: 8px 12px;
  margin-bottom: 6px;
  background: white;
  border-radius: 6px;
  border: 1px solid #e2e8f0;
}
.entry-category {
  font-size: 11px;
  font-weight: 600;
  padding: 1px 6px;
  border-left: 3px solid;
  border-radius: 2px;
  margin-right: 8px;
}
.entry-certainty {
  font-size: 11px;
  font-weight: 500;
}
.entry-content {
  font-size: 13px;
  margin-top: 4px;
  line-height: 1.4;
}
.entry-meta {
  font-size: 11px;
  color: #94a3b8;
  margin-top: 4px;
}

.empty {
  text-align: center;
  color: #94a3b8;
  font-size: 13px;
  padding: 20px;
}
</style>
