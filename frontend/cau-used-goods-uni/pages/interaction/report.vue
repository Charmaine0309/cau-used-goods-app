<template>
  <view class="page">
    <view class="notice">举报材料仅供管理员处理使用。凭证图片不作为公开资源展示，也可以不上传凭证直接提交。</view>
    <view class="card">
      <view class="field">
        <text class="field-label">举报对象</text>
        <text class="picker-value">{{ form.targetType === 'ORDER' ? '交易订单' : '商品' }}：{{ form.targetId }}</text>
      </view>
      <view class="field">
        <text class="field-label">举报原因</text>
        <picker :range="reasons" @change="form.reason = reasons[$event.detail.value]">
          <view class="picker-value">{{ form.reason || '请选择举报原因' }}</view>
        </picker>
      </view>
      <view class="field">
        <text class="field-label">补充说明</text>
        <textarea v-model="form.detail" class="textarea" maxlength="500" placeholder="请描述问题，便于管理员核实处理" />
      </view>
      <view class="field">
        <text class="field-label">凭证图片（选填，最多 3 张）</text>
        <view class="images">
          <image v-for="src in form.images" :key="src" :src="src" mode="aspectFill" />
          <view v-if="form.images.length < 3" class="image-add" @click="chooseImage">+</view>
        </view>
      </view>
      <button class="btn btn-primary" @click="submit">提交举报</button>
    </view>
  </view>
</template>

<script setup>
import { onLoad } from '@dcloudio/uni-app'
import { reactive } from 'vue'
import { tradeService } from '../../services/trade'
import { navigate, showError, showSuccess } from '../../utils/navigation'

const reasons = ['商品描述不实', '疑似禁售品', '交易纠纷', '不文明行为', '其他问题']
const form = reactive({ targetType: 'PRODUCT', targetId: '', reason: '', detail: '', images: [] })

onLoad((options) => {
  form.targetType = options.targetType || 'PRODUCT'
  form.targetId = options.targetId || 'p-1001'
})

function chooseImage() {
  uni.chooseImage({
    count: 3 - form.images.length,
    success: ({ tempFilePaths }) => form.images.push(...tempFilePaths)
  })
}

async function submit() {
  if (!form.reason || !form.detail.trim()) {
    showError(new Error('请选择举报原因并填写说明'))
    return
  }
  try {
    await tradeService.createReport({ ...form })
    showSuccess('举报已提交')
    setTimeout(() => navigate('/pages/interaction/report-list'), 500)
  } catch (error) {
    showError(error)
  }
}
</script>

<style scoped lang="scss">
.images { display: flex; gap: 16rpx; flex-wrap: wrap; }
.images image, .image-add { width: 144rpx; height: 144rpx; border-radius: 14rpx; }
.image-add { display: flex; align-items: center; justify-content: center; border: 1rpx dashed #b8c3bd; color: #91a098; background: #fbfcfb; font-size: 54rpx; }
</style>
