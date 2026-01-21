<template>
  <div class="dashboard">
    <!-- 状态卡片 -->
    <el-row :gutter="20" class="status-cards">
      <el-col :span="6">
        <el-card shadow="hover">
          <template #header>
            <div class="card-header">
              <el-icon :size="24" color="#409EFF"><Connection /></el-icon>
              <span>运行状态</span>
            </div>
          </template>
          <div class="card-content">
            <el-tag :type="statusTagType" size="large">{{ statusText }}</el-tag>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover">
          <template #header>
            <div class="card-header">
              <el-icon :size="24" color="#67C23A"><Document /></el-icon>
              <span>配置状态</span>
            </div>
          </template>
          <div class="card-content">
            <el-tag :type="systemStatus.config_exists ? 'success' : 'info'" size="large">
              {{ systemStatus.config_exists ? '已配置' : '未配置' }}
            </el-tag>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover">
          <template #header>
            <div class="card-header">
              <el-icon :size="24" color="#E6A23C"><Box /></el-icon>
              <span>Sing-box</span>
            </div>
          </template>
          <div class="card-content">
            <el-tag :type="systemStatus.singbox_exists ? 'success' : 'warning'" size="large">
              {{ systemStatus.singbox_exists ? '已安装' : '未安装' }}
            </el-tag>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover">
          <template #header>
            <div class="card-header">
              <el-icon :size="24" color="#909399"><InfoFilled /></el-icon>
              <span>进程 PID</span>
            </div>
          </template>
          <div class="card-content">
            <span class="pid-text">{{ systemStatus.singbox_pid || '-' }}</span>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 控制面板 -->
    <el-card class="control-panel">
      <template #header>
        <div class="panel-header">
          <span>服务控制</span>
        </div>
      </template>
      <div class="control-buttons">
        <el-button
          type="success"
          size="large"
          :icon="VideoPlay"
          :loading="actionLoading === 'start'"
          :disabled="systemStatus.singbox_status === 'running'"
          @click="handleStart"
        >
          启动
        </el-button>
        <el-button
          type="danger"
          size="large"
          :icon="VideoPause"
          :loading="actionLoading === 'stop'"
          :disabled="systemStatus.singbox_status !== 'running'"
          @click="handleStop"
        >
          停止
        </el-button>
        <el-button
          type="warning"
          size="large"
          :icon="Refresh"
          :loading="actionLoading === 'restart'"
          :disabled="systemStatus.singbox_status !== 'running'"
          @click="handleRestart"
        >
          重启
        </el-button>
        <el-divider direction="vertical" />
        <el-button
          type="primary"
          size="large"
          :icon="Download"
          :loading="actionLoading === 'upgrade'"
          @click="handleUpgrade"
        >
          {{ systemStatus.singbox_exists ? '更新 Sing-box' : '下载 Sing-box' }}
        </el-button>
      </div>
    </el-card>

    <!-- 版本信息 -->
    <el-card class="version-info">
      <template #header>
        <div class="panel-header">
          <span>版本信息</span>
          <el-button text type="primary" @click="fetchVersionInfo">
            <el-icon><Refresh /></el-icon>
            刷新
          </el-button>
        </div>
      </template>
      <el-descriptions :column="2" border>
        <el-descriptions-item label="Web 版本">{{ versionInfo.version || '-' }}</el-descriptions-item>
        <el-descriptions-item label="Sing-box 最新版本">{{ versionInfo.singbox_latest_version || '-' }}</el-descriptions-item>
      </el-descriptions>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  VideoPlay,
  VideoPause,
  Refresh,
  Download,
} from '@element-plus/icons-vue'
import api from '@/api'

interface SystemStatus {
  singbox_status: string
  singbox_pid: number
  config_exists: boolean
  singbox_exists: boolean
}

interface VersionInfo {
  version: string
  singbox_latest_version: string
}

const systemStatus = reactive<SystemStatus>({
  singbox_status: 'stopped',
  singbox_pid: 0,
  config_exists: false,
  singbox_exists: false,
})

const versionInfo = reactive<VersionInfo>({
  version: '',
  singbox_latest_version: '',
})

const actionLoading = ref<string | null>(null)

const statusTagType = computed(() => {
  switch (systemStatus.singbox_status) {
    case 'running':
      return 'success'
    case 'error':
      return 'danger'
    default:
      return 'info'
  }
})

const statusText = computed(() => {
  switch (systemStatus.singbox_status) {
    case 'running':
      return '运行中'
    case 'error':
      return '错误'
    default:
      return '已停止'
  }
})

async function fetchSystemStatus() {
  try {
    const response = await api.get('/system/status')
    Object.assign(systemStatus, response.data)
  } catch (error) {
    console.error('Failed to fetch system status:', error)
  }
}

async function fetchVersionInfo() {
  try {
    const response = await api.get('/system/version')
    Object.assign(versionInfo, response.data)
  } catch (error) {
    console.error('Failed to fetch version info:', error)
  }
}

async function handleStart() {
  actionLoading.value = 'start'
  try {
    await api.post('/system/start')
    ElMessage.success('启动成功')
    await fetchSystemStatus()
  } catch (error) {
    // Error handled by interceptor
  } finally {
    actionLoading.value = null
  }
}

async function handleStop() {
  actionLoading.value = 'stop'
  try {
    await api.post('/system/stop')
    ElMessage.success('停止成功')
    await fetchSystemStatus()
  } catch (error) {
    // Error handled by interceptor
  } finally {
    actionLoading.value = null
  }
}

async function handleRestart() {
  actionLoading.value = 'restart'
  try {
    await api.post('/system/restart')
    ElMessage.success('重启成功')
    await fetchSystemStatus()
  } catch (error) {
    // Error handled by interceptor
  } finally {
    actionLoading.value = null
  }
}

async function handleUpgrade() {
  try {
    await ElMessageBox.confirm(
      systemStatus.singbox_exists
        ? '确定要更新 Sing-box 到最新版本吗？更新过程中服务将会停止。'
        : '确定要下载并安装 Sing-box 吗？',
      '确认',
      { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' }
    )

    actionLoading.value = 'upgrade'
    const response = await api.post('/system/upgrade')
    ElMessage.success(`${systemStatus.singbox_exists ? '更新' : '安装'}成功: ${response.data.version}`)
    await fetchSystemStatus()
  } catch (error: any) {
    if (error !== 'cancel') {
      // Error handled by interceptor
    }
  } finally {
    actionLoading.value = null
  }
}

onMounted(() => {
  fetchSystemStatus()
  fetchVersionInfo()
})
</script>

<style scoped>
.dashboard {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.status-cards {
  margin-bottom: 0;
}

.card-header {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 500;
}

.card-content {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 60px;
}

.pid-text {
  font-size: 24px;
  font-weight: bold;
  color: #606266;
}

.control-panel .control-buttons {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}

.panel-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-weight: 500;
}

.version-info :deep(.el-descriptions__label) {
  width: 150px;
}
</style>
