<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { useEnvStore } from '../stores/env';
import { useUiStore } from '../stores/ui';
import LaunchScreen from '../components/launcher/LaunchScreen.vue';
import FailureGuide from '../components/launcher/FailureGuide.vue';

// 定义 emits
const emit = defineEmits(['finish']);

const envStore = useEnvStore();
const uiStore = useUiStore();

const detectionComplete = ref(false);

// 组件挂载时开始环境检测
onMounted(async () => {
  await performDetection();
});

// 执行环境检测
const performDetection = async () => {
  try {
    await envStore.detectAll();
    detectionComplete.value = true;

    if (envStore.allPassed) {
      // 检测通过，更新状态
      uiStore.setLaunchState('success');
    } else {
      // 检测失败，显示引导页
      uiStore.setLaunchState('failed');
    }
  } catch (error) {
    console.error('环境检测失败:', error);
    uiStore.setLaunchState('failed');
    detectionComplete.value = true;
  }
};

// 进入主页面
const handleEnterMain = () => {
  emit('finish');
};

// 重试检测
const handleRetry = async () => {
  uiStore.setLaunchState('idle');
  detectionComplete.value = false;
  envStore.reset();
  await performDetection();
};

// 跳过检测
const handleSkip = () => {
  emit('finish');
};
</script>

<template>
  <div class="launch-view">
    <!-- 启动检测画面 -->
    <LaunchScreen
      v-if="!detectionComplete || uiStore.launchState === 'detecting' || uiStore.launchState === 'success'"
      @enter-main="handleEnterMain"
    />

    <!-- 检测失败引导页 -->
    <FailureGuide
      v-else-if="uiStore.launchState === 'failed'"
      @retry="handleRetry"
      @skip="handleSkip"
    />
  </div>
</template>

<style lang="scss" scoped>
.launch-view {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  overflow: hidden;
}
</style>
