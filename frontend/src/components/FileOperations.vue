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
}

const props = defineProps<Props>()
const emit = defineEmits(['fileDeleted'])
const router = useRouter()
const apiUrl = import.meta.env.VITE_API_URL;
const loadingEdit = ref(false); // 加载状态
const loadingPrint = ref(false); // 加载状态

const handlePrint = async () => {
    // 检查文件是否为word/pdf格式
    const fileExtension = props.file.filename.split('.').pop()?.toLowerCase()
    if (!fileExtension) {
        message.error('文件名无效')
        return
    }
    // 仅允许打印Word和PDF文件
    if (!['doc', 'docx', 'pdf'].includes(fileExtension)) {
        message.error('该功能仅能打印Word和PDF文件, 其他文件使用【编辑】功能')
        return
    }
    try {
        message.success('已发送打印请求')
        loadingPrint.value = true;
        const response = await fetch(`${apiUrl}/print`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({ filename: props.file.filename })
        })
        loadingPrint.value = false;
        if (!response.ok) {
            throw new Error('打印失败')
        } else {
            message.success('打印成功')
        }
    } catch (error) {
        loadingPrint.value = false;
        message.error('打印失败')
    }
}

const handleEdit = async () => {
    loadingEdit.value = true; 
    try {
        message.success('已发送编辑请求')
        const response = await fetch(`${apiUrl}/preopen`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({ filename: props.file.filename })
        })
        if (!response.ok) {
            loadingEdit.value = false;
            throw new Error('编辑失败')
        }
        // 跳转到 /vnc 并传递默认连接参数
        router.push({ path: `/vnc`, query: { hostname: "本地服务器" } })
    } catch (error) {
        loadingEdit.value = false;
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
    } catch (error) {
        message.error('下载失败')
        console.error(error)
    }
}

const handleDelete = async () => {
    try {
        const response = await fetch(`${apiUrl}/files/${props.file.filename}`, {
            method: 'DELETE'
        })
        if (!response.ok) throw new Error('删除失败')
        message.success('删除成功')
        emit('fileDeleted')
    } catch (error) {
        message.error('删除失败')
    }
}
</script>