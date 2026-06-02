<template>
  <view class="page">
    <view class="profile-card">
      <image class="logo" src="/static/logo.png" mode="aspectFit" />
      <view>
        <text class="name">{{ user?.nickname || '校园用户' }}</text>
        <text class="status">{{ user?.authStatus === 'VERIFIED' ? '已认证学生' : '认证状态待完善' }}</text>
      </view>
    </view>
    <view class="menu-card">
      <view class="menu-item" @click="go('/pages/order/list')"><text>我的订单</text><text>›</text></view>
      <view class="menu-item" @click="go('/pages/interaction/favorites')"><text>我的收藏</text><text>›</text></view>
      <view class="menu-item" @click="go('/pages/messages/messages')"><text>消息中心</text><text>›</text></view>
      <view class="menu-item" @click="go('/pages/interaction/report-list')"><text>我的举报</text><text>›</text></view>
      <view class="menu-item" @click="go('/pages/profile-edit/profile-edit')"><text>修改资料</text><text>›</text></view>
      <view class="menu-item" @click="go('/pages/student-auth/student-auth')"><text>学生认证</text><text>›</text></view>
    </view>
  </view>
</template>

<script setup>
import { ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { getUser } from '../../utils/auth'
import { navigate } from '../../utils/navigation'

const user = ref(null)
onShow(() => { user.value = getUser() })
const go = (path) => navigate(path)
</script>

<style scoped>
.page{min-height:100vh;padding:32rpx;background:#f6f7f9;box-sizing:border-box}.profile-card,.menu-card{border-radius:16rpx;background:#fff}.profile-card{display:flex;align-items:center;padding:28rpx}.logo{width:96rpx;height:96rpx;margin-right:20rpx}.name,.status{display:block}.name{font-size:34rpx;font-weight:700}.status{margin-top:8rpx;color:#17a84b;font-size:24rpx}.menu-card{margin-top:24rpx;padding:0 24rpx}.menu-item{display:flex;justify-content:space-between;padding:28rpx 0;border-bottom:2rpx solid #eef1ef}.menu-item:last-child{border:0}
</style>
