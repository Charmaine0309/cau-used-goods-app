<template>
  <view class="page">
    <view v-if="product" class="card">
      <ProductRow :product="product" />
    </view>
    <view class="notice">平台不提供线上支付和物流服务，请选择校园内公共区域完成面交。</view>
    <view class="card">
      <view class="field">
        <text class="field-label">期望面交时间</text>
        <input v-model="form.meetTime" class="input appointment-input" placeholder="格式：2026-06-10 18:30" />
        <text class="field-tip">请填写完整日期和时间，例如：2026-06-10 18:30</text>
      </view>
      <view class="field">
        <text class="field-label">面交地点</text>
        <input v-model="form.meetLocation" class="input appointment-input" placeholder="如：东区图书馆门口" />
      </view>
      <view class="field">
        <text class="field-label">备注（选填）</text>
        <textarea v-model="form.remark" class="textarea" placeholder="可填写时间补充或其他说明" maxlength="200" />
      </view>
      <button class="btn btn-primary" :disabled="submitting" @click="submit">
        {{ submitting ? '正在提交...' : '确认提交预约' }}
      </button>
    </view>
  </view>
</template>

<script setup>
import { onLoad } from '@dcloudio/uni-app'
import { reactive, ref } from 'vue'
import ProductRow from '../../components/ProductRow.vue'
import { tradeService } from '../../services/trade'
import { navigate, showError, showSuccess } from '../../utils/navigation'

const product = ref()
const submitting = ref(false)
const productId = ref('')
const form = reactive({ meetTime: '', meetLocation: '', remark: '' })

onLoad(async (options) => {
  productId.value = options.productId || 'p-1001'
  try {
    product.value = await tradeService.getProduct(productId.value)
    form.meetLocation = product.value.meetLocation
  } catch (error) {
    showError(error)
  }
})

async function submit() {
  if (!form.meetTime || !form.meetLocation) {
    showError(new Error('请填写面交时间和地点'))
    return
  }
  const meetTime = normalizeMeetTime(form.meetTime)
  if (!meetTime) {
    showError(new Error('时间格式应为：2026-06-10 18:30'))
    return
  }
  submitting.value = true
  try {
    const order = await tradeService.createAppointment({ productId: productId.value, ...form, meetTime })
    showSuccess('预约成功')
    setTimeout(() => navigate('/pages/order/detail', { id: order.id }), 500)
  } catch (error) {
    showError(error)
  } finally {
    submitting.value = false
  }
}

function normalizeMeetTime(value) {
  const text = String(value || '').trim()
  if (/^\d{4}-\d{2}-\d{2} \d{2}:\d{2}$/.test(text)) return `${text}:00`
  if (/^\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}$/.test(text)) return text
  return ''
}
</script>

<style scoped lang="scss">
.appointment-input { min-height: 82rpx; line-height: 82rpx; }
.field-tip { display: block; margin-top: 10rpx; color: #8a9690; font-size: 22rpx; line-height: 1.5; }
</style>
