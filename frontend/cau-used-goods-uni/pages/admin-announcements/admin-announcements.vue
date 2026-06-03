<template>
  <view class="page">
    <view class="section-title">公告管理</view>

    <view class="form-card">
      <input class="input" v-model="title" placeholder="公告标题" />
      <textarea class="textarea" v-model="content" placeholder="公告内容" />
      <button class="main-button" @click="addAnnouncement">新增公告</button>
    </view>

    <view v-if="announcements.length === 0" class="empty">暂无公告</view>
    <view v-for="item in announcements" :key="item.id" class="card">
      <view class="name">{{ item.title }}</view>
      <view class="desc">{{ statusText(item.status) }}</view>
      <view class="desc">{{ item.content || '暂无内容' }}</view>
      <view class="actions">
        <button size="mini" class="pass" @click="changeStatus(item.id, 'PUBLISHED')">发布</button>
        <button size="mini" class="reject" @click="changeStatus(item.id, 'OFFLINE')">下线</button>
      </view>
    </view>
  </view>
</template>

<script setup>
import { ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { createAnnouncement, getAnnouncements, updateAnnouncementStatus } from '../../api/admin'

const announcements = ref([])
const title = ref('')
const content = ref('')

const load = async () => {
  try {
    const result = await getAnnouncements()
    announcements.value = result?.items || []
  } catch (error) {
    uni.showToast({ title: error.message || '公告加载失败', icon: 'none' })
  }
}

onShow(load)

const addAnnouncement = async () => {
  if (!title.value.trim()) {
    uni.showToast({ title: '请填写公告标题', icon: 'none' })
    return
  }
  try {
    await createAnnouncement({
      title: title.value.trim(),
      content: content.value.trim(),
      status: 'DRAFT'
    })
    title.value = ''
    content.value = ''
    uni.showToast({ title: '已新增', icon: 'success' })
    load()
  } catch (error) {
    uni.showToast({ title: error.message || '新增失败', icon: 'none' })
  }
}

const changeStatus = async (id, status) => {
  try {
    await updateAnnouncementStatus(id, status)
    uni.showToast({ title: status === 'PUBLISHED' ? '已发布' : '已下线', icon: 'success' })
    load()
  } catch (error) {
    uni.showToast({ title: error.message || '操作失败', icon: 'none' })
  }
}

const statusText = (status) => {
  const map = { DRAFT: '草稿', PUBLISHED: '已发布', OFFLINE: '已下线' }
  return map[status] || status || '未知'
}
</script>

<style scoped>
.page { min-height: 100vh; padding: 24rpx; background: #f5f6f8; box-sizing: border-box; }
.section-title { margin: 24rpx 0 18rpx; font-size: 32rpx; font-weight: 700; color: #1f2933; }
.form-card, .card, .empty { padding: 28rpx; border-radius: 16rpx; background: #fff; margin-bottom: 18rpx; }
.input, .textarea { width: 100%; padding: 0 22rpx; border-radius: 12rpx; background: #f1f4f8; font-size: 28rpx; box-sizing: border-box; }
.input { height: 76rpx; margin-bottom: 18rpx; }
.textarea { height: 160rpx; padding-top: 18rpx; margin-bottom: 18rpx; }
.main-button { height: 80rpx; line-height: 80rpx; border-radius: 12rpx; background: #17a84b; color: #fff; font-size: 28rpx; }
.name { font-size: 30rpx; font-weight: 700; color: #1f2933; }
.desc, .empty { margin-top: 10rpx; color: #667085; font-size: 26rpx; }
.actions { margin-top: 18rpx; display: flex; gap: 18rpx; }
.pass { background: #17a84b; color: #fff; }
.reject { background: #fff1f2; color: #ef4444; }
</style>
