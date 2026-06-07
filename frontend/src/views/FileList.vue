<template>
  <div class="file-manager">
    <div class="welcome-section">
      <h1 style="padding-top: 12px;">欢迎使用文件管理系统</h1>
      <p class="welcome-text">您可以在此上传、管理和下载您的文件</p>
    </div>

    <div style="margin-bottom: 16px; display: flex; gap: 12px; align-items: center; flex-wrap: wrap; animation: fadeIn 0.5s ease 0.1s both;">
      <span style="color: #666; font-size: 13px; font-weight: 500;">可直接打印:</span>
      <span style="font-size: 13px; color: #555;">Word</span>
      <a-tag v-if="softwareStatus" :color="softwareStatus.wps ? 'success' : 'error'" style="margin: 0">
        {{ softwareStatus.wps ? 'WPS ✓' : '未安装WPS' }}
      </a-tag>
      <span style="font-size: 13px; color: #555;">PDF</span>
      <a-tag v-if="softwareStatus" :color="softwareStatus.acrobat ? 'success' : 'error'" style="margin: 0">
        {{ softwareStatus.acrobat ? 'Acrobat ✓' : '未安装Acrobat' }}
      </a-tag>
    </div>

    <a-spin :spinning="loadingPrinterStatus">
      <div style="display:flex;align-items:center;gap:12px;margin-bottom:8px; animation: fadeIn 0.5s ease 0.2s both;">
        <span style="font-size:13px;font-weight:500;color:#666">当前打印机:</span>
        <span style="font-size:13px;color:#333">{{ selectedPrinter || '无' }}</span>
      </div>
      <div style="display: flex; gap: 8px; flex-wrap: wrap; margin-bottom: 16px; animation: fadeIn 0.5s ease 0.2s both;">
        <div v-for="p in printerStatusList" :key="p.name" class="printer-card" @click="showJobs(p.name)" style="cursor:pointer">
          <div style="display:flex;align-items:center;gap:8px">
            <a-radio :value="p.name" :checked="selectedPrinter === p.name" @click.stop="selectedPrinter = p.name" />
            <div class="printer-card-left">
              <div class="printer-name">{{ p.name }}</div>
              <div class="printer-jobs">{{ p.job_count > 0 ? `${p.job_count} 个任务` : '队列为空' }}</div>
            </div>
          </div>
          <a-tag :color="printerStatusColor(p.status)" style="margin:0 0 0 12px">{{ printerStatusText(p.status) }}</a-tag>
        </div>
      </div>
    </a-spin>

    <a-modal v-model:open="jobsModalVisible" :title="`${selectedPrinterName} 队列`" :footer="null">
      <a-spin :spinning="loadingJobs">
        <div v-if="!loadingJobs && jobsList.length === 0" style="text-align:center;color:#999;padding:16px">队列为空</div>
        <a-list v-else :data-source="jobsList" item-layout="horizontal">
          <template #renderItem="{ item }">
            <a-list-item>{{ item.document || `任务 #${item.job_id}` }}</a-list-item>
          </template>
        </a-list>
      </a-spin>
    </a-modal>
    
    <div class="uploader-section">
      <div class="uploader-card">
        <h3>上传文件</h3>
        <p class="uploader-hint">支持单个文件上传</p>
        <FileUploader @upload-success="fetchFiles" />
      </div>
    </div>
    
    <div class="file-list">
      <div class="list-card">
        <div style="display:flex;align-items:center;gap:12px;margin-bottom:12px">
          <h2 style="margin:0">文件列表</h2>
          <a-button @click="handleVnc">远程桌面</a-button>
        </div>
        <p class="list-hint">共 {{ files.length }} 个文件</p>
        <a-table 
          :dataSource="files" 
          :columns="columns"
          :pagination="{ pageSize: 10 }"
          class="file-table"
          :scroll="{ x: 500 }"
        >
          <template #bodyCell="{ column, record }">
            <template v-if="column.key === 'action'">
              <FileOperations :file="record" :selected-printer="selectedPrinter" @file-deleted="fetchFiles" />
            </template>
          </template>
        </a-table>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { message } from 'ant-design-vue'
import FileOperations from '@/components/FileOperations.vue'
import FileUploader from '@/components/FileUploader.vue'

interface FileInfo {
  filename: string
  size: number
  upload_time: string
}

const router = useRouter()
const files = ref<FileInfo[]>([])
const apiUrl = import.meta.env.VITE_API_URL;
const columns = [
  {
    title: '文件名',
    dataIndex: 'filename',
    key: 'filename'
  },
  {
    title: '大小',
    dataIndex: 'size',
    key: 'size'
  },
  {
    title: '上传时间',
    dataIndex: 'upload_time',
    key: 'upload_time'
  },
  {
    title: '操作',
    key: 'action'
  }
]

const fetchFiles = async () => {
  try {
    const response = await fetch(`${apiUrl}/files`)
    if (!response.ok) throw new Error('获取文件列表失败')
    const data = await response.json()
    files.value = (Array.isArray(data) ? data : data.files || [])
      .map((item: FileInfo) => ({
        filename: item.filename,
        size: formatSize(item.size),
        upload_time: formatDate(item.upload_time),
        raw_upload_time: item.upload_time // Keep raw timestamp for sorting
      }))
      .sort((a: { raw_upload_time: string }, b: { raw_upload_time: string }) => 
        new Date(b.raw_upload_time).getTime() - new Date(a.raw_upload_time).getTime()
      )
  } catch (error) {
    message.error('获取文件列表失败')
    console.error(error)
  }
}

const formatSize = (bytes: number) => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

const formatDate = (timestamp: string) => {
  const date = new Date(timestamp)
  return date.toLocaleString()
}

const softwareStatus = ref<{ wps: boolean; acrobat: boolean } | null>(null)
const loadingPrinterStatus = ref(false)
const printerStatusList = ref<{ name: string; status: number; job_count: number }[]>([])
const selectedPrinter = ref('')
const printerStatusText = (s: number) => ({ 3: 'Idle', 4: 'Printing', 7: 'Offline' }[s] ?? `Unknown`)
const printerStatusColor = (s: number) => ({ 3: 'success', 4: 'processing', 7: 'default' }[s] ?? 'warning')

const jobsModalVisible = ref(false)
const loadingJobs = ref(false)
const jobsList = ref<{ job_id: number; document: string; status: number }[]>([])
const selectedPrinterName = ref('')

const showJobs = async (name: string) => {
  selectedPrinterName.value = name
  jobsModalVisible.value = true
  loadingJobs.value = true
  try {
    const res = await fetch(`${apiUrl}/api/printer/jobs/${encodeURIComponent(name)}`)
    if (res.ok) jobsList.value = await res.json()
  } catch {}
  loadingJobs.value = false
}

const handleVnc = async () => {
  try {
    const res = await fetch(`${apiUrl}/api/vnc/connections`)
    if (res.ok) {
      const connections = await res.json()
      if (connections?.length > 0) {
        router.push({ path: '/vnc', query: { hostname: connections[0].name } })
        return
      }
    }
  } catch {}
  router.push('/vnc')
}

onMounted(async () => {
  fetchFiles()
  loadingPrinterStatus.value = true
  try {
    const [swRes, prRes] = await Promise.all([
      fetch(`${apiUrl}/api/printer/software`),
      fetch(`${apiUrl}/api/printer/status`)
    ])
    if (swRes.ok) softwareStatus.value = await swRes.json()
    if (prRes.ok) {
      printerStatusList.value = await prRes.json()
      if (printerStatusList.value.length > 0) selectedPrinter.value = printerStatusList.value[0].name
    }
  } catch {}
  loadingPrinterStatus.value = false
})
</script>

<style scoped>
.file-manager {
  background: #fff;
  max-width: 1200px;
  margin: 0 auto;
  min-height: 100vh;
}

.printer-card {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  border: 1px solid #f0f0f0;
  border-radius: 8px;
  margin-bottom: 8px;
  background: #fafafa;
}

.printer-name {
  font-size: 14px;
  font-weight: 500;
  color: #1a1a1a;
}

.printer-jobs {
  font-size: 12px;
  color: #999;
  margin-top: 2px;
}

.welcome-section {
  text-align: center;
  margin-bottom: 32px;
  animation: fadeIn 0.5s ease;
}

.welcome-section h1 {
  font-size: 28px;
  color: #1a1a1a;
  margin-bottom: 8px;
}

.welcome-text {
  color: #666;
  font-size: 16px;
}

.uploader-section {
  margin-bottom: 32px;
  animation: fadeIn 0.5s ease 0.2s both;
}

.uploader-card {
  padding: 24px;
  background: #f8f9fa;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.05);
  transition: all 0.3s;
}

.uploader-card:hover {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}

.uploader-card h3 {
  margin-bottom: 8px;
  color: #1a1a1a;
}

.uploader-hint {
  color: #666;
  margin-bottom: 16px;
  font-size: 14px;
}

.file-list {
  animation: fadeIn 0.5s ease 0.4s both;
}

.list-card {
  padding: 12px;
  background: #fff;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.05);
}

.list-hint {
  color: #666;
  margin-bottom: 16px;
  font-size: 14px;
}

.file-table {
  margin-top: 16px;
}

h2 {
  margin-bottom: 12px;
  font-weight: 500;
  color: #1a1a1a;
}

@keyframes fadeIn {
  from {
    opacity: 0;
    transform: translateY(10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}
</style>