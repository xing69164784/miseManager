<script setup>import { ref, watch, computed } from 'vue';
import { MiseService } from '../../bindings/changeme';
const props = defineProps({
  tool: {
    type: Object,
    default: null
  }
});
const emit = defineEmits(['updated']);
const availableVersions = ref([]);
const isInstalling = ref(false);
const uninstallingVersion = ref('');
const settingDefaultVersion = ref('');
const showInstallModal = ref(false);
const installVersionInput = ref('');
const errorMessage = ref('');
const loadingAvailableVersions = ref(false);
const installProgress = ref([]);
const showVersionDropdown = ref(false);
const loadAvailableVersions = async () => {
  if (!props.tool)
    return;
  loadingAvailableVersions.value = true;
  try {
    availableVersions.value = await MiseService.GetAvailableVersions(props.tool.name);
  }
  catch (err) {
    console.log('Failed to get available versions:', err);
  }
  loadingAvailableVersions.value = false;
};

const refreshToolInfo = async () => {
  await loadAvailableVersions();
  emit('updated');
};
watch(() => props.tool, () => {
  availableVersions.value = [];
  loadAvailableVersions();
}, { immediate: true });
const isDefault = (version) => {
  return props.tool?.default === version || props.tool?.version === version;
};
const handleInstall = async () => {
 if (!installVersionInput.value.trim())
 return;
 if (!props.tool) return;
 
 isInstalling.value = true;
 errorMessage.value = '';
 
 setTimeout(() => {
 installProgress.value = [{ type: 'info', message: `执行命令: mise install ${props.tool.name}@${installVersionInput.value}` }];
 }, 0);
 
 const addProgress = (type, message) => {
 setTimeout(() => {
 installProgress.value = [...installProgress.value, { type, message }];
 }, 0);
 };
 
 const timeoutPromise = new Promise((_, reject) => {
 setTimeout(() => {
 reject(new Error('安装超时，请检查网络或mise配置'));
 }, 60000);
 });
 
 try {
 addProgress('info', '正在执行...');
 
 const installPromise = MiseService.InstallTool(props.tool.name, installVersionInput.value);
 const result = await Promise.race([installPromise, timeoutPromise]);
 
 await new Promise(r => setTimeout(r, 100));
 
 if (result && result.trim()) {
 const lines = result.trim().split('\n');
 for (const line of lines) {
 if (line.trim()) {
 addProgress('info', line.trim());
 }
 }
 }
 
 await new Promise(r => setTimeout(r, 200));
 addProgress('success', `${props.tool.name} ${installVersionInput.value} 安装成功！`);
 
 setTimeout(async () => {
 showInstallModal.value = false;
 installVersionInput.value = '';
 installProgress.value = [];
 await refreshToolInfo();
 }, 1500);
 }
 catch (err) {
 addProgress('error', `安装失败: ${err.message || String(err)}`);
 }
 isInstalling.value = false;
};
const handleUninstall = async (version) => {
  uninstallingVersion.value = version;
  errorMessage.value = '';
  try {
    await MiseService.UninstallTool(props.tool.name, version);
    await refreshToolInfo();
  }
  catch (err) {
    errorMessage.value = String(err);
  }
  uninstallingVersion.value = '';
};
const handleSetDefault = async (version) => {
  settingDefaultVersion.value = version;
  errorMessage.value = '';
  try {
    await MiseService.SetDefaultVersion(props.tool.name, version);
    await refreshToolInfo();
  }
  catch (err) {
    errorMessage.value = String(err);
  }
  settingDefaultVersion.value = '';
};
const filteredVersions = computed(() => {
  const installed = props.tool?.versions || [];
  return availableVersions.value.map(v => ({
    version: v,
    installed: installed.includes(v)
  }));
});

const filteredUninstalledVersions = computed(() => {
  const installed = props.tool?.versions || [];
  const search = installVersionInput.value.toLowerCase();
  return availableVersions.value
    .filter(v => !installed.includes(v) && (search === '' || v.toLowerCase().includes(search)))
    .slice(0, 20);
});

const selectVersion = (version) => {
  installVersionInput.value = version;
  showVersionDropdown.value = false;
};
</script>

<template>
  <div class="tool-detail">
    <div v-if="!tool" class="empty-detail">
      <svg class="empty-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
        <circle cx="12" cy="12" r="10"></circle>
        <polyline points="12 6 12 12 16 14"></polyline>
      </svg>
      <p>选择工具查看详情</p>
    </div>

    <div v-else class="detail-content">
      <div v-if="errorMessage" class="error-banner">
        {{ errorMessage }}
        <button class="error-close" @click="errorMessage = ''">×</button>
      </div>

      <div class="detail-header">
        <div class="header-icon">
          {{ tool.name.charAt(0).toUpperCase() }}
        </div>
        <div class="header-info">
          <div class="header-top-row">
            <h2>{{ tool.name }}</h2>
            <span v-if="tool.installed" class="status installed">已安装</span>
          </div>
          <span v-if="tool.source" class="source">{{ tool.source }}</span>
        </div>
        <button class="install-btn" @click="showInstallModal = true" :disabled="isInstalling">
          <svg class="plus-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <line x1="12" y1="5" x2="12" y2="19"></line>
            <line x1="5" y1="12" x2="19" y2="12"></line>
          </svg>
          安装版本
        </button>
      </div>

      <div class="current-version">
        <span class="label">当前版本</span>
        <span class="value">{{ tool.version || tool.default || '无' }}</span>
      </div>

      <div class="versions-section">
        <h3>已安装版本</h3>
        <div v-if="tool.versions?.length === 0" class="empty-versions">
          <p>没有安装版本</p>
        </div>
        <div v-else class="versions-list">
          <div v-for="version in tool.versions" :key="version" class="version-item">
            <div class="version-info">
              <span class="version-number">{{ version }}</span>
              <span v-if="isDefault(version)" class="default-badge">默认</span>
            </div>
            <div class="version-actions">
              <button v-if="!isDefault(version)" class="action-btn set-default" @click="handleSetDefault(version)"
                :disabled="settingDefaultVersion !== ''">
                <svg v-if="settingDefaultVersion === version" class="spinner" viewBox="0 0 24 24">
                  <circle class="path" cx="12" cy="12" r="10" fill="none" stroke-width="4"></circle>
                </svg>
                <svg v-else viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <path d="M9 12l2 2 4-4"></path>
                </svg>
                设为默认
              </button>
              <button class="action-btn uninstall" @click="handleUninstall(version)"
                :disabled="uninstallingVersion === version">
                <svg v-if="uninstallingVersion === version" class="spinner" viewBox="0 0 24 24">
                  <circle class="path" cx="12" cy="12" r="10" fill="none" stroke-width="4"></circle>
                </svg>
                <svg v-else viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <path d="M3 6h18"></path>
                  <path d="M19 6v14c0 1-1 2-2 2H7c-1 0-2-1-2-2V6"></path>
                  <path d="M8 6V4c0-1 1-2 2-2h4c1 0 2 1 2 2v2"></path>
                </svg>
                卸载
              </button>
            </div>
          </div>
        </div>
      </div>

      <div class="available-section">
        <h3>可用版本</h3>
        <div v-if="loadingAvailableVersions" class="loading-indicator">
          <div class="spinner-small"></div>
          <span>正在加载版本...</span>
        </div>
        <div v-else-if="filteredVersions.length === 0" class="empty-versions">
          <p>未找到可用版本</p>
        </div>
        <div v-else class="available-list">
          <div v-for="item in [...filteredVersions].reverse()" :key="item.version" class="available-item">
            <span class="version-text">{{ item.version }}</span>
            <span v-if="item.installed" class="installed-badge">已安装</span>
            <button v-else class="install-small-btn"
              @click="installVersionInput = item.version; showInstallModal = true">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <line x1="12" y1="5" x2="12" y2="19"></line>
                <line x1="5" y1="12" x2="19" y2="12"></line>
              </svg>
            </button>
          </div>
        </div>
      </div>
    </div>

    <div v-if="showInstallModal" class="modal-overlay" @click.self="showInstallModal = false">
      <div class="modal-content">
        <div class="modal-header">
          <h3>安装 {{ tool?.name }}</h3>
          <button class="modal-close" @click="showInstallModal = false">×</button>
        </div>
        <div class="modal-body">
          <label>版本</label>
          <div class="version-select-wrapper">
            <input type="text" v-model="installVersionInput" placeholder="搜索或输入版本号..." class="version-input"
              @keyup.enter="handleInstall" @focus="showVersionDropdown = true"
              @blur="setTimeout(() => showVersionDropdown = false, 200)" autofocus />
            <div v-if="showVersionDropdown && filteredUninstalledVersions.length > 0" class="version-dropdown">
              <div v-for="v in filteredUninstalledVersions" :key="v" class="dropdown-item" @click="selectVersion(v)">
                {{ v }}
              </div>
              <div v-if="filteredUninstalledVersions.length === 0" class="dropdown-empty">
                没有找到匹配的版本
              </div>
            </div>
          </div>
          <div v-if="installProgress.length > 0" class="progress-section">
            <div class="progress-header">安装进度</div>
            <div class="progress-logs">
              <div v-for="(log, index) in installProgress" :key="index" class="progress-log">
                <span class="log-icon">{{ log.type === 'success' ? '✓' : log.type === 'error' ? '✗' : '•' }}</span>
                <span class="log-text">{{ log.message }}</span>
              </div>
            </div>
            <div v-if="isInstalling" class="progress-bar">
              <div class="progress-fill"></div>
            </div>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn-cancel" @click="showInstallModal = false" :disabled="isInstalling">取消</button>
          <button class="btn-install" @click="handleInstall" :disabled="!installVersionInput.trim() || isInstalling">
            <svg v-if="isInstalling" class="spinner-mini" viewBox="0 0 24 24">
              <circle class="path" cx="12" cy="12" r="10" fill="none" stroke-width="4"></circle>
            </svg>
            {{ isInstalling ? '安装中...' : '安装' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.tool-detail {
  flex: 1;
  display: flex;
  overflow: hidden;
  position: relative;
}

.empty-detail {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  color: #858585;
}

.empty-icon {
  width: 64px;
  height: 64px;
  margin-bottom: 16px;
  opacity: 0.3;
}

.empty-detail p {
  margin: 0;
  font-size: 14px;
}

.detail-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow-y: auto;
  overflow-x: hidden;
  padding: 24px;
  max-height: 100vh;
  box-sizing: border-box;
}

.error-banner {
  display: flex;
  justify-content: space-between;
  align-items: center;
  background-color: #4e1818;
  color: #ff7b72;
  padding: 12px 16px;
  border-radius: 8px;
  margin-bottom: 20px;
}

.error-close {
  background: none;
  border: none;
  color: #ff7b72;
  font-size: 20px;
  cursor: pointer;
  padding: 0;
  line-height: 1;
}

.detail-header {
  display: flex;
  align-items: center;
  gap: 16px;
  margin-bottom: 24px;
}

.header-icon {
  width: 56px;
  height: 56px;
  border-radius: 12px;
  background: linear-gradient(135deg, #569cd6 0%, #4ec9b0 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 24px;
  font-weight: 600;
  color: #fff;
}

.header-info {
  flex: 1;
  min-width: 0;
}

.header-top-row {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 4px;
}

.header-info h2 {
  margin: 0;
  font-size: 20px;
  font-weight: 600;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.source {
  font-size: 12px;
  color: #858585;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.status {
  font-size: 11px;
  padding: 1px 6px;
  border-radius: 3px;
  flex-shrink: 0;
}

.status.installed {
  background-color: #2d5a27;
  color: #6a9955;
}

.install-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 16px;
  background-color: #094771;
  border: none;
  border-radius: 8px;
  color: #fff;
  cursor: pointer;
  font-size: 14px;
  transition: background-color 0.2s;
}

.install-btn:hover:not(:disabled) {
  background-color: #0a5388;
}

.install-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.plus-icon {
  width: 16px;
  height: 16px;
}

.current-version {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px;
  background-color: #2d2d30;
  border-radius: 8px;
  margin-bottom: 24px;
}

.current-version .label {
  font-size: 14px;
  color: #858585;
}

.current-version .value {
  font-size: 16px;
  font-weight: 600;
  color: #4ec9b0;
}

.versions-section,
.available-section {
  margin-bottom: 24px;
}

.versions-section h3,
.available-section h3 {
  margin: 0 0 12px 0;
  font-size: 16px;
  font-weight: 600;
}

.empty-versions {
  padding: 32px;
  text-align: center;
  background-color: #2d2d30;
  border-radius: 8px;
  color: #858585;
}

.empty-versions p {
  margin: 0;
  font-size: 14px;
}

.versions-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.version-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  background-color: #2d2d30;
  border-radius: 8px;
}

.version-info {
  display: flex;
  align-items: center;
  gap: 8px;
}

.version-number {
  font-size: 14px;
}

.default-badge {
  font-size: 11px;
  background-color: #094771;
  color: #569cd6;
  padding: 2px 6px;
  border-radius: 4px;
}

.version-actions {
  display: flex;
  gap: 8px;
}

.action-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 12px;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-size: 12px;
  transition: opacity 0.2s;
}

.action-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.action-btn svg {
  width: 14px;
  height: 14px;
}

.action-btn.set-default {
  background-color: #2d5a27;
  color: #6a9955;
}

.action-btn.set-default:hover {
  background-color: #3a7033;
}

.action-btn.uninstall {
  background-color: #4e1818;
  color: #ff7b72;
}

.action-btn.uninstall:hover {
  background-color: #611f1f;
}

.loading-indicator {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 32px;
  gap: 8px;
  color: #858585;
}

.spinner-small {
  width: 20px;
  height: 20px;
  border: 2px solid #3c3c3c;
  border-top-color: #569cd6;
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

.available-list {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.available-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 12px;
  background-color: #2d2d30;
  border-radius: 6px;
}

.version-text {
  font-size: 13px;
}

.installed-badge {
  font-size: 11px;
  color: #6a9955;
  background-color: #2d5a27;
  padding: 2px 6px;
  border-radius: 4px;
}

.install-small-btn {
  background-color: #094771;
  color: #569cd6;
  border: none;
  border-radius: 4px;
  padding: 6px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
}

.install-small-btn:hover {
  background-color: #0a5388;
}

.install-small-btn svg {
  width: 14px;
  height: 14px;
}

.more-versions {
  text-align: center;
  padding: 8px;
  font-size: 12px;
  color: #858585;
}

.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.7);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 100;
}

.modal-content {
  background-color: #252526;
  border-radius: 12px;
  width: 400px;
  max-width: 90%;
  overflow: hidden;
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  border-bottom: 1px solid #3c3c3c;
}

.modal-header h3 {
  margin: 0;
  font-size: 16px;
}

.modal-close {
  background: none;
  border: none;
  color: #858585;
  font-size: 20px;
  cursor: pointer;
  padding: 0;
  line-height: 1;
}

.modal-body {
  padding: 20px;
}

.modal-body label {
  display: block;
  margin-bottom: 8px;
  font-size: 14px;
  color: #858585;
}

.version-input {
  width: 100%;
  padding: 10px 12px;
  background-color: #1e1e1e;
  border: 1px solid #3c3c3c;
  border-radius: 8px;
  color: #e4e4e4;
  font-size: 14px;
  box-sizing: border-box;
}

.version-input:focus {
  outline: none;
  border-color: #569cd6;
}

.suggestions {
  margin-top: 12px;
}

.suggestions span {
  font-size: 12px;
  color: #858585;
  margin-right: 8px;
}

.suggestion-btn {
  background-color: #3c3c3c;
  color: #e4e4e4;
  border: none;
  border-radius: 4px;
  padding: 4px 8px;
  font-size: 12px;
  cursor: pointer;
  margin-right: 6px;
  margin-bottom: 4px;
}

.suggestion-btn:hover {
  background-color: #4a4a4a;
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  padding: 16px 20px;
  border-top: 1px solid #3c3c3c;
}

.btn-cancel,
.btn-install {
  padding: 10px 20px;
  border: none;
  border-radius: 8px;
  font-size: 14px;
  cursor: pointer;
}

.btn-cancel {
  background-color: #3c3c3c;
  color: #e4e4e4;
}

.btn-cancel:hover {
  background-color: #4a4a4a;
}

.btn-install {
  background-color: #094771;
  color: #fff;
  display: flex;
  align-items: center;
  gap: 8px;
}

.btn-install:hover:not(:disabled) {
  background-color: #0a5388;
}

.btn-install:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.spinner-mini {
  width: 14px;
  height: 14px;
  animation: spin 1s linear infinite;
}

.spinner-mini .path {
  stroke: #fff;
  stroke-linecap: round;
  animation: dash 1.5s ease-in-out infinite;
}

@keyframes dash {
  0% {
    stroke-dasharray: 1, 150;
    stroke-dashoffset: 0;
  }

  50% {
    stroke-dasharray: 90, 150;
    stroke-dashoffset: -35;
  }

  100% {
    stroke-dasharray: 90, 150;
    stroke-dashoffset: -124;
  }
}

.progress-section {
  margin-top: 16px;
  padding: 12px;
  background-color: #1e1e1e;
  border-radius: 8px;
  border: 1px solid #3c3c3c;
}

.progress-header {
  font-size: 12px;
  font-weight: 600;
  color: #858585;
  margin-bottom: 8px;
}

.progress-logs {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.progress-log {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 12px;
}

.log-icon {
  font-size: 10px;
  width: 14px;
  text-align: center;
}

.progress-log .log-icon {
  color: #569cd6;
}

.progress-log:last-child .log-icon {
  color: #6a9955;
}

.progress-log:last-child.error .log-icon {
  color: #ff7b72;
}

.log-text {
  color: #e4e4e4;
}

.progress-log:last-child.error .log-text {
  color: #ff7b72;
}

.progress-bar {
  margin-top: 10px;
  height: 4px;
  background-color: #3c3c3c;
  border-radius: 2px;
  overflow: hidden;
}

.progress-fill {
  height: 100%;
  background: linear-gradient(90deg, #569cd6, #4ec9b0);
  border-radius: 2px;
  animation: progressMove 1s ease-in-out infinite;
}

@keyframes progressMove {
  0% {
    width: 0%;
  }

  50% {
    width: 70%;
  }

  100% {
    width: 100%;
  }
}

.version-select-wrapper {
  position: relative;
  width: 100%;
}

.version-dropdown {
  position: absolute;
  top: calc(100% + 4px);
  left: 0;
  right: 0;
  max-height: 200px;
  overflow-y: auto;
  background-color: #1e1e1e;
  border: 1px solid #3c3c3c;
  border-radius: 6px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
  z-index: 100;
}

.dropdown-item {
  padding: 8px 12px;
  font-size: 13px;
  color: #e4e4e4;
  cursor: pointer;
  border-bottom: 1px solid #2d2d2d;
}

.dropdown-item:last-child {
  border-bottom: none;
}

.dropdown-item:hover {
  background-color: #3c3c3c;
}

.dropdown-empty {
  padding: 12px;
  text-align: center;
  font-size: 12px;
  color: #858585;
}
</style>