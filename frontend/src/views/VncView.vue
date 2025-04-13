<template>
  <div class="vnc-container">
    <div class="vnc-card-list">
      <div
        v-for="(connection, index) in vncConnections"
        :key="index"
        class="vnc-card"
        :class="{ active: activeConnectionIndex === index }"
        @click="toggleConnection(index)"
      >
        <div class="vnc-card-content">
          <div class="vnc-card-info">
            <span class="vnc-card-title">{{ connection.name }}</span><br/>
            <span class="vnc-card-url">{{ connection.url }}</span>
          </div>
        </div>
        <div class="vnc-card-actions" v-if="index !== 0">
          <a-dropdown>
            <MoreOutlined class="action-icon" @click.stop />
            <template #overlay>
              <a-menu>
                <a-menu-item @click.stop="showEditConnectionModal(index)">
                  <EditOutlined />
                  <span>修改</span>
                </a-menu-item>
                <a-menu-item @click.stop="deleteConnection(index)">
                  <DeleteOutlined />
                  <span>删除</span>
                </a-menu-item>
              </a-menu>
            </template>
          </a-dropdown>
        </div>
      </div>
      <div class="vnc-card add-card" @click="showAddConnectionModal">
        <div class="vnc-card-content">
          <span class="vnc-card-url">添加新连接</span>
        </div>
      </div>
    </div>
    <!-- 添加编辑连接的模态框 -->
    <a-modal
      v-model:visible="isEditConnectionModalVisible"
      title="修改连接"
      @ok="editConnection"
    >
      <a-input v-model:value="currentConnectionName" placeholder="名称" style="margin-bottom: 8px;" />
      <a-input v-model:value="currentConnectionUrl" placeholder="地址 (格式: host:port)" style="margin-bottom: 8px;" />
      <a-input v-model:value="currentConnectionPassword" placeholder="密码(可选)" type="password" />
    </a-modal>
    <VncContainer
      ref="vncContainerRef"
      :vncUrl="vncUrl"
      :vncPassword="vncPassword"
      @connected="onConnected"
      @disconnected="onDisconnected"
      @error="onError"
    />
    <a-modal
      v-model:visible="isAddConnectionModalVisible"
      title="添加新连接"
      @ok="addConnection"
    >
      <a-input v-model:value="currentConnectionName" placeholder="名称" style="margin-bottom: 8px;" />
      <a-input v-model:value="currentConnectionUrl" placeholder="地址 (格式: host:port)" style="margin-bottom: 8px;" />
      <a-input v-model:value="currentConnectionPassword" placeholder="密码(可选)" type="password" />
    </a-modal>
  </div>
</template>

<script setup lang="js">
import { ref, onMounted, watch } from 'vue';
import { useRoute } from 'vue-router';
import { message } from 'ant-design-vue';
import { EditOutlined, DeleteOutlined, MoreOutlined } from '@ant-design/icons-vue';
import VncContainer from '@/components/VncContainer.vue';
// import axios from 'axios';

const apiUrl = import.meta.env.VITE_API_URL;
const route = useRoute(); // 获取路由对象

const vncUrl = ref('');
const vncPassword = ref('');
const vncContainerRef = ref();

const vncConnections = ref([]);
const isAddConnectionModalVisible = ref(false);
const isEditConnectionModalVisible = ref(false);
const currentConnectionName = ref('');
const currentConnectionUrl = ref('');
const currentConnectionPassword = ref('');
const currentConnectionIndex = ref(0);
const activeConnectionIndex = ref(null);

// 自动连接到指定的 VNC 服务器
const autoConnect = () => {
  const hostname = route.query.hostname; // 从路由参数中获取 hostname
  if (hostname) {
    const connectionIndex = vncConnections.value.findIndex(conn => conn.name === hostname);
    if (connectionIndex !== -1) {
      toggleConnection(connectionIndex); // 自动切换到指定连接
    } else {
      message.warning(`未找到名为 "${hostname}" 的连接，默认连接到第一个`);
      if (vncConnections.value.length > 0) {
        toggleConnection(0); // 默认连接到第一个
      }
    }
  } else if (vncConnections.value.length > 0) {
    toggleConnection(0); // 如果没有指定 hostname，默认连接到第一个
  }
};

const connectVNC = () => {
  if (!vncUrl.value) {
    message.warning('请输入VNC服务器地址');
  } else {
    vncContainerRef.value?.connectVNC();
  }
};

const disconnectVNC = () => {
  vncContainerRef.value?.disconnectVNC();
};

const onConnected = () => {
  message.success('VNC连接成功');
};

const onDisconnected = () => {
  message.info('VNC已断开');
};

const onError = (error) => {
  const errorMessage = error?.message || error?.toString() || '未知错误';
  message.error(`VNC错误: ${errorMessage}`);
};

const toggleConnection = async (index) => {
  try {
    // 如果点击当前活动连接，则断开
    if (activeConnectionIndex.value === index) {
      activeConnectionIndex.value = null;
      vncUrl.value = '';
      vncPassword.value = '';
      disconnectVNC();
      return;
    }

    // 如果有活动连接，先断开
    if (activeConnectionIndex.value !== null) {
      disconnectVNC();
      await new Promise(resolve => setTimeout(resolve, 500)); // 等待断开连接完成
    }

    // 获取新连接信息
    const connection = vncConnections.value[index];
    if (!connection?.url) {
      throw new Error('连接数据无效');
    }

    // 设置新连接信息
    vncUrl.value = connection.url;
    vncPassword.value = connection.password || '';
    activeConnectionIndex.value = index;

    // 建立新连接
    await new Promise(resolve => setTimeout(resolve, 100)); // 确保状态更新
    connectVNC();
  } catch (error) {
    message.error(`切换连接失败: ${error.message}`);
    // 发生错误时重置状态
    activeConnectionIndex.value = null;
    vncUrl.value = '';
    vncPassword.value = '';
  }
};

const showEditConnectionModal = (index) => {
  const connection = vncConnections.value[index];
  currentConnectionName.value = connection.name;
  currentConnectionUrl.value = connection.url;
  currentConnectionPassword.value = connection.password || '';
  isEditConnectionModalVisible.value = true;
  currentConnectionIndex.value = index;
}

const showAddConnectionModal = () => {
  currentConnectionName.value = '';
  currentConnectionUrl.value = '';
  currentConnectionPassword.value = '';
  isAddConnectionModalVisible.value = true;
};

const addConnection = async () => {
  if (!currentConnectionUrl.value) {
    message.warning('请填写地址');
    return;
  }

  // 检查并补全端口
  if (!currentConnectionUrl.value.includes(':')) {
    currentConnectionUrl.value += ':5900';
  }

  // 生成默认名称
  let defaultName = '新连接';
  let counter = 1;
  while (vncConnections.value.some(conn => conn.name === defaultName)) {
    defaultName = `新连接${counter++}`;
  }

  const connection = {
    name: currentConnectionName.value || defaultName,
    url: currentConnectionUrl.value,
    password: currentConnectionPassword.value,
  };

  try {
    const response = await fetch(`${apiUrl}/api/vnc/connections`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(connection)
    });
    if (!response.ok) throw new Error('添加连接失败');
    vncConnections.value.push(connection);
    message.success('连接已添加');
  } catch (error) {
    message.error('添加连接失败');
  }

  isAddConnectionModalVisible.value = false;
};

const editConnection = async () => {
  let index = currentConnectionIndex.value;
  let connection = {
    name: currentConnectionName.value,
    url: currentConnectionUrl.value,
    password: currentConnectionPassword.value,
  }
  console.log(connection);
  try {
    const response = await fetch(`${apiUrl}/api/vnc/connections/${index}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(connection)
    });
    if (!response.ok) throw new Error('更新连接失败');
    message.success('连接已更新');
    // 更新连接数据
    vncConnections.value[index] = {
      name: currentConnectionName.value,
      url: currentConnectionUrl.value,
    };
  } catch (error) {
    message.error('更新连接失败');
  }
};

const deleteConnection = async (index) => {
  try {
    const response = await fetch(`${apiUrl}/api/vnc/connections/${index}`, {
      method: 'DELETE'
    });
    if (!response.ok) throw new Error('删除连接失败');
    vncConnections.value.splice(index, 1);
    message.success('连接已删除');
  } catch (error) {
    message.error('删除连接失败');
  }
};

const loadConnections = async () => {
  try {
    const response = await fetch(`${apiUrl}/api/vnc/connections`);
    if (!response.ok) throw new Error('加载连接失败');
    vncConnections.value = await response.json();
  } catch (error) {
    message.warning('加载连接失败，使用默认连接');
    vncConnections.value = [
      { name: '本地服务器', url: 'localhost:5900', password: '' },
    ];
  }
};

onMounted(async () => {
  await loadConnections();
  if (route.query.hostname) {
    autoConnect();
  }
});
</script>

<style scoped>
.vnc-container {
  height: 92vh;
  width: 100%;
  padding: 16px;
  display: flex;
  flex-direction: column;
  box-sizing: border-box;
  overflow: auto;
}

.vnc-header {
  font-size: 18px;
  font-weight: bold;
  margin-bottom: 12px;
  text-align: center;
}

.vnc-card-list {
  display: flex;
  overflow-x: auto;
  gap: 12px;
  padding: 8px 0;
}

.vnc-card {
  flex: 0 0 auto;
  width: 200px;
  height: 48px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 12px;
  border: 1px solid #e8e8e8;
  border-radius: 4px;
  cursor: pointer;
  background-color: #fff;
  transition: all 0.2s ease;
  margin: 0;
}

.vnc-card:hover {
  border-color: #40a9ff;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.vnc-card.active {
  background-color: #e6f7ff;
  border-color: #1890ff;
  color: #1890ff;
}

.vnc-card-content {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  justify-content: center;
  min-width: 0;
}

.vnc-card-title {
  font-weight: 500;
  font-size: 14px;
  color: #262626;
  margin-bottom: 2px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  width: 100%;
}

.vnc-card-url {
  font-size: 12px;
  color: #8c8c8c;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  width: 100%;
}

.add-card {
  justify-content: center;
  align-items: center;
  background-color: #fafafa;
  border: 1px dashed #d9d9d9;
}

.add-card:hover {
  border-color: #40a9ff;
  color: #40a9ff;
}
</style>