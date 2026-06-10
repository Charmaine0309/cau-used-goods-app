<template>
  <view class="page">
    <view class="section-title">订单管理</view>

    <scroll-view class="filter-row" scroll-x enhanced :show-scrollbar="false">
      <view class="filter-inner">
        <view
          v-for="item in statusFilters"
          :key="item.value"
          :class="['filter-chip', activeStatus === item.value ? 'active' : '']"
          @tap.stop="selectStatus(item.value)"
        >
          {{ item.label }} {{ countByStatus(item.value) }}
        </view>
      </view>
    </scroll-view>

    <view v-if="loading" class="empty">加载中</view>
    <view v-else-if="filteredOrders.length === 0" class="empty">暂无订单</view>

    <view v-for="item in filteredOrders" :key="item.id" class="order-card">
      <view class="card-head">
        <view>
          <view class="title">{{ item.productTitleSnapshot || '订单商品' }}</view>
          <view class="sub">订单 #{{ item.id }} · {{ item.orderNo || '暂无订单号' }}</view>
        </view>
        <view :class="['status-badge', item.status]">{{ statusText(item.status) }}</view>
      </view>

      <view class="product-row">
        <image v-if="productImage(item)" class="product-image" :src="productImage(item)" mode="aspectFill" />
        <view v-else class="product-image placeholder">商品</view>
        <view class="product-main">
          <view class="price">￥{{ money(item.productPriceSnapshot) }}</view>
          <view class="meta">买家：{{ orderUserName(item, 'buyer') }}</view>
          <view class="meta">卖家：{{ orderUserName(item, 'seller') }}</view>
        </view>
      </view>

      <view class="detail-row"><text>面交时间</text><text>{{ formatTime(item.meetTime) || '未填写' }}</text></view>
      <view class="detail-row"><text>面交地点</text><text>{{ item.meetLocation || '未填写' }}</text></view>
      <view class="detail-row"><text>创建时间</text><text>{{ formatTime(item.createTime) }}</text></view>
      <view v-if="item.remark" class="detail-row"><text>备注</text><text>{{ item.remark }}</text></view>
      <view v-if="item.cancelReason" class="handle-result">{{ item.cancelReason }}</view>

      <view class="actions">
        <button v-if="canExceptionClose(item.status)" size="mini" class="reject" @click="openCloseModal(item)">异常关闭</button>
        <button v-else size="mini" class="muted" disabled>不可关闭</button>
      </view>
    </view>

    <view v-if="closeModal.visible" class="modal-mask" @click="closeCloseModal">
      <view class="reason-sheet" @click.stop>
        <view class="sheet-title">异常关闭订单</view>
        <view class="sheet-sub">订单 #{{ closeModal.order?.id || '' }}</view>
        <view class="reason-list">
          <view
            v-for="reason in closeReasons"
            :key="reason"
            :class="['reason-chip', closeModal.reason === reason ? 'active' : '']"
            @click="closeModal.reason = reason"
          >
            {{ reason }}
          </view>
        </view>
        <textarea v-model="closeModal.note" class="reason-input" maxlength="200" placeholder="补充说明，可不填" placeholder-class="reason-placeholder" />
        <view class="sheet-actions">
          <button class="sheet-button cancel" @click="closeCloseModal">取消</button>
          <button class="sheet-button confirm" :loading="submitting" @click="submitExceptionClose">确认关闭</button>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup>
import { computed, reactive, ref } from 'vue'
import { onLoad, onShow } from '@dcloudio/uni-app'
import { exceptionCloseAdminOrder, getAdminOrders } from '../../api/admin'
import { normalizeImage } from '../../utils/product-format'
import { displayRelatedUserName } from '../../utils/user-format'

const loading = ref(false)
const submitting = ref(false)
const activeStatus = ref('ALL')
const orders = ref([])
const relatedType = ref('')
const relatedId = ref(0)
const closeModal = reactive({ visible: false, order: null, reason: '', note: '' })

const statusFilters = [
  { label: '全部', value: 'ALL' },
  { label: '待确认', value: 'PENDING_CONFIRM' },
  { label: '待面交', value: 'WAIT_MEET' },
  { label: '已完成', value: 'COMPLETED' },
  { label: '已取消', value: 'CANCELED' },
  { label: '异常关闭', value: 'EXCEPTION_CLOSED' }
]

const closeReasons = ['买卖双方协商取消', '交易存在纠纷', '商品违规或信息异常', '长时间未完成交易', '其他原因']

const filteredOrders = computed(() => {
  if (activeStatus.value === 'ALL') return orders.value
  return orders.value.filter((item) => item.status === activeStatus.value)
})

const load = async () => {
  loading.value = true
  try {
    const result = await getAdminOrders('ALL')
    orders.value = result?.items || []
  } catch (error) {
    uni.showToast({ title: error.message || '订单加载失败', icon: 'none' })
  } finally {
    loading.value = false
  }
}

onLoad((query) => {
  relatedType.value = String(query.relatedType || '').toUpperCase()
  relatedId.value = Number(query.relatedId || 0)
})

onShow(load)

const selectStatus = (status) => { activeStatus.value = status }
const countByStatus = (status) => status === 'ALL' ? orders.value.length : orders.value.filter((item) => item.status === status).length
const statusText = (status) => ({ PENDING_CONFIRM: '待确认', WAIT_MEET: '待面交', COMPLETED: '已完成', CANCELED: '已取消', CANCELLED: '已取消', EXCEPTION_CLOSED: '异常关闭' }[status] || status || '未知')
const canExceptionClose = (status) => ['PENDING_CONFIRM', 'WAIT_MEET'].includes(status)
const money = (value) => Number(value || 0).toFixed(2)
const productImage = (item) => normalizeImage(item?.productImage || '')
const formatTime = (value) => value ? String(value).replace('T', ' ').replace(/\+\d{2}:\d{2}$/, '').slice(0, 16) : ''
const orderUserName = (item, role) => displayRelatedUserName(item, role, `用户${item?.[`${role}Id`] || ''}`)

const openCloseModal = (order) => {
  closeModal.visible = true
  closeModal.order = order
  closeModal.reason = ''
  closeModal.note = ''
}

const closeCloseModal = () => {
  if (submitting.value) return
  closeModal.visible = false
}

const buildReason = () => {
  const note = String(closeModal.note || '').trim()
  return note ? `${closeModal.reason}。补充说明：${note}` : closeModal.reason
}

const submitExceptionClose = async () => {
  if (!closeModal.reason) {
    uni.showToast({ title: '请选择关闭原因', icon: 'none' })
    return
  }
  if (!closeModal.order || submitting.value) return
  submitting.value = true
  try {
    const payload = { reason: buildReason() }
    if (relatedType.value && relatedId.value) {
      payload.relatedType = relatedType.value
      payload.relatedId = relatedId.value
    }
    await exceptionCloseAdminOrder(closeModal.order.id, payload)
    uni.showToast({ title: '已异常关闭', icon: 'success' })
    closeModal.visible = false
    await load()
  } catch (error) {
    uni.showToast({ title: error.message || '操作失败', icon: 'none' })
  } finally {
    submitting.value = false
  }
}
</script>

<style scoped>
.page { min-height: 100vh; padding: 24rpx; background: #f5f6f8; box-sizing: border-box; }
.section-title { margin: 24rpx 0 18rpx; font-size: 32rpx; font-weight: 700; color: #1f2933; }
.filter-row { width: 100%; margin-bottom: 22rpx; white-space: nowrap; }
.filter-inner { display: inline-flex; gap: 14rpx; min-width: 100%; }
.filter-chip { flex-shrink: 0; padding: 14rpx 22rpx; border-radius: 999rpx; background: #fff; color: #667085; font-size: 26rpx; }
.filter-chip.active { background: #17a84b; color: #fff; font-weight: 700; }
.order-card, .empty { padding: 28rpx; border-radius: 16rpx; background: #fff; margin-bottom: 18rpx; }
.card-head { display: flex; align-items: flex-start; justify-content: space-between; gap: 20rpx; }
.title { font-size: 32rpx; font-weight: 700; color: #1f2933; }
.sub { margin-top: 8rpx; font-size: 22rpx; color: #98a2b3; }
.status-badge { flex-shrink: 0; padding: 8rpx 14rpx; border-radius: 999rpx; background: #eef2f6; color: #667085; font-size: 22rpx; }
.status-badge.PENDING_CONFIRM, .status-badge.WAIT_MEET { background: #fee2e2; color: #ef4444; }
.status-badge.COMPLETED { background: #dcfce7; color: #16a34a; }
.product-row { display: flex; gap: 18rpx; margin-top: 22rpx; padding: 18rpx; border-radius: 14rpx; background: #f8fafc; }
.product-image { width: 112rpx; height: 112rpx; border-radius: 12rpx; background: #eef2f6; flex-shrink: 0; }
.placeholder { display: flex; align-items: center; justify-content: center; color: #98a2b3; font-size: 22rpx; }
.product-main { flex: 1; min-width: 0; }
.price { font-size: 30rpx; font-weight: 700; color: #1f2933; }
.meta { margin-top: 8rpx; font-size: 24rpx; color: #667085; }
.detail-row { display: flex; justify-content: space-between; gap: 24rpx; padding-top: 16rpx; font-size: 24rpx; color: #667085; }
.detail-row text:last-child { flex: 1; text-align: right; color: #1f2933; word-break: break-all; }
.handle-result { margin-top: 18rpx; padding: 16rpx; border-radius: 12rpx; background: #f8fafc; color: #667085; font-size: 24rpx; }
.actions { margin-top: 22rpx; display: flex; gap: 14rpx; }
.actions button { margin: 0; }
.reject { background: #fff1f2; color: #ef4444; }
.muted { background: #f8fafc; color: #98a2b3; }
.modal-mask { position: fixed; left: 0; right: 0; top: 0; bottom: 0; z-index: 99; display: flex; align-items: flex-end; background: rgba(15, 23, 42, 0.42); }
.reason-sheet { width: 100%; padding: 30rpx 28rpx 36rpx; border-radius: 28rpx 28rpx 0 0; background: #fff; box-sizing: border-box; }
.sheet-title { font-size: 34rpx; font-weight: 700; color: #1f2933; }
.sheet-sub { margin-top: 8rpx; color: #98a2b3; font-size: 24rpx; }
.reason-list { display: flex; flex-wrap: wrap; gap: 14rpx; margin-top: 26rpx; }
.reason-chip { padding: 14rpx 18rpx; border-radius: 999rpx; background: #f8fafc; color: #475467; font-size: 24rpx; border: 2rpx solid transparent; }
.reason-chip.active { border-color: #17a84b; background: #f0fdf4; color: #16a34a; font-weight: 700; }
.reason-input { width: 100%; min-height: 150rpx; margin-top: 24rpx; padding: 20rpx; border-radius: 16rpx; background: #f8fafc; font-size: 26rpx; color: #1f2933; box-sizing: border-box; }
.reason-placeholder { color: #98a2b3; }
.sheet-actions { display: grid; grid-template-columns: 1fr 1fr; gap: 18rpx; margin-top: 24rpx; }
.sheet-button { height: 84rpx; line-height: 84rpx; border-radius: 14rpx; font-size: 28rpx; }
.sheet-button.cancel { background: #f8fafc; color: #667085; }
.sheet-button.confirm { background: #17a84b; color: #fff; }
</style>
