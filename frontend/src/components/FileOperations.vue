<template>
    <a-space>
        <a-spin :spinning="loadingPrint">
            <a-button type="primary" size="small" @click="handlePrint">
            <template #icon><printer-outlined /></template>
            打印
            </a-button>
        </a-spin>
        <a-spin :spinning="loadingEdit">
            <a-button type="primary" size="small" @click="handleEdit">
                <template #icon><edit-outlined /></template>
                编辑
            </a-button>
        </a-spin>
        <a-button type="primary" size="small" @click="handleDownload">
            <template #icon><download-outlined /></template>
            下载
        </a-button>
        <a-button type="primary" danger size="small" @click="handleDelete">
            <template #icon><delete-outlined /></template>
            删除
        </a-button>
    </a-space>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { message } from 'ant-design-vue'
import { DeleteOutlined, PrinterOutlined, DownloadOutlined, EditOutlined } from '@ant-design/icons-vue'
import { useRouter } from 'vue-router'

interface Props {
    file: {
        filename: string
        size: number | string
        upload_time: string
    }
    selectedPrinter?: string
}

const props = defineProps<Props>()
const emit = defineEmits(['fileDeleted'])
const router = useRouter()
const apiUrl = import.meta.env.VITE_API_URL
const loadingEdit = ref(false)
const loadingPrint = ref(false)

const handlePrint = async () => {
    const ext = props.file.filename.split('.').pop()?.toLowerCase()
    if (!ext) { message.error('文件名无效'); return }
    if (!['doc', 'docx', 'pdf'].includes(ext)) {
        message.error('该功能仅能打印Word和PDF文件, 其他文件使用【编辑】功能')
        return
    }
    try {
        loadingPrint.value = true
        const response = await fetch(`${apiUrl}/print`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                filename: props.file.filename,
                printer_name: props.selectedPrinter ?? ''
            })
        })
        if (!response.ok) throw new Error()
        message.success('打印成功')
    } catch {
        message.error('打印失败')
    } finally {
        loadingPrint.value = false
    }
}

const handleEdit = async () => {
    loadingEdit.value = true
    try {
        message.success('已发送编辑请求')
        const response = await fetch(`${apiUrl}/preopen`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ filename: props.file.filename })
        })
        if (!response.ok) {
            loadingEdit.value = false
            throw new Error('编辑失败')
        }
        router.push({ path: `/vnc`, query: { hostname: "本地服务器" } })
    } catch {
        loadingEdit.value = false
        message.error('编辑失败')
    }
}

const handleDownload = async () => {
    try {
        const response = await fetch(`${apiUrl}/files/${props.file.filename}`)
        if (!response.ok) throw new Error('下载失败')
        const blob = await response.blob()
        const url = window.URL.createObjectURL(blob)
        const a = document.createElement('a')
        a.href = url
        a.download = props.file.filename
        document.body.appendChild(a)
        a.click()
        document.body.removeChild(a)
        window.URL.revokeObjectURL(url)
        message.success('下载成功')
    } catch {
        message.error('下载失败')
    }
}

const handleDelete = async () => {
    try {
        const response = await fetch(`${apiUrl}/files/${props.file.filename}`, { method: 'DELETE' })
        if (!response.ok) throw new Error('删除失败')
        message.success('删除成功')
        emit('fileDeleted')
    } catch {
        message.error('删除失败')
    }
}
</script>
