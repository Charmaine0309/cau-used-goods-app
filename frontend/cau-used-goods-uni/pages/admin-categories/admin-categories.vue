<template>
  <view class="page">
    <view class="page-title">标签管理</view>

    <view class="form-card">
      <view class="form-title">新增商品分类标签</view>
      <input v-model="form.name" class="input" placeholder="请输入分类名称" maxlength="20" />
      <input v-model="form.sortOrder" class="input" placeholder="排序值，可不填" type="number" />
      <button class="submit-button" :loading="submitting" @click="submit">添加标签</button>
    </view>

    <view class="section-title">已有标签</view>
    <view v-if="categories.length === 0" class="empty">暂无分类标签</view>
    <view v-for="item in categories" :key="item.id" class="tag-card">
      <view>
        <view class="tag-name">{{ item.name }}</view>
        <view class="tag-desc">排序 {{ item.sortOrder || 0 }}</view>
      </view>
      <view :class="['tag-status', item.status === 'ENABLED' ? 'enabled' : 'disabled']">
        {{ item.status === 'ENABLED' ? '启用' : '停用' }}
      </view>
    </view>
  </view>
</template>

<script setup>
import { ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { createAdminCategory, getAdminCategories } from '../../api/admin'

const categories = ref([])
const submitting = ref(false)
const form = ref({
  name: '',
  sortOrder: ''
})

const load = async () => {
  try {
    categories.value = await getAdminCategories()
  } catch (error) {
    uni.showToast({ title: error.message || '加载失败', icon: 'none' })
  }
}

onShow(load)

const submit = async () => {
  const name = form.value.name.trim()
  if (!name) {
    uni.showToast({ title: '请输入分类名称', icon: 'none' })
    return
  }
  if (submitting.value) return
  submitting.value = true
  try {
    await createAdminCategory({
      name,
      sortOrder: Number(form.value.sortOrder || 0)
    })
    uni.showToast({ title: '添加成功', icon: 'success' })
    form.value = { name: '', sortOrder: '' }
    load()
  } catch (error) {
    uni.showToast({ title: error.message || '添加失败', icon: 'none' })
  } finally {
    submitting.value = false
  }
}
</script>

<style scoped>
.page {
  min-height: 100vh;
  padding: 28rpx 24rpx;
  background: #f5f6f8;
  box-sizing: border-box;
}

.page-title {
  margin: 18rpx 0 24rpx;
  font-size: 38rpx;
  font-weight: 700;
  color: #1f2933;
}

.form-card,
.tag-card,
.empty {
  padding: 28rpx;
  border-radius: 16rpx;
  background: #fff;
  box-sizing: border-box;
}

.form-title,
.section-title {
  font-size: 32rpx;
  font-weight: 700;
  color: #1f2933;
}

.section-title {
  margin: 32rpx 0 18rpx;
}

.input {
  height: 88rpx;
  margin-top: 22rpx;
  padding: 0 24rpx;
  border-radius: 12rpx;
  background: #f1f5f9;
  font-size: 28rpx;
  box-sizing: border-box;
}

.submit-button {
  margin-top: 26rpx;
  height: 88rpx;
  line-height: 88rpx;
  border-radius: 12rpx;
  background: #17a84b;
  color: #fff;
  font-size: 30rpx;
}

.tag-card {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 18rpx;
}

.tag-name {
  font-size: 30rpx;
  font-weight: 700;
  color: #1f2933;
}

.tag-desc,
.empty {
  margin-top: 8rpx;
  font-size: 24rpx;
  color: #8a96a8;
}

.tag-status {
  padding: 8rpx 16rpx;
  border-radius: 999rpx;
  font-size: 22rpx;
}

.tag-status.enabled {
  background: #dcfce7;
  color: #16a34a;
}

.tag-status.disabled {
  background: #eef2f6;
  color: #667085;
}
</style>
