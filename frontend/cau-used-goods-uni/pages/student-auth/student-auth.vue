<template>
  <view class="page">
    <view class="title">学生认证</view>
    <view class="status">当前状态：{{ statusText }}</view>

    <view class="form-item">
      <text class="label">姓名</text>
      <input class="input" v-model="form.realName" placeholder="请输入姓名" />
    </view>

    <view class="form-item">
      <text class="label">学号</text>
      <input class="input" v-model="form.studentId" type="number" placeholder="请输入学号" />
    </view>

    <view class="form-item">
      <text class="label">学院</text>
      <input class="input" v-model="form.college" placeholder="请输入学院" />
    </view>

    <button class="submit-button" :loading="loading" @click="handleSubmit">
      提交认证
    </button>

    <button class="secondary-button" @click="goHome">
      返回首页
    </button>
  </view>
</template>

<script setup>
import { computed, reactive, ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { getCurrentUser, getStudentVerification, submitStudentVerification } from '../../api/auth'
import { setUser } from '../../utils/auth'

const loading = ref(false)
const authStatus = ref('UNVERIFIED')
const form = reactive({
  realName: '',
  studentId: '',
  college: ''
})

const statusMap = {
  UNVERIFIED: '未认证',
  PENDING: '审核中',
  VERIFIED: '已认证',
  REJECTED: '已驳回'
}

const statusText = computed(() => statusMap[authStatus.value] || '未认证')

onShow(async () => {
  try {
    const user = await getCurrentUser()
    setUser(user)
    authStatus.value = user.authStatus || 'UNVERIFIED'
    const verification = await getStudentVerification()
    form.realName = verification.realName || ''
    form.studentId = verification.studentId || ''
    form.college = verification.college || ''
  } catch (error) {
    if (error.message) {
      uni.showToast({ title: error.message, icon: 'none' })
    }
  }
})

const validateForm = () => {
  if (!/^[\u4e00-\u9fa5]{2,20}$/.test(form.realName.trim())) return '姓名需填写2到20个汉字'
  if (!/^\d{6,20}$/.test(form.studentId.trim())) return '学号需填写6到20位数字'
  if (!/^[\u4e00-\u9fa5]{2,30}$/.test(form.college.trim())) return '学院需填写2到30个汉字'
  return ''
}

const goHome = () => {
  uni.switchTab({
    url: '/pages/home/home'
  })
}

const handleSubmit = async () => {
  const message = validateForm()
  if (message) {
    uni.showToast({
      title: message,
      icon: 'none'
    })
    return
  }

  if (loading.value) return

  loading.value = true
  try {
    await submitStudentVerification({
      studentId: form.studentId.trim(),
      realName: form.realName.trim(),
      college: form.college.trim()
    })

    uni.showToast({
      title: '提交成功',
      icon: 'success'
    })
    const user = await getCurrentUser()
    setUser(user)
    goHome()
  } catch (error) {
    uni.showToast({
      title: error.message || '提交失败',
      icon: 'none'
    })
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.page {
  min-height: 100vh;
  padding: 48rpx;
  background: #f6f7f9;
  box-sizing: border-box;
}

.title {
  margin-bottom: 16rpx;
  font-size: 40rpx;
  font-weight: 700;
  color: #1f2933;
}

.status {
  margin-bottom: 40rpx;
  color: #6b7280;
  font-size: 28rpx;
}

.form-item {
  margin-bottom: 32rpx;
}

.label {
  display: block;
  margin-bottom: 12rpx;
  font-size: 28rpx;
  color: #374151;
}

.input {
  height: 88rpx;
  padding: 0 24rpx;
  border-radius: 12rpx;
  background: #ffffff;
  font-size: 28rpx;
}

.submit-button {
  margin-top: 56rpx;
  height: 88rpx;
  line-height: 88rpx;
  border-radius: 12rpx;
  background: #1aad19;
  color: #ffffff;
  font-size: 32rpx;
}

.secondary-button {
  margin-top: 24rpx;
  height: 88rpx;
  line-height: 88rpx;
  border-radius: 12rpx;
  background: #ffffff;
  color: #374151;
  font-size: 30rpx;
}
</style>
