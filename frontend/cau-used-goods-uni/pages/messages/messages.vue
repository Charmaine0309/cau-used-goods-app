<template>
  <view class="page">
    <view v-if="messages.length">
      <view v-for="message in messages" :key="message.id" class="card message" @click="open(message.id)">
        <view class="dot" :class="{ read: message.read }" />
        <view class="body">
          <view class="head"><text class="message-title">{{ message.title }}</text><text class="time">{{ message.createdAt }}</text></view>
          <text class="content">{{ message.content }}</text>
        </view>
      </view>
    </view>
    <EmptyState v-else title="暂无消息" detail="订单进度与举报处理结果会在这里提醒你" />
  </view>
</template>

<script setup>
import { ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import EmptyState from '../../components/EmptyState.vue'
import { tradeService } from '../../services/trade'
import { navigate, showError } from '../../utils/navigation'

const messages = ref([])
onShow(async () => {
  try { messages.value = await tradeService.getMessages() } catch (error) { showError(error) }
})
const open = (id) => navigate('/pages/interaction/message-detail', { id })
</script>

<style scoped>
.page{min-height:100vh;padding:24rpx;background:#f6f7f9}.card{margin-bottom:18rpx;padding:24rpx;border-radius:18rpx;background:#fff}.message{display:flex;gap:14rpx}.dot{width:16rpx;height:16rpx;margin-top:10rpx;border-radius:50%;background:#f2a23a}.dot.read{background:#ccd4d0}.body{flex:1;min-width:0}.head{display:flex;justify-content:space-between;gap:12rpx}.message-title{font-weight:700}.time{color:#98a39d;font-size:21rpx}.content{display:block;overflow:hidden;margin-top:10rpx;color:#738077;font-size:25rpx;text-overflow:ellipsis;white-space:nowrap}
</style>
