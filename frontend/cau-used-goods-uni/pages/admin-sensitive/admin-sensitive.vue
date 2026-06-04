<template>
  <view class="page">
    <view class="section-title">敏感词管理</view>

    <view class="form-card">
      <input class="input" v-model="word" placeholder="输入敏感词" />
      <picker :range="wordTypeOptions" range-key="label" :value="wordTypeIndex" @change="changeWordType">
        <view class="input picker-input">{{ wordTypeOptions[wordTypeIndex].label }}</view>
      </picker>
      <button class="main-button" @click="addWord">新增敏感词</button>
    </view>

    <view v-if="words.length === 0" class="empty">暂无敏感词</view>
    <view v-for="item in words" :key="item.id" class="card">
      <view class="name">{{ item.word }}</view>
      <view class="desc">{{ item.wordType }} · {{ statusText(item.status) }}</view>
      <view class="actions">
        <button size="mini" class="pass" @click="enableWord(item)">启用</button>
        <button size="mini" class="reject" @click="removeWord(item.id)">禁用</button>
      </view>
    </view>
  </view>
</template>

<script setup>
import { ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { createSensitiveWord, deleteSensitiveWord, getSensitiveWords, updateSensitiveWord } from '../../api/admin'

const words = ref([])
const word = ref('')
const wordTypeOptions = [
  { label: '违规禁用词', value: 'FORBIDDEN' },
  { label: '风险提示词', value: 'RISK' }
]
const wordTypeIndex = ref(0)

const load = async () => {
  try {
    const result = await getSensitiveWords()
    words.value = result?.items || []
  } catch (error) {
    uni.showToast({ title: error.message || '敏感词加载失败', icon: 'none' })
  }
}

onShow(load)

const addWord = async () => {
  if (!word.value.trim()) {
    uni.showToast({ title: '请填写敏感词', icon: 'none' })
    return
  }
  try {
    await createSensitiveWord({
      word: word.value.trim(),
      wordType: wordTypeOptions[wordTypeIndex.value].value,
      status: 'ENABLED'
    })
    word.value = ''
    uni.showToast({ title: '已新增', icon: 'success' })
    load()
  } catch (error) {
    uni.showToast({ title: error.message || '新增失败', icon: 'none' })
  }
}

const enableWord = async (item) => {
  try {
    await updateSensitiveWord(item.id, {
      word: item.word,
      wordType: item.wordType,
      status: 'ENABLED'
    })
    uni.showToast({ title: '已启用', icon: 'success' })
    load()
  } catch (error) {
    uni.showToast({ title: error.message || '启用失败', icon: 'none' })
  }
}

const removeWord = async (id) => {
  try {
    await deleteSensitiveWord(id)
    uni.showToast({ title: '已禁用', icon: 'success' })
    load()
  } catch (error) {
    uni.showToast({ title: error.message || '禁用失败', icon: 'none' })
  }
}

const statusText = (status) => status === 'ENABLED' ? '启用中' : '已禁用'

const changeWordType = (event) => {
  wordTypeIndex.value = Number(event.detail.value || 0)
}
</script>

<style scoped>
.page { min-height: 100vh; padding: 24rpx; background: #f5f6f8; box-sizing: border-box; }
.section-title { margin: 24rpx 0 18rpx; font-size: 32rpx; font-weight: 700; color: #1f2933; }
.form-card, .card, .empty { padding: 28rpx; border-radius: 16rpx; background: #fff; margin-bottom: 18rpx; }
.input { height: 76rpx; margin-bottom: 18rpx; padding: 0 22rpx; border-radius: 12rpx; background: #f1f4f8; font-size: 28rpx; }
.picker-input { line-height: 76rpx; color: #1f2933; box-sizing: border-box; }
.main-button { height: 80rpx; line-height: 80rpx; border-radius: 12rpx; background: #17a84b; color: #fff; font-size: 28rpx; }
.name { font-size: 30rpx; font-weight: 700; color: #1f2933; }
.desc, .empty { margin-top: 10rpx; color: #667085; font-size: 26rpx; }
.actions { margin-top: 18rpx; display: flex; gap: 18rpx; }
.pass { background: #17a84b; color: #fff; }
.reject { background: #fff1f2; color: #ef4444; }
</style>
