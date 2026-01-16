<script setup lang="ts">
import { ref, onMounted } from 'vue';
import LaunchView from './views/LaunchView.vue';
import MainView from './views/MainView.vue';

const isLaunched = ref(false);

const onclickMinimise = () => {
  // @ts-ignore - Wails runtime
  if (window.runtime) {
    // @ts-ignore
    window.runtime.WindowMinimise();
  }
};

const onclickQuit = () => {
  // @ts-ignore - Wails runtime
  if (window.runtime) {
    // @ts-ignore
    window.runtime.Quit();
  }
};

// 模拟启动流程
onMounted(() => {
  // 阻止右键菜单（可选）
  document.addEventListener('contextmenu', (e) => {
    e.preventDefault();
  });
});
</script>

<template>
  <div id="app">
    <LaunchView v-if="!isLaunched" @finish="isLaunched = true" />
    <MainView v-else />
  </div>
</template>

<style lang="scss">
@import url("./assets/css/reset.css");
@import url("./assets/css/font.css");

html {
  width: 100%;
  height: 100%;
}

body {
  width: 100%;
  height: 100%;
  margin: 0;
  padding: 0;
  font-family: 'JetBrainsMono';
  background-color: transparent;
}

#app {
  position: relative;
  width: 100%;
  height: 100%;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  overflow: hidden;
}
</style>
