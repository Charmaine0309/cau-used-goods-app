<template>
  <view class="page">
    <view class="profile-card">
      <image v-if="avatarUrl" class="avatar" :src="avatarUrl" mode="aspectFill" />
      <view v-else class="avatar placeholder">头像</view>
      <view class="profile-main">
        <view class="nickname">{{ user.nickname || '微信用户' }}</view>
        <view class="status">{{ identityText }}</view>
      </view>
      <button class="edit-button" size="mini" @click="goProfileEdit">修改资料</button>
    </view>

    <view class="menu-card">
      <view v-if="!isAdmin" class="menu-item" @click="goMyProducts">我发布的<text>›</text></view>
      <view v-if="!isAdmin" class="menu-item" @click="goPublish">发布闲置商品<text>›</text></view>
      <view v-if="!isAdmin" class="menu-item" @click="goSoldOrders">我卖出的<text>›</text></view>
      <view v-if="!isAdmin" class="menu-item" @click="goBoughtOrders">我买到的<text>›</text></view>
      <view v-if="!isAdmin" class="menu-item" @click="goFavorites">我的收藏<text>›</text></view>
      <view v-if="!isAdmin" class="menu-item" @click="goMessages">消息中心<text>›</text></view>
      <view v-if="!isAdmin" class="menu-item" @click="goReportList">我的举报<text>›</text></view>
      <view v-if="!isAdmin" class="menu-item" @click="goStudentAuth">学生认证<text>›</text></view>
      <view v-if="!isAdmin" class="menu-item" @click="goAddress">地址管理<text>›</text></view>
      <view v-if="isAdmin" class="menu-item" @click="goAdmin">后台管理<text>›</text></view>
    </view>

    <button class="logout-button" @click="logout">退出登录</button>
  </view>
</template>

<script setup>
import { computed, ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { getCurrentUser } from '../../api/auth'
import { clearAuth, getUser, setUser } from '../../utils/auth'
import { BASE_URL } from '../../utils/request'

const user = ref(getUser() || {})

const authMap = {
  UNVERIFIED: '未认证',
  PENDING: '审核中',
  VERIFIED: '已认证',
  REJECTED: '已驳回'
}

const authText = computed(() => authMap[user.value?.authStatus] || '未认证')
const isAdmin = computed(() => user.value?.role === 'ADMIN')
const identityText = computed(() => isAdmin.value ? '管理员' : `学生认证：${authText.value}`)

const avatarUrl = computed(() => {
  const url = user.value?.avatarUrl || ''
  if (!url) return ''
  if (url.startsWith('http://') || url.startsWith('https://')) return url
  if (url.startsWith('/uploads/')) return BASE_URL + url
  return url
})

onShow(async () => {
  try {
    const current = await getCurrentUser()
    user.value = current
    setUser(current)
  } catch (error) {
    if (error.message) {
      uni.showToast({ title: error.message, icon: 'none' })
    }
  }
})

const goProfileEdit = () => uni.navigateTo({ url: '/pages/profile-edit/profile-edit' })
const goMyProducts = () => uni.navigateTo({ url: '/pages/my-products/my-products' })
const goSoldOrders = () => uni.navigateTo({ url: '/pages/my-orders/my-orders?role=seller' })
const goBoughtOrders = () => uni.navigateTo({ url: '/pages/my-orders/my-orders?role=buyer' })
const goPublish = () => uni.switchTab({ url: '/pages/publish/publish' })
const goStudentAuth = () => uni.navigateTo({ url: '/pages/student-auth/student-auth' })
const goAddress = () => uni.navigateTo({ url: '/pages/address/address' })
const goAdmin = () => uni.navigateTo({ url: '/pages/admin/admin' })
const goFavorites = () => uni.navigateTo({ url: '/pages/interaction/favorites' })
const goMessages = () => uni.switchTab({ url: '/pages/messages/messages' })
const goReportList = () => uni.navigateTo({ url: '/pages/interaction/report-list' })

const logout = () => {
  clearAuth()
  uni.reLaunch({ url: '/pages/login/login' })
}
</script>

<style scoped>
.page {
  min-height: 100vh;
  padding: 24rpx;
  background: #f5f6f8;
  box-sizing: border-box;
}

.profile-card {
  display: flex;
  align-items: center;
  min-height: 156rpx;
  padding: 28rpx;
  border-radius: 16rpx;
  background: #fff;
}

.avatar {
  display: flex;
  width: 104rpx;
  height: 104rpx;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  border-radius: 52rpx;
  background: #dce3ea;
  color: #8b98a7;
  font-size: 24rpx;
}

.profile-main {
  flex: 1;
  margin-left: 24rpx;
}

.nickname {
  font-size: 34rpx;
  font-weight: 700;
  color: #1f2933;
}

.status {
  margin-top: 12rpx;
  font-size: 26rpx;
  color: #667085;
}

.edit-button {
  background: #eef7f0;
  color: #17a84b;
}

.menu-card {
  margin-top: 24rpx;
  border-radius: 16rpx;
  background: #fff;
  overflow: hidden;
}

.menu-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 104rpx;
  line-height: 104rpx;
  padding: 0 28rpx;
  border-bottom: 1rpx solid #eef0f3;
  color: #1f2933;
  font-size: 30rpx;
  font-weight: 600;
  box-sizing: border-box;
}

.menu-item:last-child {
  border-bottom: 0;
}

.menu-item text {
  color: #b2bdca;
  font-size: 38rpx;
}

.logout-button {
  margin-top: 28rpx;
  height: 88rpx;
  line-height: 88rpx;
  border-radius: 12rpx;
  background: #fff;
  color: #ef4444;
  font-size: 30rpx;
}
</style>
