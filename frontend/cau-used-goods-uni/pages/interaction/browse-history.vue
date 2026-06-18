<template>
  <view class="page">
    <view class="header">
      <text class="title">&#27983;&#35272;&#21382;&#21490;</text>
      <text class="subtitle">&#25353;&#26368;&#36817;&#26597;&#30475;&#26102;&#38388;&#25490;&#24207;</text>
    </view>

    <view v-if="records.length" class="history-list">
      <view
        v-for="record in records"
        :key="record.id"
        class="history-card"
        @click="openProduct(record.productId)"
      >
        <image v-if="record.image" class="cover" :src="record.image" mode="aspectFill" />
        <view v-else class="cover placeholder">&#22270;</view>
        <view class="content">
          <view class="row">
            <text class="name">{{ record.title }}</text>
            <text class="time">{{ record.time }}</text>
          </view>
          <view class="seller">{{ record.sellerName || 'CAU \u540c\u5b66' }}</view>
          <view class="bottom">
            <text class="price">&#165;{{ record.price }}</text>
            <text class="status">{{ statusText(record.status) }}</text>
          </view>
        </view>
      </view>
    </view>

    <EmptyState
      v-else
      title="&#36824;&#27809;&#26377;&#27983;&#35272;&#35760;&#24405;"
      detail="&#25171;&#24320;&#21830;&#21697;&#35814;&#24773;&#21518;&#65292;&#20250;&#25353;&#26368;&#26032;&#27983;&#35272;&#26102;&#38388;&#26174;&#31034;&#22312;&#36825;&#37324;"
    />
  </view>
</template>

<script setup>
import { ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import EmptyState from '../../components/EmptyState.vue'
import { navigate } from '../../utils/navigation'
import { getBrowseHistory } from '../../utils/browse-history'

const records = ref([])

onShow(load)

function pad(value) {
  return String(value).padStart(2, '0')
}

function formatTime(value) {
  if (!value) return ''
  if (typeof value === 'string' && /^\d{4}-\d{2}-\d{2} \d{2}:\d{2}/.test(value)) {
    return value.slice(0, 16)
  }
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())} ${pad(date.getHours())}:${pad(date.getMinutes())}`
}

function statusText(status) {
  const map = {
    ON_SALE: '\u5728\u552e',
    LOCKED: '\u4ea4\u6613\u4e2d',
    SOLD: '\u5df2\u552e\u51fa',
    OFF_SHELF: '\u5df2\u4e0b\u67b6'
  }
  return map[status] || status || ''
}

function load() {
  records.value = getBrowseHistory().map((item) => ({
    ...item,
    time: formatTime(item.viewedAt)
  }))
}

function openProduct(productId) {
  if (!productId) return
  navigate('/pages/detail/detail', { id: productId })
}
</script>

<style scoped>
.page {
  min-height: 100vh;
  padding: 30rpx 26rpx 46rpx;
  background: #f3f8f5;
  box-sizing: border-box;
}

.header {
  margin-bottom: 24rpx;
}

.title,
.subtitle {
  display: block;
}

.title {
  color: #20352b;
  font-size: 38rpx;
  line-height: 48rpx;
  font-weight: 700;
}

.subtitle {
  margin-top: 8rpx;
  color: #98a2b3;
  font-size: 24rpx;
}

.history-list {
  display: flex;
  flex-direction: column;
  gap: 18rpx;
}

.history-card {
  display: flex;
  min-height: 180rpx;
  padding: 18rpx;
  border-radius: 24rpx;
  background: #fff;
  box-shadow: 0 12rpx 30rpx rgba(31, 106, 73, .05);
  box-sizing: border-box;
}

.cover {
  width: 146rpx;
  height: 146rpx;
  flex-shrink: 0;
  border-radius: 18rpx;
  background: #e8efeb;
}

.placeholder {
  display: flex;
  align-items: center;
  justify-content: center;
  color: #9aa5a1;
  font-size: 26rpx;
}

.content {
  display: flex;
  flex: 1;
  min-width: 0;
  margin-left: 20rpx;
  flex-direction: column;
}

.row,
.bottom {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.name {
  flex: 1;
  min-width: 0;
  color: #20352b;
  font-size: 30rpx;
  line-height: 38rpx;
  font-weight: 700;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.time,
.seller,
.status {
  color: #98a2b3;
  font-size: 23rpx;
}

.time {
  margin-left: 16rpx;
}

.seller {
  margin-top: 14rpx;
}

.bottom {
  margin-top: auto;
}

.price {
  color: #e36a3e;
  font-size: 32rpx;
  font-weight: 700;
}

.status {
  padding: 6rpx 14rpx;
  border-radius: 999rpx;
  background: #edf4f1;
  color: #23734f;
}
</style>
