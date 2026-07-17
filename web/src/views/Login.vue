<template>
  <div class="login-bg">
    <!-- 背景装饰 -->
    <div class="bg-shapes">
      <div class="shape shape-1"></div>
      <div class="shape shape-2"></div>
      <div class="shape shape-3"></div>
    </div>

    <div class="login-wrapper">
      <!-- 左侧品牌区 -->
      <div class="login-brand">
        <div class="brand-icon">
          <el-icon :size="48" color="#fff"><Setting /></el-icon>
        </div>
        <h1>ITCFG 配置中台</h1>
        <p>企业级配置管理与自动化部署平台</p>
        <div class="brand-features">
          <div class="feature-item">
            <el-icon><CircleCheck /></el-icon> 多环境配置管理
          </div>
          <div class="feature-item">
            <el-icon><CircleCheck /></el-icon> 一键导出部署包
          </div>
          <div class="feature-item">
            <el-icon><CircleCheck /></el-icon> Agent 自动化部署
          </div>
          <div class="feature-item">
            <el-icon><CircleCheck /></el-icon> 版本快照与回滚
          </div>
        </div>
      </div>

      <!-- 右侧登录表单 -->
      <div class="login-card">
        <h2>欢迎登录</h2>
        <p style="color: #909399; margin-bottom: 28px">请输入您的账号密码</p>

        <el-form ref="formRef" :model="form" :rules="rules" label-width="0" size="large">
          <el-form-item prop="username">
            <el-input
              v-model="form.username"
              placeholder="用户名"
              :prefix-icon="User"
            />
          </el-form-item>
          <el-form-item prop="password">
            <el-input
              v-model="form.password"
              type="password"
              placeholder="密码"
              show-password
              :prefix-icon="Lock"
              @keyup.enter="handleLogin"
            />
          </el-form-item>
          <el-form-item>
            <el-button type="primary" style="width: 100%" @click="handleLogin" :loading="loading" round>
              登 录
            </el-button>
          </el-form-item>
        </el-form>

        <div class="login-footer">
          <span>默认管理员: admin / admin123</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { User, Lock } from '@element-plus/icons-vue'
import { login, setToken, setUser } from '../api'

const router = useRouter()
const formRef = ref<FormInstance>()
const loading = ref(false)

const form = ref({
  username: '',
  password: '',
})

const rules: FormRules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }],
}

const handleLogin = async () => {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return

  loading.value = true
  try {
    const res: any = await login(form.value.username, form.value.password)
    const data = res.data
    setToken(data.token)
    setUser({ id: data.user_id, username: data.username, nickname: data.nickname, role: data.role })
    ElMessage.success(`欢迎回来，${data.nickname || data.username}`)
    router.push('/')
  } catch {
    ElMessage.error('用户名或密码错误')
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-bg {
  height: 100vh;
  display: flex; align-items: center; justify-content: center;
  background: linear-gradient(135deg, #0f1a2e 0%, #1a3350 30%, #1e3a5f 60%, #152238 100%);
  position: relative; overflow: hidden;
}
.bg-shapes { position: absolute; inset: 0; pointer-events: none; }
.shape {
  position: absolute; border-radius: 50%;
  background: rgba(64,158,255,0.08);
}
.shape-1 { width: 600px; height: 600px; top: -200px; right: -100px; }
.shape-2 { width: 400px; height: 400px; bottom: -150px; left: -100px; }
.shape-3 { width: 300px; height: 300px; top: 40%; left: 45%; background: rgba(103,194,58,0.06); }

.login-wrapper {
  display: flex; background: #fff; border-radius: 16px;
  box-shadow: 0 24px 80px rgba(0,0,0,0.3);
  overflow: hidden; z-index: 1;
  max-width: 900px; width: 90%;
}
.login-brand {
  flex: 1; background: linear-gradient(160deg, #1e3a5f 0%, #409EFF 100%);
  padding: 48px 40px; color: #fff;
  display: flex; flex-direction: column; justify-content: center;
}
.brand-icon { margin-bottom: 20px; }
.login-brand h1 { font-size: 26px; margin: 0 0 8px; font-weight: 700; letter-spacing: 2px; }
.login-brand > p { opacity: 0.75; font-size: 14px; margin: 0 0 36px; }
.brand-features { display: flex; flex-direction: column; gap: 14px; }
.feature-item { display: flex; align-items: center; gap: 10px; font-size: 14px; opacity: 0.85; }

.login-card {
  flex: 1; padding: 48px 40px;
  display: flex; flex-direction: column; justify-content: center;
}
.login-card h2 { font-size: 24px; margin: 0 0 4px; color: #303133; font-weight: 600; }
.login-footer { text-align: center; color: #c0c4cc; font-size: 12px; margin-top: 16px; }

@media (max-width: 768px) {
  .login-brand { display: none; }
  .login-wrapper { max-width: 400px; }
}
</style>
