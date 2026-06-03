<template>
  <view class="page">
    <view class="title">CAU二手交易平台</view>

    <button class="login-button" :loading="loading" @click="handleDevLogin">
      微信登录
    </button>
  </view>
</template>

<script setup>
import { ref } from 'vue'
import { devLogin } from '../../api/auth'
import { saveLoginResult } from '../../utils/auth'

const loading = ref(false)

const goHome = () => {
  uni.reLaunch({
    url: '/pages/home/home'
  })
}

const goAdmin = () => {
  uni.reLaunch({
    url: '/pages/admin/admin'
  })
}

const handleDevLogin = async () => {
  if (loading.value) return
  loading.value = true
  try {
    const result = await devLogin({
      openid: 'frontend_a_dev_user'
    })

    saveLoginResult(result)
    uni.showToast({
      title: '登录成功',
      icon: 'success'
    })
    if (result?.user?.role === 'ADMIN') {
      goAdmin()
      return
    }
    goHome()
  } catch (error) {
    uni.showToast({
      title: error.message || '登录失败',
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
  padding: 120rpx 48rpx;
  background: #f6f7f9;
  box-sizing: border-box;
}

.title {
  margin-top: 180rpx;
  font-size: 44rpx;
  font-weight: 700;
  color: #1f2933;
  text-align: center;
}

.login-button {
  margin-top: 120rpx;
  height: 88rpx;
  line-height: 88rpx;
  border-radius: 12rpx;
  background: #1aad19;
  color: #ffffff;
  font-size: 32rpx;
}

</style>
