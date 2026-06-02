<template>
  <view class="empty">
    <text class="empty-icon">{{ icon }}</text>
    <text class="empty-title">{{ title }}</text>
    <text class="empty-detail">{{ detail }}</text>
  </view>
</template>

<script setup>
defineProps({
  icon: { type: String, default: '○' },
  title: { type: String, default: '暂时没有内容' },
  detail: { type: String, default: '稍后再来看看' }
})
</script>

<style scoped lang="scss">
.empty {
  display: flex;
  align-items: center;
  flex-direction: column;
  padding: 120rpx 30rpx;
  text-align: center;
}

.empty-icon { color: #a8b5ae; font-size: 72rpx; }
.empty-title { margin-top: 16rpx; color: #425148; font-size: 30rpx; font-weight: 700; }
.empty-detail { margin-top: 10rpx; color: #8b9690; font-size: 25rpx; }
</style>
