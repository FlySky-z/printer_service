<template>
  <a-layout class="layout">
    <a-layout-header class="header">
      <MenuUnfoldOutlined
        v-if="collapsed"
        class="trigger"
        @click="toggleCollapsed"
      />
      <MenuFoldOutlined
        v-else
        class="trigger"
        @click="toggleCollapsed"
      />
      <h2 style="margin-left: 24px">打印管理系统</h2>
    </a-layout-header>
    <a-layout>
      <a-layout-sider
        v-model:collapsed="collapsed"
        collapsible
        :trigger="null"
        width="200"
        :class="['sider', collapsed ? 'collapsed' : '', isMobileVisible ? 'mobile-visible' : '']"   
      >
        <a-menu
          mode="inline"
          v-model:selectedKeys="selectedKeys"
        >
          <a-menu-item key="1" @click="() => $router.push('/')">
            <template #icon>
              <PrinterOutlined />
            </template>
            <span>打印</span>
          </a-menu-item>
          <a-menu-item key="2" @click="() => $router.push('/vnc')">
            <template #icon>
              <DesktopOutlined />
            </template>
            <span>VNC</span>
          </a-menu-item>
        </a-menu>
      </a-layout-sider>
      <a-layout-content :class="['content', collapsed ? 'collapsed' : '']" @click="handleOutsideClick">
        <router-view />
      </a-layout-content>
    </a-layout>
  </a-layout>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import {
  PrinterOutlined,
  DesktopOutlined,
  MenuUnfoldOutlined,
  MenuFoldOutlined
} from '@ant-design/icons-vue'

const collapsed = ref(false)
const selectedKeys = ref<string[]>(['1'])
const isMobile = ref(false)
const isMobileVisible = ref(false)

// 检测是否为移动设备
const checkMobile = () => {
  isMobile.value = window.innerWidth <= 768
  if (isMobile.value && !collapsed.value) {
    collapsed.value = true
  }
}

// 切换侧边栏
const toggleCollapsed = () => {
  collapsed.value = !collapsed.value
  if (isMobile.value) {
    isMobileVisible.value = !isMobileVisible.value
  }
}

// 处理点击外部区域
const handleOutsideClick = () => {
  if (isMobile.value && isMobileVisible.value) {
    isMobileVisible.value = false
  }
}

// 监听窗口大小变化
onMounted(() => {
  checkMobile()
  window.addEventListener('resize', checkMobile)
})

onUnmounted(() => {
  window.removeEventListener('resize', checkMobile)
})
</script>

<style scoped>
.layout {
  min-height: 100vh;
}

.sider {
  background: #fff;
  transition: all 0.3s ease;
  box-shadow: 2px 0 8px 0 rgba(29, 35, 41, 0.05);
  position: fixed;
  top: 64px;
  bottom: 0;
  left: 0;
  z-index: 10;
  width: 200px;
}

.sider.collapsed {
  width: 80px;
}

.sider :deep(.ant-menu) {
  padding: 0;
}

.sider :deep(.ant-menu-item) {
  margin: 8px 0;
  height: 40px;
  line-height: 40px;
  color: #000;
}

.header {
  background: #fff;
  padding: 0 24px;
  display: flex;
  align-items: center;
  box-shadow: 0 1px 4px rgba(0, 21, 41, 0.08);
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  z-index: 20;
}

.trigger {
  font-size: 18px;
  line-height: 64px;
  padding: 0 24px;
  cursor: pointer;
  transition: color 0.3s;
}

.trigger:hover {
  color: #1890ff;
}

.content {
  transition: margin-left 0.3s ease;
  margin-left: 200px;
  margin-top: 64px;
  background: #fff;
  min-height: calc(100vh - 88px);
}

.content.collapsed {
  margin-left: 80px;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .sider {
    transform: translateX(-100%);
  }
  
  .sider.mobile-visible {
    transform: translateX(0);
  }
  
  .content, .content.collapsed {
    margin-left: 0;
  }
  
  .header h2 {
    font-size: 18px;
  }
  
  .trigger {
    padding: 0 16px;
  }
}
</style>