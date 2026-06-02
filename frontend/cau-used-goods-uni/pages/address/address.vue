<template>
  <view class="page">
    <view class="form-card">
      <input class="input" v-model="form.receiver" placeholder="收件人姓名" />
      <input class="input" v-model="form.phone" type="number" maxlength="11" placeholder="手机号" />
      <input class="input" v-model="form.detail" placeholder="详细地址" />
      <button class="primary-button" @click="saveAddress">保存地址</button>
    </view>

    <view class="list">
      <view v-for="item in addresses" :key="item.id" class="address-card">
        <view class="address-title">
          {{ item.receiver }} {{ item.phone }}
          <text v-if="item.isDefault" class="tag">默认</text>
        </view>
        <view class="address-detail">{{ item.detail }}</view>
        <view class="actions">
          <text @click="setDefault(item.id)">设为默认</text>
          <text class="danger" @click="remove(item.id)">删除</text>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup>
import { reactive, ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { addAddress, listAddresses, removeAddress, setDefaultAddress } from '../../api/address'

const addresses = ref([])
const form = reactive({
  receiver: '',
  phone: '',
  detail: ''
})

const load = async () => {
  addresses.value = await listAddresses()
}

onShow(load)

const validate = () => {
  if (!/^[\u4e00-\u9fa5]{2,20}$/.test(form.receiver.trim())) return '收件人需填写2到20个汉字'
  if (!/^1[3-9]\d{9}$/.test(form.phone.trim())) return '手机号格式不正确'
  if (form.detail.trim().length < 4) return '请填写详细地址'
  return ''
}

const saveAddress = async () => {
  const message = validate()
  if (message) {
    uni.showToast({ title: message, icon: 'none' })
    return
  }
  await addAddress({
    receiver: form.receiver.trim(),
    phone: form.phone.trim(),
    detail: form.detail.trim()
  })
  form.receiver = ''
  form.phone = ''
  form.detail = ''
  uni.showToast({ title: '地址已保存', icon: 'success' })
  load()
}

const setDefault = async (id) => {
  await setDefaultAddress(id)
  load()
}

const remove = async (id) => {
  await removeAddress(id)
  load()
}
</script>

<style scoped>
.page { min-height: 100vh; padding: 24rpx; background: #f5f6f8; box-sizing: border-box; }
.form-card, .address-card { padding: 28rpx; border-radius: 16rpx; background: #fff; }
.input { height: 84rpx; margin-bottom: 20rpx; padding: 0 24rpx; border-radius: 12rpx; background: #f0f3f7; font-size: 28rpx; box-sizing: border-box; }
.primary-button { height: 88rpx; line-height: 88rpx; border-radius: 12rpx; background: #17a84b; color: #fff; font-size: 30rpx; }
.list { margin-top: 24rpx; display: flex; flex-direction: column; gap: 20rpx; }
.address-title { font-size: 30rpx; font-weight: 700; color: #1f2933; }
.address-detail { margin-top: 14rpx; font-size: 26rpx; color: #667085; }
.tag { margin-left: 12rpx; padding: 4rpx 12rpx; border-radius: 999rpx; background: #eef7f0; color: #17a84b; font-size: 22rpx; }
.actions { margin-top: 20rpx; display: flex; gap: 32rpx; color: #17a84b; font-size: 26rpx; }
.danger { color: #ef4444; }
</style>
