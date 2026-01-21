<template>
  <div class="page-container">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>订阅管理</span>
          <el-button type="primary" :icon="Plus">添加订阅</el-button>
        </div>
      </template>
      <el-empty description="暂无订阅，功能开发中..." />
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { Plus } from '@element-plus/icons-vue'
</script>

<style scoped>
.page-container {
  padding: 0;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
</style>
