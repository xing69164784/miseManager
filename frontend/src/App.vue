<script setup>
import { ref, onMounted } from 'vue'
import ToolList from './components/ToolList.vue'
import ToolDetail from './components/ToolDetail.vue'
import { MiseService } from '../bindings/changeme'

const tools = ref([])
const selectedTool = ref(null)
const loading = ref(true)
const error = ref('')
const miseVersion = ref('')
const showAddToolModal = ref(false)
const addToolQuery = ref('')
const searchingRegistry = ref(false)
const registryResults = ref([])
const installingTool = ref('')
const installProgress = ref([])
const registryLoaded = ref(false)
const showDoctorModal = ref(false)
const doctorResult = ref('')
const doctorLoading = ref(false)
const showTerminalModal = ref(false)
const terminalInput = ref('')
const terminalOutput = ref([])
const terminalRunning = ref(false)
const activeTab = ref('tools')
const logContent = ref('')
const logLoading = ref(false)
let logTimer = null
let searchTimer = null

const preloadRegistry = async () => {
  if (registryLoaded.value) return
  searchingRegistry.value = true
  try {
    await MiseService.SearchRegistry('')
    registryLoaded.value = true
  } catch (err) {
    console.log('Preload registry failed:', err)
  }
  searchingRegistry.value = false
}

const searchRegistry = () => {
  if (searchTimer) clearTimeout(searchTimer)
  if (!addToolQuery.value.trim()) {
    registryResults.value = []
    return
  }
  searchTimer = setTimeout(async () => {
    searchingRegistry.value = true
    try {
      registryResults.value = await MiseService.SearchRegistry(addToolQuery.value.trim())
    } catch (err) {
      console.log('Search failed:', err)
      registryResults.value = []
    }
    searchingRegistry.value = false
  }, 500)
}

const openAddToolModal = () => {
  showAddToolModal.value = true
  preloadRegistry()
}

const installNewTool = async (toolName) => {
  installingTool.value = toolName
  installProgress.value = [{ type: 'info', message: `执行命令: mise install ${toolName}@latest...` }]
  
  const addProgress = (type, message) => {
    installProgress.value = [...installProgress.value, { type, message }]
  }
  
  const timeoutPromise = new Promise((_, reject) => {
    setTimeout(() => {
      reject(new Error('安装超时，请检查网络或mise配置'))
    }, 120000)
  })
  
  try {
    addProgress('info', '正在执行...')
    const installPromise = MiseService.InstallTool(toolName, 'latest')
    const result = await Promise.race([installPromise, timeoutPromise])
    if (result && result.trim()) {
      const lines = result.trim().split('\n')
      for (const line of lines) {
        if (line.trim()) addProgress('info', line.trim())
      }
    }
    addProgress('success', `${toolName} 安装成功！`)
    setTimeout(async () => {
      showAddToolModal.value = false
      addToolQuery.value = ''
      registryResults.value = []
      installProgress.value = []
      installingTool.value = ''
      await loadTools()
    }, 1500)
  } catch (err) {
    addProgress('error', `安装失败: ${err.message || String(err)}`)
  }
  installingTool.value = ''
}

const runDoctor = async () => {
  showDoctorModal.value = true
  doctorLoading.value = true
  doctorResult.value = ''
  try {
    const result = await MiseService.MiseDoctor()
    doctorResult.value = result || '没有输出'
  } catch (err) {
    doctorResult.value = `执行失败: ${err.message || String(err)}`
  }
  doctorLoading.value = false
}

const executeCommand = async () => {
  const cmd = terminalInput.value.trim()
  if (!cmd || terminalRunning.value) return
  terminalRunning.value = true
  terminalOutput.value = [...terminalOutput.value, { type: 'input', text: `$ mise ${cmd}` }]
  terminalInput.value = ''
  try {
    const result = await MiseService.RunCustomCommand(cmd)
    if (result && result.trim()) {
      const lines = result.trim().split('\n')
      for (const line of lines) {
        terminalOutput.value = [...terminalOutput.value, { type: 'output', text: line }]
      }
    } else {
      terminalOutput.value = [...terminalOutput.value, { type: 'output', text: '(无输出)' }]
    }
  } catch (err) {
    terminalOutput.value = [...terminalOutput.value, { type: 'error', text: err.message || String(err) }]
  }
  terminalRunning.value = false
  setTimeout(() => {
    const el = document.querySelector('.terminal-output')
    if (el) el.scrollTop = el.scrollHeight
  }, 50)
}

const loadLogs = async () => {
  logLoading.value = true
  try {
    logContent.value = await MiseService.GetLogs()
  } catch (err) {
    logContent.value = `加载日志失败: ${err.message || String(err)}`
  }
  logLoading.value = false
  setTimeout(() => {
    const el = document.querySelector('.log-content')
    if (el) el.scrollTop = el.scrollHeight
  }, 50)
}

const clearLogs = async () => {
  try {
    await MiseService.ClearLogs()
    logContent.value = ''
  } catch (err) {
    console.log('Clear logs failed:', err)
  }
}

const switchTab = (tab) => {
  activeTab.value = tab
  if (tab === 'logs') {
    loadLogs()
    logTimer = setInterval(loadLogs, 3000)
  } else {
    if (logTimer) {
      clearInterval(logTimer)
      logTimer = null
    }
  }
}

const loadTools = async () => {
  loading.value = true
  error.value = ''
  try {
    tools.value = await MiseService.ListInstalledTools()
  } catch (err) {
    error.value = String(err)
  }
  loading.value = false
}

const selectTool = (tool) => {
  selectedTool.value = tool
}

const refreshTools = () => {
  loadTools()
  selectedTool.value = null
}

const handleToolUpdated = async () => {
  try {
    const toolsList = await MiseService.ListInstalledTools()
    tools.value = toolsList
    if (selectedTool.value) {
      const updated = toolsList.find(t => t.name === selectedTool.value.name)
      if (updated) {
        selectedTool.value = updated
      }
    }
  } catch (err) {
    error.value = String(err)
  }
}

onMounted(async () => {
  await loadTools()
  try {
    miseVersion.value = await MiseService.GetVersion()
  } catch (err) {
    console.log('Failed to get mise version:', err)
  }
})
</script>

<template>
  <div class="app-container">
    <header class="header">
      <div class="header-left">
        <div class="title-row">
          <h1 class="title">Mise 管理器</h1>
          <span v-if="miseVersion" class="version">{{ miseVersion }}</span>
        </div>
      </div>
      <button class="refresh-btn" @click="refreshTools" :disabled="loading" title="刷新工具列表">
        <svg v-if="loading" class="spinner" viewBox="0 0 24 24">
          <circle class="path" cx="12" cy="12" r="10" fill="none" stroke-width="4"></circle>
        </svg>
        <svg v-else class="refresh-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <polyline points="23 4 23 10 17 10"></polyline>
          <polyline points="1 20 1 14 7 14"></polyline>
          <path d="M3.51 9a9 9 0 0 1 14.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0 0 20.49 15"></path>
        </svg>
      </button>
      <button class="add-tool-btn" @click="openAddToolModal" title="安装新工具">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <line x1="12" y1="5" x2="12" y2="19"></line>
          <line x1="5" y1="12" x2="19" y2="12"></line>
        </svg>
      </button>
      <button class="doctor-btn" @click="runDoctor" title="Mise Doctor">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M22 12h-4l-3 9L9 3l-3 9H2"></path>
        </svg>
      </button>
      <button class="terminal-btn" @click="showTerminalModal = true; terminalOutput = []" title="命令终端">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <polyline points="4 17 10 11 4 5"></polyline>
          <line x1="12" y1="19" x2="20" y2="19"></line>
        </svg>
      </button>
    </header>

    <div class="tab-bar">
      <button class="tab-btn" :class="{ active: activeTab === 'tools' }" @click="switchTab('tools')">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="14" height="14">
          <path d="M14.7 6.3a1 1 0 0 0 0 1.4l1.6 1.6a1 1 0 0 0 1.4 0l3.77-3.77a6 6 0 0 1-7.94 7.94l-6.91 6.91a2.12 2.12 0 0 1-3-3l6.91-6.91a6 6 0 0 1 7.94-7.94l-3.76 3.76z"></path>
        </svg>
        工具管理
      </button>
      <button class="tab-btn" :class="{ active: activeTab === 'logs' }" @click="switchTab('logs')">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="14" height="14">
          <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"></path>
          <polyline points="14 2 14 8 20 8"></polyline>
          <line x1="16" y1="13" x2="8" y2="13"></line>
          <line x1="16" y1="17" x2="8" y2="17"></line>
          <polyline points="10 9 9 9 8 9"></polyline>
        </svg>
        日志
      </button>
    </div>

    <div v-if="error" class="error-message">
      {{ error }}
    </div>

    <div v-if="activeTab === 'tools'" class="main-content">
      <ToolList 
        :tools="tools" 
        :loading="loading" 
        :selected-tool="selectedTool"
        @select="selectTool" 
      />
      <ToolDetail 
        :tool="selectedTool" 
        @updated="handleToolUpdated"
      />
    </div>

    <div v-if="activeTab === 'logs'" class="log-page">
      <div class="log-toolbar">
        <span class="log-title">日志记录</span>
        <span class="log-path">~/.mise-gui.log</span>
        <button class="log-refresh-btn" @click="loadLogs" :disabled="logLoading">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="14" height="14">
            <polyline points="23 4 23 10 17 10"></polyline>
            <polyline points="1 20 1 14 7 14"></polyline>
            <path d="M3.51 9a9 9 0 0 1 14.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0 0 20.49 15"></path>
          </svg>
          刷新
        </button>
        <button class="log-clear-btn" @click="clearLogs">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="14" height="14">
            <polyline points="3 6 5 6 21 6"></polyline>
            <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path>
          </svg>
          清空
        </button>
      </div>
      <pre class="log-content">{{ logContent || '暂无日志' }}</pre>
    </div>

    <div v-if="showAddToolModal" class="modal-overlay" @click.self="showAddToolModal = false">
      <div class="add-tool-modal">
        <div class="modal-header">
          <h3>安装新工具</h3>
          <button class="modal-close" @click="showAddToolModal = false">×</button>
        </div>
        <div class="modal-body">
          <div class="search-row">
            <input 
              type="text" 
              v-model="addToolQuery"
              placeholder="搜索工具名称..."
              class="search-input"
              @input="searchRegistry"
              autofocus
            />
          </div>
          <div v-if="searchingRegistry" class="search-loading">
            {{ registryLoaded ? '搜索中...' : '正在加载工具注册表，首次可能需要较长时间...' }}
          </div>
          <div v-else-if="registryResults.length > 0" class="registry-list">
            <div 
              v-for="tool in registryResults" 
              :key="tool.short" 
              class="registry-item"
              :class="{ 'installing': installingTool === tool.short }"
            >
              <div class="registry-info">
                <div class="registry-name">{{ tool.short }}</div>
                <div v-if="tool.description" class="registry-desc">{{ tool.description }}</div>
                <div class="registry-backends">
                  <span v-for="b in tool.backends.slice(0, 2)" :key="b" class="backend-tag">{{ b }}</span>
                </div>
              </div>
              <button 
                class="install-tool-btn" 
                @click="installNewTool(tool.short)"
                :disabled="installingTool !== ''"
              >
                <svg v-if="installingTool === tool.short" class="spinner-mini" viewBox="0 0 24 24">
                  <circle class="path" cx="12" cy="12" r="10" fill="none" stroke-width="4"></circle>
                </svg>
                {{ installingTool === tool.short ? '安装中' : '安装' }}
              </button>
            </div>
          </div>
          <div v-else-if="addToolQuery && !searchingRegistry" class="search-empty">
            没有找到匹配的工具
          </div>
          <div v-else class="search-hint">
            输入关键词搜索可安装的工具
          </div>
          <div v-if="installProgress.length > 0" class="progress-section">
            <div class="progress-logs">
              <div v-for="(log, index) in installProgress" :key="index" class="progress-log">
                <span class="log-icon">{{ log.type === 'success' ? '✓' : log.type === 'error' ? '✗' : '•' }}</span>
                <span class="log-text">{{ log.message }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div v-if="showDoctorModal" class="modal-overlay" @click.self="showDoctorModal = false">
      <div class="doctor-modal">
        <div class="modal-header">
          <h3>Mise Doctor</h3>
          <button class="modal-close" @click="showDoctorModal = false">×</button>
        </div>
        <div class="modal-body">
          <div v-if="doctorLoading" class="doctor-loading">
            <svg class="spinner" viewBox="0 0 24 24">
              <circle class="path" cx="12" cy="12" r="10" fill="none" stroke-width="4"></circle>
            </svg>
            正在诊断...
          </div>
          <pre v-else class="doctor-result">{{ doctorResult }}</pre>
        </div>
        <div class="modal-footer">
          <button class="btn-cancel" @click="showDoctorModal = false">关闭</button>
          <button class="btn-retry" @click="doctorLoading = true; MiseService.MiseDoctor().then(r => { doctorResult = r || '没有输出'; doctorLoading = false; }).catch(e => { doctorResult = '执行失败: ' + (e.message || String(e)); doctorLoading = false; })">重新诊断</button>
        </div>
      </div>
    </div>

    <div v-if="showTerminalModal" class="modal-overlay" @click.self="showTerminalModal = false">
      <div class="terminal-modal">
        <div class="modal-header">
          <h3>Mise 命令终端</h3>
          <button class="modal-close" @click="showTerminalModal = false">×</button>
        </div>
        <div class="terminal-body">
          <div class="terminal-output">
            <div v-for="(line, index) in terminalOutput" :key="index" class="terminal-line" :class="line.type">
              {{ line.text }}
            </div>
            <div v-if="terminalOutput.length === 0" class="terminal-hint">
              输入 mise 子命令执行，例如：ls、doctor、registry python、install go@1.21.0
            </div>
          </div>
          <div class="terminal-input-row">
            <span class="terminal-prompt">$ mise</span>
            <input 
              type="text" 
              v-model="terminalInput"
              class="terminal-input"
              placeholder="输入命令..."
              @keyup.enter="executeCommand"
              :disabled="terminalRunning"
              autofocus
            />
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.app-container {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  background-color: #1e1e1e;
  color: #e4e4e4;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  background-color: #252526;
  border-bottom: 1px solid #3c3c3c;
  white-space: nowrap;
}

.tab-bar {
  display: flex;
  background-color: #252526;
  border-bottom: 1px solid #3c3c3c;
  padding: 0 16px;
  gap: 4px;
}

.tab-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  background: none;
  border: none;
  border-bottom: 2px solid transparent;
  color: #858585;
  cursor: pointer;
  font-size: 13px;
  transition: all 0.2s;
}

.tab-btn:hover {
  color: #e4e4e4;
}

.tab-btn.active {
  color: #e4e4e4;
  border-bottom-color: #569cd6;
}

.log-page {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.log-toolbar {
  display: flex;
  align-items: center;
  padding: 8px 16px;
  background-color: #252526;
  border-bottom: 1px solid #3c3c3c;
  gap: 12px;
}

.log-title {
  font-size: 13px;
  font-weight: 500;
  color: #e4e4e4;
}

.log-path {
  font-size: 12px;
  color: #858585;
  font-family: 'Consolas', 'Courier New', monospace;
}

.log-refresh-btn, .log-clear-btn {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 4px 10px;
  background-color: #3c3c3c;
  border: none;
  border-radius: 4px;
  color: #e4e4e4;
  cursor: pointer;
  font-size: 12px;
  margin-left: auto;
}

.log-refresh-btn:hover, .log-clear-btn:hover {
  background-color: #4c4c4c;
}

.log-clear-btn {
  margin-left: 8px;
}

.log-content {
  flex: 1;
  margin: 0;
  padding: 16px;
  background-color: #0c0c0c;
  color: #e4e4e4;
  font-family: 'Consolas', 'Courier New', monospace;
  font-size: 12px;
  line-height: 1.6;
  white-space: pre-wrap;
  word-wrap: break-word;
  overflow-y: auto;
}

.header-left {
  display: flex;
  align-items: center;
  flex: 1;
  min-width: 0;
}

.title-row {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
  row-gap: 2px;
}

.title {
  font-size: 16px;
  font-weight: 600;
  margin: 0;
  color: #569cd6;
  flex-shrink: 0;
}

.version {
  font-size: 10px;
  color: #858585;
  background-color: #3c3c3c;
  padding: 1px 6px;
  border-radius: 3px;
  flex-shrink: 0;
}

.refresh-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  padding: 6px;
  background-color: #094771;
  border: none;
  border-radius: 4px;
  color: #fff;
  cursor: pointer;
  flex-shrink: 0;
  transition: background-color 0.2s;
}

.refresh-btn:hover:not(:disabled) {
  background-color: #0a5388;
}

.refresh-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.refresh-icon {
  width: 16px;
  height: 16px;
}

.spinner {
  width: 16px;
  height: 16px;
  animation: spin 1s linear infinite;
}

.spinner .path {
  stroke: #fff;
  stroke-linecap: round;
  animation: dash 1.5s ease-in-out infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

@keyframes dash {
  0% { stroke-dasharray: 1, 150; stroke-dashoffset: 0; }
  50% { stroke-dasharray: 90, 150; stroke-dashoffset: -35; }
  100% { stroke-dasharray: 90, 150; stroke-dashoffset: -124; }
}

.error-message {
  background-color: #4e1818;
  color: #ff7b72;
  padding: 12px 24px;
  border-bottom: 1px solid #3c3c3c;
}

.main-content {
  flex: 1;
  display: flex;
  overflow: hidden;
}

.add-tool-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  padding: 6px;
  background-color: #2ea043;
  border: none;
  border-radius: 4px;
  color: #fff;
  cursor: pointer;
  flex-shrink: 0;
  transition: background-color 0.2s;
  margin-left: 8px;
}

.add-tool-btn:hover {
  background-color: #3fb950;
}

.add-tool-btn svg {
  width: 16px;
  height: 16px;
}

.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.add-tool-modal {
  background-color: #252526;
  border: 1px solid #3c3c3c;
  border-radius: 8px;
  width: 560px;
  max-height: 80vh;
  display: flex;
  flex-direction: column;
}

.add-tool-modal .modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  border-bottom: 1px solid #3c3c3c;
}

.add-tool-modal .modal-header h3 {
  margin: 0;
  font-size: 16px;
  color: #e4e4e4;
}

.modal-close {
  background: none;
  border: none;
  color: #858585;
  font-size: 20px;
  cursor: pointer;
  padding: 0 4px;
  line-height: 1;
}

.modal-close:hover {
  color: #e4e4e4;
}

.add-tool-modal .modal-body {
  padding: 16px 20px;
  overflow-y: auto;
  flex: 1;
}

.search-row {
  margin-bottom: 12px;
}

.search-input {
  width: 100%;
  padding: 10px 12px;
  background-color: #1e1e1e;
  border: 1px solid #3c3c3c;
  border-radius: 6px;
  color: #e4e4e4;
  font-size: 14px;
  outline: none;
  box-sizing: border-box;
}

.search-input:focus {
  border-color: #569cd6;
}

.search-input::placeholder {
  color: #858585;
}

.search-loading, .search-empty, .search-hint {
  text-align: center;
  padding: 24px;
  color: #858585;
  font-size: 13px;
}

.registry-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.registry-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 12px;
  background-color: #1e1e1e;
  border: 1px solid #3c3c3c;
  border-radius: 6px;
  transition: border-color 0.2s;
}

.registry-item:hover {
  border-color: #569cd6;
}

.registry-info {
  flex: 1;
  min-width: 0;
}

.registry-name {
  font-size: 14px;
  font-weight: 600;
  color: #569cd6;
}

.registry-desc {
  font-size: 12px;
  color: #858585;
  margin-top: 2px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.registry-backends {
  display: flex;
  gap: 4px;
  margin-top: 4px;
  flex-wrap: wrap;
}

.backend-tag {
  font-size: 10px;
  padding: 1px 4px;
  background-color: #3c3c3c;
  border-radius: 3px;
  color: #858585;
}

.install-tool-btn {
  padding: 6px 12px;
  background-color: #2ea043;
  border: none;
  border-radius: 4px;
  color: #fff;
  font-size: 12px;
  cursor: pointer;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  gap: 4px;
  transition: background-color 0.2s;
}

.install-tool-btn:hover:not(:disabled) {
  background-color: #3fb950;
}

.install-tool-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.spinner-mini {
  width: 12px;
  height: 12px;
  animation: spin 1s linear infinite;
}

.spinner-mini .path {
  stroke: #fff;
  stroke-linecap: round;
  animation: dash 1.5s ease-in-out infinite;
}

.progress-section {
  margin-top: 12px;
  padding: 10px;
  background-color: #1e1e1e;
  border-radius: 6px;
  border: 1px solid #3c3c3c;
}

.progress-logs {
  display: flex;
  flex-direction: column;
  gap: 3px;
}

.progress-log {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
}

.log-icon {
  font-size: 10px;
  width: 12px;
  text-align: center;
  color: #569cd6;
}

.progress-log:last-child .log-icon {
  color: #6a9955;
}

.log-text {
  color: #e4e4e4;
}

.doctor-btn, .terminal-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  padding: 6px;
  border: none;
  border-radius: 4px;
  color: #fff;
  cursor: pointer;
  flex-shrink: 0;
  transition: background-color 0.2s;
  margin-left: 8px;
}

.doctor-btn {
  background-color: #b5842a;
}

.doctor-btn:hover {
  background-color: #d4a030;
}

.terminal-btn {
  background-color: #6e4ec4;
}

.terminal-btn:hover {
  background-color: #8b6ce0;
}

.doctor-btn svg, .terminal-btn svg {
  width: 16px;
  height: 16px;
}

.doctor-modal {
  background-color: #252526;
  border: 1px solid #3c3c3c;
  border-radius: 8px;
  width: 640px;
  max-height: 80vh;
  display: flex;
  flex-direction: column;
}

.doctor-loading {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12px;
  padding: 40px;
  color: #858585;
  font-size: 14px;
}

.doctor-loading .spinner {
  width: 20px;
  height: 20px;
  animation: spin 1s linear infinite;
}

.doctor-loading .spinner .path {
  stroke: #569cd6;
  stroke-linecap: round;
  animation: dash 1.5s ease-in-out infinite;
}

.doctor-result {
  background-color: #1e1e1e;
  border: 1px solid #3c3c3c;
  border-radius: 6px;
  padding: 16px;
  color: #e4e4e4;
  font-family: 'Consolas', 'Courier New', monospace;
  font-size: 13px;
  line-height: 1.6;
  white-space: pre-wrap;
  word-wrap: break-word;
  max-height: 50vh;
  overflow-y: auto;
  margin: 0;
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  padding: 12px 20px;
  border-top: 1px solid #3c3c3c;
}

.btn-cancel {
  padding: 8px 16px;
  background-color: #3c3c3c;
  border: none;
  border-radius: 4px;
  color: #e4e4e4;
  cursor: pointer;
  font-size: 13px;
}

.btn-cancel:hover {
  background-color: #4c4c4c;
}

.btn-retry {
  padding: 8px 16px;
  background-color: #094771;
  border: none;
  border-radius: 4px;
  color: #fff;
  cursor: pointer;
  font-size: 13px;
}

.btn-retry:hover {
  background-color: #0a5388;
}

.terminal-modal {
  background-color: #252526;
  border: 1px solid #3c3c3c;
  border-radius: 8px;
  width: 700px;
  max-height: 80vh;
  display: flex;
  flex-direction: column;
}

.terminal-body {
  display: flex;
  flex-direction: column;
  height: 450px;
}

.terminal-output {
  flex: 1;
  overflow-y: auto;
  padding: 16px;
  background-color: #0c0c0c;
  font-family: 'Consolas', 'Courier New', monospace;
  font-size: 13px;
  line-height: 1.5;
  border-bottom: 1px solid #3c3c3c;
}

.terminal-line {
  white-space: pre-wrap;
  word-wrap: break-word;
}

.terminal-line.input {
  color: #569cd6;
}

.terminal-line.output {
  color: #e4e4e4;
}

.terminal-line.error {
  color: #ff7b72;
}

.terminal-hint {
  color: #858585;
  font-style: italic;
}

.terminal-input-row {
  display: flex;
  align-items: center;
  padding: 8px 16px;
  background-color: #1e1e1e;
  gap: 8px;
}

.terminal-prompt {
  color: #6a9955;
  font-family: 'Consolas', 'Courier New', monospace;
  font-size: 13px;
  flex-shrink: 0;
}

.terminal-input {
  flex: 1;
  background: none;
  border: none;
  color: #e4e4e4;
  font-family: 'Consolas', 'Courier New', monospace;
  font-size: 13px;
  outline: none;
}

.terminal-input::placeholder {
  color: #858585;
}
</style>