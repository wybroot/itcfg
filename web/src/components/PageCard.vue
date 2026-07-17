<template>
  <el-card class="page-card" shadow="never">
    <template v-if="title || subtitle || $slots.actions" #header>
      <div class="card-header">
        <div>
          <div v-if="title" class="card-title">{{ title }}</div>
          <div v-if="subtitle" class="card-subtitle">{{ subtitle }}</div>
        </div>
        <div v-if="$slots.actions" class="card-actions">
          <slot name="actions" />
        </div>
      </div>
    </template>
    <slot />
  </el-card>
</template>

<script setup lang="ts">
defineProps<{
  title?: string
  subtitle?: string
}>()
</script>

<style scoped>
.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
}

.card-title {
  color: var(--itcfg-text-primary);
  font-size: 16px;
  font-weight: 650;
}

.card-subtitle {
  margin-top: 4px;
  color: var(--itcfg-text-secondary);
  font-size: 13px;
}

.card-actions {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
  justify-content: flex-end;
}
</style>
