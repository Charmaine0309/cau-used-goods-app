<template>
  <view class="page">
    <view class="avatar-row">
      <image v-if="avatarUrl" class="avatar" :src="avatarUrl" mode="aspectFill" />
      <view v-else class="avatar placeholder">头像</view>
      <button class="ghost-button" size="mini" :loading="uploading" @click="chooseAvatar">上传头像</button>
    </view>

    <view class="form-card">
      <view class="label">昵称</view>
      <input class="input" v-model="nickname" placeholder="请输入昵称" />
      <view class="label">手机号</view>
      <input class="input" v-model="phone" type="number" maxlength="11" placeholder="请输入手机号" />
      <button class="primary-button" :loading="loading" @click="saveProfile">保存资料</button>
    </view>

    <button class="wechat-button" @click="askUseWechatProfile">沿用微信昵称和头像</button>
  </view>
</template>

<script setup>
import { ref } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { getUser, setUser } from '../../utils/auth'
import { updateProfile, uploadAvatar } from '../../api/auth'

const BASE_URL = 'http://127.0.0.1:8080'
const nickname = ref('')
const phone = ref('')
const avatarUrl = ref('')
const loading = ref(false)
const uploading = ref(false)

const normalizeAvatar = (url) => {
  if (!url) return ''
  if (url.startsWith('http://') || url.startsWith('https://')) return url
  if (url.startsWith('/uploads/')) return BASE_URL + url
  return url
}

onLoad(() => {
  const user = getUser() || {}
  nickname.value = user.nickname || ''
  phone.value = user.phone || ''
  avatarUrl.value = normalizeAvatar(user.avatarUrl)
})

const askUseWechatProfile = () => {
  uni.showModal({
    title: '使用微信资料',
    content: '是否沿用你的微信昵称和头像？',
    confirmText: '使用',
    cancelText: '不用',
    success: (res) => {
      if (res.confirm) useWechatProfile()
    }
  })
}

const useWechatProfile = () => {
  if (!uni.getUserProfile) {
    uni.showToast({ title: '当前工具不支持获取微信资料', icon: 'none' })
    return
  }
  uni.getUserProfile({
    desc: '用于完善个人资料',
    success: async (res) => {
      const info = res.userInfo || {}
      nickname.value = info.nickName || nickname.value
      avatarUrl.value = info.avatarUrl || avatarUrl.value
      try {
        const user = await updateProfile({
          nickname: nickname.value,
          avatarUrl: info.avatarUrl || ''
        })
        setUser(user)
        avatarUrl.value = normalizeAvatar(user.avatarUrl)
        uni.showToast({ title: '已使用微信资料', icon: 'success' })
      } catch (error) {
        uni.showToast({ title: error.message || '保存微信资料失败', icon: 'none' })
      }
    },
    fail: () => uni.showToast({ title: '已取消使用微信资料', icon: 'none' })
  })
}

const chooseAvatar = () => {
  uni.chooseMedia({
    count: 1,
    mediaType: ['image'],
    sourceType: ['album', 'camera'],
    success: async (res) => {
      const filePath = res.tempFiles && res.tempFiles[0] && res.tempFiles[0].tempFilePath
      if (!filePath) return
      uploading.value = true
      uni.showLoading({ title: '上传中' })
      try {
        const data = await uploadAvatar(filePath)
        if (data.user) {
          setUser(data.user)
          avatarUrl.value = normalizeAvatar(data.user.avatarUrl)
        }
        uni.showToast({ title: '头像已更新', icon: 'success' })
      } catch (error) {
        uni.showToast({ title: error.message || '头像上传失败', icon: 'none' })
      } finally {
        uploading.value = false
        uni.hideLoading()
      }
    }
  })
}

const saveProfile = async () => {
  if (loading.value) return
  const nextNickname = nickname.value.trim()
  const nextPhone = phone.value.trim()
  if (!nextNickname && !nextPhone) {
    uni.showToast({ title: '请填写昵称或手机号', icon: 'none' })
    return
  }
  if (nextPhone && !/^1[3-9]\d{9}$/.test(nextPhone)) {
    uni.showToast({ title: '手机号格式不正确', icon: 'none' })
    return
  }
  const payload = {}
  if (nextNickname) payload.nickname = nextNickname
  if (nextPhone) payload.phone = nextPhone
  loading.value = true
  try {
    const user = await updateProfile(payload)
    setUser(user)
    uni.showToast({ title: '资料已保存', icon: 'success' })
    setTimeout(() => uni.navigateBack(), 600)
  } catch (error) {
    uni.showToast({ title: error.message || '保存失败', icon: 'none' })
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.page { min-height: 100vh; padding: 32rpx; background: #f5f6f8; box-sizing: border-box; }
.avatar-row { display: flex; align-items: center; justify-content: space-between; gap: 24rpx; margin-bottom: 28rpx; }
.avatar { width: 120rpx; height: 120rpx; border-radius: 60rpx; background: #dce3ea; display: flex; align-items: center; justify-content: center; color: #8b98a7; }
.form-card { padding: 32rpx; border-radius: 16rpx; background: #ffffff; }
.label { margin: 24rpx 0 12rpx; color: #667085; font-size: 28rpx; }
.input { height: 88rpx; padding: 0 24rpx; border-radius: 12rpx; background: #f0f3f7; font-size: 30rpx; box-sizing: border-box; }
.primary-button { margin-top: 36rpx; height: 88rpx; line-height: 88rpx; border-radius: 12rpx; background: #17a84b; color: #ffffff; font-size: 30rpx; }
.ghost-button { flex-shrink: 0; margin: 0; background: #ffffff; color: #17a84b; }
.wechat-button { background: #ffffff; color: #17a84b; }
.wechat-button { margin-top: 24rpx; }
</style>
