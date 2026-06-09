<template>
  <view v-if="order" class="page">
    <view class="card status-card">
      <view class="status-copy">
        <text class="status-title">{{ status.label }}</text>
        <text class="status-tip">{{ statusTip }}</text>
      </view>
      <StatusBadge :label="status.label" :tone="status.tone" />
    </view>

    <view class="card">
      <ProductRow :product="order.product" />
    </view>

    <view class="card info">
      <view class="info-row"><text class="info-label">订单编号</text><text class="info-value">{{ order.id }}</text></view>
      <view class="info-row"><text class="info-label">预约时间</text><text class="info-value">{{ order.meetTime }}</text></view>
      <view class="info-row"><text class="info-label">面交地点</text><text class="info-value">{{ order.meetLocation }}</text></view>
      <view class="info-row"><text class="info-label">备注</text><text class="info-value">{{ order.remark || '无' }}</text></view>
      <view v-if="order.expireTime && order.status === 'PENDING_CONFIRM'" class="info-row">
        <text class="info-label">确认时限</text><text class="info-value">{{ order.expireTime }}</text>
      </view>
    </view>

    <view v-if="order.status === 'WAIT_MEET'" class="notice">
      为保护隐私，联系方式仅在待面交阶段向交易双方展示：{{ order.contact || '后端暂未返回联系方式字段' }}
    </view>

    <view class="actions">
      <button v-if="isSeller && order.status === 'PENDING_CONFIRM'" class="btn btn-primary" @click="change('confirm')">确认预约</button>
      <button v-if="isSeller && order.status === 'WAIT_MEET'" class="btn btn-primary" @click="change('complete')">确认完成交易</button>
      <button v-if="canCancel" class="btn btn-plain" @click="cancel">取消订单</button>
      <button v-if="order.status === 'COMPLETED' && !order.reviewed && !isSeller" class="btn btn-primary" @click="review">去评价</button>
      <button class="btn btn-plain" @click="report">举报交易问题</button>
    </view>
  </view>
</template>

<script setup>
import { onLoad } from '@dcloudio/uni-app'
import { computed, ref } from 'vue'
import ProductRow from '../../components/ProductRow.vue'
import StatusBadge from '../../components/StatusBadge.vue'
import { tradeService } from '../../services/trade'
import { getUser } from '../../utils/auth'
import { ORDER_STATUS } from '../../utils/constants'
import { navigate, showError, showSuccess } from '../../utils/navigation'

const order = ref()
const currentUserId = computed(() => String(getUser()?.id || ''))
const status = computed(() => ORDER_STATUS[order.value?.status] || { label: '', tone: 'muted' })
const isSeller = computed(() => String(order.value?.sellerId) === currentUserId.value)
const canCancel = computed(() => ['PENDING_CONFIRM', 'WAIT_MEET'].includes(order.value?.status))
const statusTip = computed(() => ({
  PENDING_CONFIRM: '卖家需在 24 小时内处理预约',
  WAIT_MEET: '请按约定时间在校园内公共区域完成面交',
  COMPLETED: '线下交易已完成',
  CANCELED: '订单已关闭，商品将按规则恢复在售',
  CANCELLED: '订单已关闭，商品将按规则恢复在售',
  EXCEPTION_CLOSED: '订单由管理员介入关闭'
}[order.value?.status] || ''))

let id = ''
onLoad((options) => {
  id = options.id
  load()
})

async function load() {
  try {
    order.value = await tradeService.getOrder(id)
  } catch (error) {
    showError(error)
  }
}

async function change(action, payload) {
  try {
    order.value = await tradeService.changeOrderStatus(id, action, payload)
    showSuccess('操作成功')
  } catch (error) {
    showError(error)
  }
}

function cancel() {
  uni.showModal({
    title: '取消订单',
    content: '确认取消当前订单吗？取消后商品将按规则恢复在售。',
    success: ({ confirm }) => confirm && change('cancel', { reason: '用户主动取消' })
  })
}

function review() {
  navigate('/pages/interaction/review', { orderId: id })
}

function report() {
  navigate('/pages/interaction/report', { targetType: 'ORDER', targetId: id })
}
</script>

<style scoped lang="scss">
.page {
  box-sizing: border-box;
  width: 100%;
  max-width: 100vw;
  overflow-x: hidden;
  padding: 20rpx 24rpx 48rpx;
}

.card {
  box-sizing: border-box;
  width: 100%;
  max-width: 100%;
  margin-bottom: 20rpx;
  padding: 24rpx;
  border-radius: 20rpx;
  background: #fff;
}

.status-card {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 18rpx;
}

.status-copy {
  display: flex;
  flex: 1;
  min-width: 0;
  flex-direction: column;
  gap: 10rpx;
}

.status-title {
  font-size: 36rpx;
  font-weight: 700;
}

.status-tip {
  color: #738077;
  font-size: 24rpx;
  line-height: 1.5;
}

.info-row {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 28rpx;
  padding: 16rpx 0;
  border-bottom: 1rpx solid #eef1ef;
  font-size: 25rpx;
  line-height: 1.5;
}

.info-row:last-child {
  border-bottom: 0;
}

.info-label {
  flex: 0 0 150rpx;
  color: #738077;
}

.info-value {
  flex: 1;
  min-width: 0;
  color: #36443c;
  text-align: right;
  overflow-wrap: break-word;
  word-break: break-all;
}

.actions {
  box-sizing: border-box;
  width: 100%;
  max-width: 100%;
}

.actions .btn {
  box-sizing: border-box;
  min-width: 0;
}
</style>
