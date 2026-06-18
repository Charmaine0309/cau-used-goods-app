<template>
  <view class="page">
    <view v-if="!fixedRole" class="tabs">
      <view class="tab" :class="{ active: role === 'buyer' }" @click="switchRole('buyer')">我买到的</view>
      <view class="tab" :class="{ active: role === 'seller' }" @click="switchRole('seller')">我卖出的</view>
    </view>

    <view v-if="orders.length" class="list">
      <view v-for="item in orders" :key="item.id" class="order">
        <view class="order-head">
          <text class="order-no">订单 {{ item.orderNo }}</text>
          <text class="status" :class="item.status">{{ getOrderStatusText(item.status) }}</text>
        </view>
        <view class="body" @click="openOrder(item)">
          <image v-if="item.productImage" class="cover" :src="normalizeImage(item.productImage)" mode="aspectFill" />
          <view v-else class="cover placeholder">暂无图片</view>
          <view class="content">
            <view class="name">{{ item.productTitleSnapshot }}</view>
            <view class="meta">{{ role === 'buyer' ? '卖家' : '买家' }}：{{ peerName(item) }}</view>
            <view class="meta">面交：{{ item.meetLocation || '待协商' }}</view>
            <view class="price">￥{{ formatPrice(item.productPriceSnapshot) }}</view>
          </view>
        </view>
        <view v-if="item.remark" class="remark">备注：{{ item.remark }}</view>
        <view class="foot">
          <text class="time">{{ formatDateTime(item.createTime) }}</text>
          <view class="actions">
            <button v-if="role === 'seller' && item.status === 'PENDING_CONFIRM'" class="action primary" @click.stop="confirm(item)">确认</button>
            <button v-if="role === 'seller' && item.status === 'WAIT_MEET'" class="action primary" @click.stop="complete(item)">完成</button>
            <button
              v-if="role === 'buyer' && item.status === 'COMPLETED'"
              class="action primary"
              :class="{ disabled: isReviewed(item) }"
              :disabled="isReviewed(item)"
              @click.stop="review(item)"
            >
              {{ isReviewed(item) ? '已评价' : '去评价' }}
            </button>
            <button v-if="canCancel(item)" class="action warn" @click.stop="cancel(item)">取消</button>
          </view>
        </view>
      </view>
    </view>

    <view v-else-if="!loading" class="empty">
      <view class="empty-title">{{ role === 'buyer' ? '还没有买到商品' : '还没有卖出商品' }}</view>
      <view class="muted">{{ role === 'buyer' ? '去首页逛逛校园同学的新鲜发布' : '有买家预约后会显示在这里' }}</view>
      <button v-if="role === 'buyer'" class="empty-button" @click="goHome">去逛逛</button>
    </view>
    <view class="load-state">{{ loading ? '正在加载...' : finished && orders.length ? '已经到底啦' : '' }}</view>
  </view>
</template>

<script setup>
import { computed, ref } from 'vue'
import { onLoad, onPullDownRefresh, onReachBottom, onShow } from '@dcloudio/uni-app'
import { cancelOrder, completeOrder, confirmOrder, listMyOrders } from '../../api/product'
import { formatPrice, normalizeImage } from '../../utils/product-format'
import { displayRelatedUserName } from '../../utils/user-format'

const role = ref('buyer')
const fixedRole = ref(false)
const orders = ref([])
const total = ref(0)
const page = ref(1)
const loading = ref(false)
const loaded = ref(false)
const finished = computed(() => orders.value.length >= total.value && total.value > 0)

const orderStatusMap = {
  PENDING_CONFIRM: '待卖家确认',
  WAIT_MEET: '待面交',
  COMPLETED: '已完成',
  CANCELED: '已取消',
  CANCELLED: '已取消',
  EXCEPTION_CLOSED: '异常关闭'
}

const getOrderStatusText = (status) => orderStatusMap[status] || status || '未知状态'
const pad = (value) => String(value).padStart(2, '0')
const formatDateTime = (value) => {
  if (!value) return ''
  if (typeof value === 'string' && /^\d{4}-\d{2}-\d{2} \d{2}:\d{2}/.test(value)) return value.slice(0, 16)
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())} ${pad(date.getHours())}:${pad(date.getMinutes())}`
}
const peerName = (item) => role.value === 'buyer'
  ? displayRelatedUserName(item, 'seller', 'CAU 同学')
  : displayRelatedUserName(item, 'buyer', 'CAU 同学')
const canCancel = (item) => ['PENDING_CONFIRM', 'WAIT_MEET'].includes(item.status)
const isReviewed = (item) => item.reviewed === true || uni.getStorageSync(`order-reviewed-${item.id}`) === true

const loadOrders = async (reset = false) => {
  if (loading.value || (!reset && finished.value)) return
  loading.value = true
  try {
    const nextPage = reset ? 1 : page.value
    const result = await listMyOrders({ role: role.value, page: nextPage, pageSize: 10 })
    const list = result.items || []
    orders.value = reset ? list : orders.value.concat(list)
    total.value = result.total || 0
    page.value = nextPage + 1
  } catch (error) {
    uni.showToast({ title: error.message, icon: 'none' })
  } finally {
    loading.value = false
    loaded.value = true
    uni.stopPullDownRefresh()
  }
}

const switchRole = (value) => {
  if (fixedRole.value) return
  if (role.value === value) return
  role.value = value
  uni.setNavigationBarTitle({ title: value === 'buyer' ? '我买到的' : '我卖出的' })
  loadOrders(true)
}

const openOrder = (item) => {
  if (!item?.id) {
    uni.showToast({ title: '订单信息缺失，暂时无法查看', icon: 'none' })
    return
  }
  uni.navigateTo({ url: `/pages/order/detail?id=${item.id}` })
}
const goHome = () => uni.switchTab({ url: '/pages/home/home' })
const review = (item) => {
  if (isReviewed(item)) return
  uni.navigateTo({ url: `/pages/interaction/review?orderId=${item.id}` })
}

const confirm = (item) => {
  uni.showModal({
    title: '确认预约？',
    content: '确认后订单进入待面交状态。',
    success: async ({ confirm }) => {
      if (!confirm) return
      try {
        await confirmOrder(item.id)
        uni.showToast({ title: '已确认', icon: 'success' })
        loadOrders(true)
      } catch (error) {
        uni.showToast({ title: error.message, icon: 'none' })
      }
    }
  })
}

const complete = (item) => {
  uni.showModal({
    title: '完成交易？',
    content: '请确认已完成线下面交。',
    success: async ({ confirm }) => {
      if (!confirm) return
      try {
        await completeOrder(item.id)
        uni.showToast({ title: '交易已完成', icon: 'success' })
        loadOrders(true)
      } catch (error) {
        uni.showToast({ title: error.message, icon: 'none' })
      }
    }
  })
}

const cancel = (item) => {
  uni.showModal({
    title: '取消订单',
    editable: true,
    placeholderText: '请填写取消原因',
    success: async ({ confirm, content }) => {
      if (!confirm) return
      try {
        await cancelOrder(item.id, content?.trim() || '用户取消')
        uni.showToast({ title: '已取消', icon: 'success' })
        loadOrders(true)
      } catch (error) {
        uni.showToast({ title: error.message, icon: 'none' })
      }
    }
  })
}

onLoad((options) => {
  fixedRole.value = options.role === 'buyer' || options.role === 'seller'
  role.value = options.role === 'seller' ? 'seller' : 'buyer'
  uni.setNavigationBarTitle({ title: role.value === 'buyer' ? '我买到的' : '我卖出的' })
  loadOrders(true)
})
onShow(() => {
  if (loaded.value) loadOrders(true)
})
onReachBottom(() => loadOrders())
onPullDownRefresh(() => loadOrders(true))
</script>

<style scoped>
.page { min-height: 100vh; padding: 28rpx; box-sizing: border-box; }
.tabs { display: flex; gap: 14rpx; margin-bottom: 22rpx; padding: 8rpx; border-radius: 18rpx; background: #fff; }
.tab { flex: 1; height: 66rpx; border-radius: 14rpx; color: #65706c; font-size: 27rpx; line-height: 66rpx; text-align: center; }
.tab.active { background: #23734f; color: #fff; font-weight: 700; }
.list { display: flex; flex-direction: column; gap: 18rpx; }
.order { padding: 20rpx; border-radius: 18rpx; background: #fff; box-shadow: 0 6rpx 18rpx rgba(28, 68, 52, .05); }
.order-head, .body, .foot, .actions { display: flex; align-items: center; }
.order-head, .foot { justify-content: space-between; gap: 12rpx; }
.order-no { color: #89938f; font-size: 22rpx; }
.status { flex-shrink: 0; padding: 6rpx 12rpx; border-radius: 999rpx; background: #e7f4ec; color: #23734f; font-size: 21rpx; }
.status.PENDING_CONFIRM { background: #fff2d8; color: #9a6a1d; }
.status.CANCELLED, .status.EXCEPTION_CLOSED { background: #eef0f0; color: #7a817e; }
.body { margin-top: 18rpx; align-items: flex-start; }
.cover { display: flex; width: 156rpx; height: 156rpx; margin-right: 18rpx; flex-shrink: 0; align-items: center; justify-content: center; border-radius: 14rpx; background: #e8efeb; color: #9aa5a1; font-size: 22rpx; }
.content { min-width: 0; flex: 1; }
.name { display: -webkit-box; overflow: hidden; color: #26342f; font-size: 29rpx; font-weight: 700; line-height: 1.4; -webkit-box-orient: vertical; -webkit-line-clamp: 2; }
.meta, .time, .muted, .load-state { color: #929c98; font-size: 23rpx; }
.meta { margin-top: 10rpx; }
.price { margin-top: 12rpx; color: #e36a3e; font-size: 32rpx; font-weight: 700; }
.remark { margin-top: 16rpx; padding: 14rpx 18rpx; border-radius: 12rpx; background: #f6f8f5; color: #65706c; font-size: 23rpx; }
.foot { margin-top: 18rpx; align-items: flex-start; }
.actions { justify-content: flex-end; gap: 10rpx; flex-wrap: wrap; }
.action { box-sizing: border-box; width: auto; min-width: 116rpx; height: 58rpx; padding: 0 22rpx; border-radius: 999rpx; background: #edf4f1; color: #23734f; font-size: 24rpx; line-height: 58rpx; white-space: nowrap; }
.action.primary { background: #23734f; color: #fff; }
.action.warn { background: #f4f1ed; color: #9a7745; }
.action.disabled,
.action[disabled] { background: #eef0f0; color: #9aa2a8; }
.empty { margin-top: 140rpx; padding: 44rpx 28rpx; border-radius: 18rpx; background: #fff; text-align: center; }
.empty-title { margin-bottom: 12rpx; color: #26342f; font-size: 32rpx; font-weight: 700; }
.empty-button { width: 180rpx; height: 66rpx; margin-top: 28rpx; border-radius: 999rpx; background: #23734f; color: #fff; font-size: 26rpx; line-height: 66rpx; }
.load-state { padding: 28rpx; text-align: center; }
</style>
