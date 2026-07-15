<template>
  <div>
    <el-row :gutter="20">
      <el-col :span="6">
        <el-card shadow="hover">
          <template #header>
            <div style="display: flex; justify-content: space-between; align-items: center">
              <span>客户数量</span>
              <el-icon :size="24" color="#409EFF"><User /></el-icon>
            </div>
          </template>
          <div style="font-size: 32px; font-weight: bold; color: #409EFF">{{ stats.customers }}</div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover">
          <template #header>
            <div style="display: flex; justify-content: space-between; align-items: center">
              <span>组件模板</span>
              <el-icon :size="24" color="#67C23A"><Grid /></el-icon>
            </div>
          </template>
          <div style="font-size: 32px; font-weight: bold; color: #67C23A">{{ stats.components }}</div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover">
          <template #header>
            <div style="display: flex; justify-content: space-between; align-items: center">
              <span>今日部署</span>
              <el-icon :size="24" color="#E6A23C"><Clock /></el-icon>
            </div>
          </template>
          <div style="font-size: 32px; font-weight: bold; color: #E6A23C">{{ stats.todayDeploys }}</div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover">
          <template #header>
            <div style="display: flex; justify-content: space-between; align-items: center">
              <span>部署成功率</span>
              <el-icon :size="24" color="#F56C6C"><CircleCheck /></el-icon>
            </div>
          </template>
          <div style="font-size: 32px; font-weight: bold; color: #F56C6C">{{ stats.successRate }}%</div>
        </el-card>
      </el-col>
    </el-row>

    <el-card style="margin-top: 20px">
      <template #header>
        <span>快速操作</span>
      </template>
      <el-space wrap>
        <el-button type="primary" @click="$router.push('/customers')">
          <el-icon><Plus /></el-icon> 新建客户
        </el-button>
        <el-button type="success" @click="$router.push('/components')">
          <el-icon><Grid /></el-icon> 查看组件
        </el-button>
        <el-button type="warning" @click="$router.push('/deploy-records')">
          <el-icon><Clock /></el-icon> 部署记录
        </el-button>
      </el-space>
    </el-card>

    <el-card style="margin-top: 20px">
      <template #header>
        <span>使用流程</span>
      </template>
      <el-steps :active="0" align-center>
        <el-step title="创建客户" description="录入客户信息" />
        <el-step title="配置环境" description="创建生产/测试环境" />
        <el-step title="填写配置" description="按组件填写配置变量" />
        <el-step title="导出部署包" description="一键导出完整部署包" />
        <el-step title="现场部署" description="Agent 一键部署" />
      </el-steps>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { reactive } from 'vue'

const stats = reactive({
  customers: 0,
  components: 3,
  todayDeploys: 0,
  successRate: 100,
})
</script>