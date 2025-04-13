<template>
  <div id="noVNC_container" ref="vncContainerRef"></div>
</template>

<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount } from 'vue';


const props = defineProps({
  vncUrl: {
    type: String,
    required: true,
  },
  vncPassword: {
    type: String,
    default: '',
  },
});

const emit = defineEmits(['connected', 'disconnected', 'error']);

const rfb = ref<any | null>(null);
const vncContainerRef = ref<HTMLDivElement | null>(null);
const apiUrl = import.meta.env.VITE_API_URL 
  ? import.meta.env.VITE_API_URL.replace(/^https?:\/\//, '') 
  : window.location.host;

// 创建ResizeObserver来监听容器大小变化
let resizeObserver: ResizeObserver | null = null;

// 处理容器大小变化
const handleContainerResize = () => {
  if (rfb.value) {
    applyScaling();
  }
};

// 应用缩放逻辑
const applyScaling = () => {
  if (!rfb.value || !vncContainerRef.value) return;

  const containerWidth = vncContainerRef.value.clientWidth;
  const containerHeight = vncContainerRef.value.clientHeight;
  if (rfb.value.scaleViewport !== undefined) {
    rfb.value.scaleViewport = true;
  }

  const fbWidth = rfb.value._fbWidth || 0;
  const fbHeight = rfb.value._fbHeight || 0;

  if (fbWidth > 0 && fbHeight > 0) {
    const scaleX = containerWidth / fbWidth;
    const scaleY = containerHeight / fbHeight;
    const scale = Math.min(scaleX, scaleY) * 0.95;

    if (typeof rfb.value.setViewportScale === 'function') {
      rfb.value.setViewportScale(scale);
    }
  }
};

const connectVNC = async () => {
  if (!props.vncUrl) {
    emit('error', { success: false, message: 'VNC服务器地址不能为空' });
    return;
  }

  const container = vncContainerRef.value;
  if (!container) {
    emit('error', { success: false, message: '无法找到VNC容器' });
    return;
  }

  try {
    const RFB = (await import('@novnc/novnc/lib/rfb')).default;

    // const protocol = window.location.protocol === 'https:' ? 'wss://' : 'ws://';
    // 当前只支持ws协议
    const protocol = 'ws://';
    const wsUrl = `${protocol}${apiUrl}/websockify?host=${encodeURIComponent(props.vncUrl)}`;
    console.log('VNC连接地址:', wsUrl);
    if (rfb.value) {
      rfb.value.disconnect();
      rfb.value = null;
    }

    rfb.value = new RFB(container, wsUrl, {
      credentials: { password: props.vncPassword },
      shared: true,
      clipViewport: true,
      scaleViewport: true,
      resizeSession: true,
      encrypt: false, // 禁用加密
    });

    rfb.value.addEventListener('connect', () => {
      emit('connected', { success: true, message: 'VNC连接成功' });
      setTimeout(applyScaling, 1000);
      window.addEventListener('resize', applyScaling);
    });

    rfb.value.addEventListener('disconnect', () => {
      window.removeEventListener('resize', applyScaling);
      emit('disconnected', { success: false, message: 'VNC连接断开' });
    });

    rfb.value.addEventListener('securityfailure', (e: any) => {
      emit('error', { success: false, message: `安全验证失败: ${e.detail.reason}` });
    });
  } catch (error) {
    emit('error', { success: false, message: `连接失败: ${error instanceof Error ? error.message : String(error)}` });
  }
};

const disconnectVNC = () => {
  if (rfb.value) {
    window.removeEventListener('resize', applyScaling);
    rfb.value.disconnect();
    rfb.value = null;
  }
};

onMounted(() => {
  if ('ResizeObserver' in window) {
    resizeObserver = new ResizeObserver(handleContainerResize);
    if (vncContainerRef.value) {
      resizeObserver.observe(vncContainerRef.value);
    }
  }
});

onBeforeUnmount(() => {
  if (resizeObserver) {
    resizeObserver.disconnect();
    resizeObserver = null;
  }
  disconnectVNC();
});

// 暴露connectVNC和disconnectVNC方法
defineExpose({
  connectVNC,
  disconnectVNC,
});
</script>

<style scoped>
#noVNC_container {
  flex: 1;
  min-height: 300px;
  width: 100%;
  border: 1px solid #d9d9d9;
  overflow: hidden;
  position: relative;
}
</style>