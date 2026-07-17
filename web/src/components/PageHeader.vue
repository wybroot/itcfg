<template>
  <div class="page-header">
    <div class="page-header-main">
      <el-button v-if="back" text class="back-button" @click="goBack">
        <el-icon><ArrowLeft /></el-icon>
        {{ backText }}
      </el-button>
      <h1 class="page-title">{{ title }}</h1>
      <div v-if="subtitle" class="page-subtitle">{{ subtitle }}</div>
    </div>
    <div v-if="$slots.actions" class="page-actions">
      <slot name="actions" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'

const props = withDefaults(defineProps<{
  title: string
  subtitle?: string
  back?: boolean | string
  backText?: string
}>(), {
  subtitle: '',
  back: false,
  backText: '返回',
})

const router = useRouter()

const goBack = () => {
  if (typeof props.back === 'string') {
    router.push(props.back)
    return
  }
  router.back()
}
</script>

<style scoped>
.back-button {
  margin: 0 0 8px -8px;
  color: var(--itcfg-text-secondary);
}
</style>
