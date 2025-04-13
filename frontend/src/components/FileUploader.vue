<template>
    <div class="file-uploader">
        <a-upload-dragger
            :action="apiUrl"
            @change="handleChange"
            :before-upload="beforeUpload"
        >
            <p class="ant-upload-drag-icon">
                <upload-outlined />
            </p>
            <p class="ant-upload-text">
                将文件拖到此处，或<em>点击上传</em>
            </p>
        </a-upload-dragger>
    </div>
</template>

<script setup lang="ts">
import { message } from 'ant-design-vue'
import { UploadOutlined } from '@ant-design/icons-vue'

const emit = defineEmits(['uploadSuccess'])
const apiUrl = `${import.meta.env.VITE_API_URL}/files`

const handleChange = (info: any) => {
    if (info.file.status === 'done') {
        message.success('上传成功')
        emit('uploadSuccess')
    } else if (info.file.status === 'error') {
        message.error('上传失败')
    }
}

const beforeUpload = (file: File) => {
    if (!file) {
        message.error('文件不能为空')
        return false
    }
    // const isLt10M = file.size / 1024 / 1024 < 10
    // if (!isLt10M) {
    //   message.error('文件大小不能超过 10MB!')
    //   return false
    // }
    return true
}
</script>

<style scoped>
.file-uploader {
  width: 100%;
}
</style>