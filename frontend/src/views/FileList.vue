<template>
  <div class="file-manager">
    <div class="welcome-section">
      <h1 style="padding-top: 12px;">欢迎使用文件管理系统</h1>
      <p class="welcome-text">您可以在此上传、管理和下载您的文件</p>
    </div>
    
    <div class="uploader-section">
      <div class="uploader-card">
        <h3>上传文件</h3>
        <p class="uploader-hint">支持单个文件上传</p>
        <FileUploader @upload-success="fetchFiles" />
      </div>
    </div>
    
    <div class="file-list">
      <div class="list-card">
        <h2>文件列表</h2>
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
              <FileOperations :file="record" @file-deleted="fetchFiles" />
            </template>
          </template>
        </a-table>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { message } from 'ant-design-vue'
import FileOperations from '@/components/FileOperations.vue'
import FileUploader from '@/components/FileUploader.vue'

interface FileInfo {
  filename: string
  size: number
  upload_time: string
}

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

onMounted(() => {
  fetchFiles()
})
</script>

<style scoped>
.file-manager {
  background: #fff;
  max-width: 1200px;
  margin: 0 auto;
  min-height: 100vh;
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