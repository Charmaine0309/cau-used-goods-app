<template>
  <view class="page">
    <view class="chat-head">
      <view class="chat-title">{{ title || '私信沟通' }}</view>
      <view class="chat-subtitle">请在平台内沟通交易细节，注意保护个人隐私</view>
    </view>

    <scroll-view scroll-y class="messages" :scroll-into-view="lastMessageId">
      <view v-for="item in displayMessages" :id="`msg-${item.id}`" :key="item.id">
        <view v-if="item.showTime" class="time-divider">{{ item.timeText }}</view>
        <view class="message-row" :class="{ mine: item.mine }">
          <image
            v-if="!item.mine && targetAvatar"
            class="avatar image-avatar"
            :src="targetAvatar"
            mode="aspectFill"
            @click.stop="openUser(item.senderId)"
          />
          <view v-else-if="!item.mine" class="avatar seller-avatar" @click.stop="openUser(item.senderId)">{{ targetInitial }}</view>
          <view class="bubble">
            <view class="content">{{ item.content }}</view>
          </view>
          <image
            v-if="item.mine && mineAvatar"
            class="avatar image-avatar"
            :src="mineAvatar"
            mode="aspectFill"
            @click.stop="openUser(currentUserId)"
          />
          <view v-else-if="item.mine" class="avatar buyer-avatar" @click.stop="openUser(currentUserId)">{{ mineInitial }}</view>
        </view>
      </view>
      <view v-if="!messages.length && !loading" class="empty">还没有消息，先打个招呼吧</view>
    </scroll-view>

    <view class="composer">
      <input v-model.trim="draft" class="input" maxlength="500" confirm-type="send" placeholder="输入消息" @confirm="send" />
      <button class="send" :disabled="!draft || sending" @click="send">发送</button>
    </view>
  </view>
</template>

<script setup>
import { computed, ref } from 'vue'
import { onLoad, onPullDownRefresh } from '@dcloudio/uni-app'
import { listMessages, markConversationRead, sendMessage } from '../../api/chat'
import { getPublicProfile } from '../../api/user'
import { getUser } from '../../utils/auth'
import { BASE_URL } from '../../utils/request'
import { navigate } from '../../utils/navigation'

const conversationId = ref('')
const title = ref('')
const targetUserId = ref('')
const messages = ref([])
const draft = ref('')
const loading = ref(false)
const sending = ref(false)
const targetProfile = ref(null)
const mineProfile = ref(null)

const currentUserId = computed(() => getUser()?.id || getUser()?.userId || '')
const currentUser = computed(() => getUser() || {})
const targetAvatar = computed(() => normalizeImage(targetProfile.value?.avatarUrl))
const mineAvatar = computed(() => normalizeImage(currentUser.value.avatarUrl || mineProfile.value?.avatarUrl))
const targetName = computed(() => targetProfile.value?.nickname || title.value || '同学')
const mineName = computed(() => currentUser.value.nickname || mineProfile.value?.nickname || '我')
const targetInitial = computed(() => targetName.value.slice(0, 1) || '同')
const mineInitial = computed(() => mineName.value.slice(0, 1) || '我')
const lastMessageId = computed(() => {
  const last = messages.value[messages.value.length - 1]
  return last ? `msg-${last.id}` : ''
})

function normalizeImage(url) {
  if (!url) return ''
  return /^https?:\/\//.test(url) ? url : `${BASE_URL}${url}`
}

const messageTime = (item) => new Date(String(item.createTime || '').replace(/-/g, '/')).getTime()
const formatTime = (value) => {
  const date = new Date(String(value || '').replace(/-/g, '/'))
  if (Number.isNaN(date.getTime())) return value || ''
  const pad = (n) => String(n).padStart(2, '0')
  return `${pad(date.getMonth() + 1)}-${pad(date.getDate())} ${pad(date.getHours())}:${pad(date.getMinutes())}`
}

const displayMessages = computed(() => {
  let previousTime = 0
  return messages.value.map((item, index) => {
    const currentTime = messageTime(item)
    const showTime = index === 0 || !previousTime || currentTime - previousTime > 5 * 60 * 1000
    if (currentTime) previousTime = currentTime
    return {
      ...item,
      mine: Number(item.senderId) === Number(currentUserId.value),
      showTime,
      timeText: formatTime(item.createTime)
    }
  })
})

const load = async () => {
  if (!conversationId.value) return
  loading.value = true
  try {
    const result = await listMessages(conversationId.value, { page: 1, pageSize: 50 })
    messages.value = result.items || []
    await markConversationRead(conversationId.value)
  } catch (error) {
    uni.showToast({ title: error.message || '消息加载失败', icon: 'none' })
  } finally {
    loading.value = false
  }
}

const loadProfiles = async () => {
  const jobs = []
  if (targetUserId.value) {
    jobs.push(getPublicProfile(targetUserId.value).then((data) => { targetProfile.value = data }).catch(() => {}))
  }
  if (currentUserId.value && !currentUser.value.avatarUrl) {
    jobs.push(getPublicProfile(currentUserId.value).then((data) => { mineProfile.value = data }).catch(() => {}))
  }
  await Promise.all(jobs)
}

const send = async () => {
  if (!draft.value || sending.value) return
  const original = draft.value
  sending.value = true
  try {
    const message = await sendMessage(conversationId.value, original)
    messages.value = messages.value.concat(message)
    draft.value = ''
    if (message.content !== original) {
      uni.showToast({ title: '禁止使用敏感词汇，已替换为 *', icon: 'none' })
    }
  } catch (error) {
    uni.showToast({ title: error.message || '发送失败', icon: 'none' })
  } finally {
    sending.value = false
  }
}

function openUser(id) {
  const profileId = id || targetUserId.value
  if (!profileId) return
  navigate('/pages/user-profile/user-profile', { id: profileId })
}

onLoad(async (options) => {
  conversationId.value = options.conversationId || ''
  title.value = options.title ? decodeURIComponent(options.title) : ''
  targetUserId.value = options.targetUserId || ''
  await Promise.all([load(), loadProfiles()])
})

onPullDownRefresh(async () => {
  try {
    await load()
  } finally {
    uni.stopPullDownRefresh()
  }
})
</script>

<style scoped>
.page { min-height: 100vh; padding-bottom: 120rpx; background: #f5f8f6; box-sizing: border-box; }
.chat-head { padding: 28rpx; background: #fff; }
.chat-title { color: #26342f; font-size: 32rpx; font-weight: 700; }
.chat-subtitle { margin-top: 10rpx; color: #89938f; font-size: 23rpx; }
.messages { height: calc(100vh - 210rpx); padding: 20rpx 28rpx 24rpx; box-sizing: border-box; }
.time-divider { width: fit-content; max-width: 420rpx; margin: 18rpx auto; padding: 6rpx 16rpx; border-radius: 999rpx; background: #dfe7e3; color: #7b8782; font-size: 21rpx; text-align: center; }
.message-row { display: flex; align-items: flex-end; gap: 12rpx; margin-bottom: 18rpx; }
.message-row.mine { justify-content: flex-end; }
.bubble { max-width: 520rpx; padding: 18rpx 22rpx; border-radius: 18rpx; background: #fff; color: #26342f; box-sizing: border-box; }
.avatar { display: flex; width: 58rpx; height: 58rpx; flex-shrink: 0; align-items: center; justify-content: center; border-radius: 50%; color: #fff; font-size: 22rpx; font-weight: 700; }
.seller-avatar { background: #6b8b7e; }
.buyer-avatar { background: #23734f; }
.image-avatar { background: #e8ecef; }
.mine .bubble { background: #23734f; color: #fff; }
.content { font-size: 27rpx; line-height: 1.55; word-break: break-word; }
.empty { margin-top: 180rpx; color: #929c98; text-align: center; }
.composer { position: fixed; right: 0; bottom: 0; left: 0; display: flex; gap: 14rpx; padding: 18rpx 22rpx calc(18rpx + env(safe-area-inset-bottom)); background: #fff; box-sizing: border-box; }
.input { flex: 1; height: 76rpx; padding: 0 24rpx; border-radius: 999rpx; background: #f2f5f3; font-size: 27rpx; box-sizing: border-box; }
.send { width: 132rpx; height: 76rpx; border-radius: 999rpx; background: #23734f; color: #fff; font-size: 27rpx; line-height: 76rpx; }
.send[disabled] { background: #b8c5c0; }
</style>
