<script setup>
defineProps({
  tools: {
    type: Array,
    required: true
  },
  loading: {
    type: Boolean,
    default: false
  },
  selectedTool: {
    type: Object,
    default: null
  }
})

const emit = defineEmits(['select'])
</script>

<template>
  <div class="tool-list">
    <div class="list-header">
      <h2>工具</h2>
      <span class="count">{{ tools.length }} 已安装</span>
    </div>
    
    <div class="list-content">
      <div v-if="loading" class="loading-container">
        <div class="loading-spinner"></div>
        <span>正在加载工具...</span>
      </div>
      
      <div v-else-if="tools.length === 0" class="empty-state">
        <svg class="empty-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
          <path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z"></path>
          <polyline points="3.27 6.96 12 12.01 20.73 6.96"></polyline>
          <line x1="12" y1="22.08" x2="12" y2="12"></line>
        </svg>
        <p>没有安装工具</p>
      </div>
      
      <div v-else class="tools-grid">
        <div 
          v-for="tool in tools" 
          :key="tool.name"
          class="tool-card"
          :class="{ selected: selectedTool?.name === tool.name }"
          @click="emit('select', tool)"
        >
          <div class="tool-icon">
            {{ tool.name.charAt(0).toUpperCase() }}
          </div>
          <div class="tool-info">
            <span class="tool-name">{{ tool.name }}</span>
            <span class="tool-version">{{ tool.version || tool.default }}</span>
          </div>
          <div class="tool-versions-count">
            {{ tool.versions?.length || 0 }} 个版本
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.tool-list {
  width: 320px;
  display: flex;
  flex-direction: column;
  background-color: #252526;
  border-right: 1px solid #3c3c3c;
  overflow: hidden;
}

.list-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  border-bottom: 1px solid #3c3c3c;
}

.list-header h2 {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
}

.count {
  font-size: 12px;
  color: #858585;
  background-color: #3c3c3c;
  padding: 2px 8px;
  border-radius: 10px;
}

.list-content {
  flex: 1;
  overflow-y: auto;
}

.loading-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 40px 20px;
  gap: 12px;
}

.loading-spinner {
  width: 32px;
  height: 32px;
  border: 2px solid #3c3c3c;
  border-top-color: #569cd6;
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px 20px;
  color: #858585;
}

.empty-icon {
  width: 48px;
  height: 48px;
  margin-bottom: 16px;
  opacity: 0.5;
}

.empty-state p {
  margin: 0;
  font-size: 14px;
}

.tools-grid {
  padding: 12px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.tool-card {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px;
  background-color: #2d2d30;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s;
  border: 1px solid transparent;
}

.tool-card:hover {
  background-color: #37373d;
}

.tool-card.selected {
  background-color: #094771;
  border-color: #569cd6;
}

.tool-icon {
  width: 36px;
  height: 36px;
  border-radius: 8px;
  background: linear-gradient(135deg, #569cd6 0%, #4ec9b0 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 14px;
  font-weight: 600;
  color: #fff;
}

.tool-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.tool-name {
  font-size: 14px;
  font-weight: 500;
}

.tool-version {
  font-size: 12px;
  color: #858585;
}

.tool-versions-count {
  font-size: 11px;
  color: #6a9955;
  background-color: #2d5a27;
  padding: 2px 8px;
  border-radius: 10px;
}
</style>