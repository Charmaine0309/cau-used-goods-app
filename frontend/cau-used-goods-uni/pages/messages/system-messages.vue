<template>
  <view class="page">
    <view class="header">
      <view>
        <text class="title">系统消息</text>
        <text class="subtitle">订单进度、举报处理和平台通知</text>
      </view>
      <button v-if="unreadCount" class="read-all" :disabled="marking" @click="markAllRead">全部已读</button>
    </view>

    <view v-if="messages.length" class="list">
      <view v-for="item in messages" :key="item.id" class="card" @click="open(item)">
        <view class="icon" :class="{ read: item.read }">{{ item.read ? '✓' : '!' }}</view>
        <view class="body">
          <view class="head">
            <text class="message-title">{{ item.title }}</text>
            <text class="tag" :class="{ read: item.read }">{{ item.read ? '已读' : '未读' }}</text>
          </view>
          <text class="time">{{ item.createdAt }}</text>
          <text class="content">{{ item.content }}</text>
        </view>
      </view>
    </view>

    <EmptyState v-else title="暂无系统消息" detail="有新的交易进度或平台通知时会显示在这里" />
  </view>
</template>

<script setup>
import { computed, ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import EmptyState from '../../components/EmptyState.vue'
import { tradeService } from '../../services/trade'
import { navigate, showError } from '../../utils/navigation'

const SYSTEM_TYPES = ['ORDER_CREATED', 'ORDER_CONFIRMED', 'ORDER_CANCELED', 'ORDER_TIMEOUT', 'REPORT_HANDLED', 'SYSTEM_NOTICE']
const messages = ref([])
const marking = ref(false)
const unreadCount = computed(() => messages.value.filter((item) => !item.read).length)

onShow(async () => {
  try {
    const list = await tradeService.getMessages()
    messages.value = list.filter((item) => SYSTEM_TYPES.includes(item.type || item.messageType))
  } catch (error) {
    showError(error)
  }
})

function open(item) {
  navigate('/pages/interaction/message-detail', { id: item.id })
}

async function markAllRead() {
  if (!unreadCount.value || marking.value) return
  marking.value = true
  try {
    const unread = messages.value.filter((item) => !item.read)
    await Promise.all(unread.map((item) => tradeService.markMessageRead(item.id).catch(() => null)))
    messages.value = messages.value.map((item) => ({ ...item, read: true }))
    uni.removeTabBarBadge({ index: 2 })
    uni.showToast({ title: '已全部标为已读', icon: 'success' })
  } catch (error) {
    showError(error)
  } finally {
    marking.value = false
  }
}
</script>

<style scoped>
.page { min-height: 100vh; padding: 30rpx 28rpx 48rpx; background: #f5f8f6; box-sizing: border-box; }
.header { display: flex; align-items: flex-start; justify-content: space-between; gap: 20rpx; margin-bottom: 24rpx; padding: 28rpx 26rpx; border-radius: 28rpx; background: linear-gradient(135deg, #23734f, #3e9b72); box-shadow: 0 12rpx 32rpx rgba(35, 115, 79, .18); }
.title, .subtitle { display: block; }
.title { color: #fff; font-size: 42rpx; font-weight: 800; }
.subtitle { margin-top: 10rpx; color: rgba(255,255,255,.78); font-size: 24rpx; }
.read-all { flex-shrink: 0; min-width: 144rpx; height: 62rpx; padding: 0 20rpx; border-radius: 999rpx; background: rgba(255,255,255,.94); color: #23734f; font-size: 24rpx; line-height: 62rpx; }
.read-all[disabled] { background: rgba(255,255,255,.62); color: #8aa999; }
.list { display: flex; flex-direction: column; gap: 22rpx; }
.card { display: flex; gap: 22rpx; padding: 30rpx 28rpx; border-radius: 28rpx; background: #fff; box-shadow: 0 10rpx 30rpx rgba(28, 68, 52, .06); }
.icon { display: flex; width: 54rpx; height: 54rpx; flex: 0 0 54rpx; align-items: center; justify-content: center; border-radius: 50%; background: #fff1e8; color: #e36a3e; font-size: 26rpx; font-weight: 800; }
.icon.read { background: #eef4f1; color: #7c8a84; }
.body { flex: 1; min-width: 0; }
.head { display: flex; align-items: center; justify-content: space-between; gap: 18rpx; }
.message-title { flex: 1; min-width: 0; overflow: hidden; color: #26342f; font-size: 31rpx; font-weight: 800; text-overflow: ellipsis; white-space: nowrap; }
.tag { flex-shrink: 0; padding: 6rpx 14rpx; border-radius: 999rpx; background: #fff3dd; color: #bd7a16; font-size: 21rpx; }
.tag.read { background: #eef2f0; color: #8b9691; }
.time { display: block; margin-top: 8rpx; color: #9ba5a0; font-size: 22rpx; }
.content { display: block; margin-top: 18rpx; color: #5c6862; font-size: 26rpx; line-height: 1.7; }
</style>
