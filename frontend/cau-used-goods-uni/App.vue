<script>
import { listConversations } from './api/chat'
import { tradeService } from './services/trade'

async function updateMessageBadge() {
  try {
    const [systemList, chatResult] = await Promise.all([
      tradeService.getMessages().catch(() => []),
      listConversations({ page: 1, pageSize: 50 }).catch(() => ({ items: [] }))
    ])
    const systemUnread = systemList.filter((item) => !(item.read ?? item.readStatus === 'READ')).length
    const chatUnread = (chatResult.items || []).reduce((sum, item) => sum + Number(item.unreadCount || 0), 0)
    const total = systemUnread + chatUnread
    if (total > 0) {
      uni.setTabBarBadge({ index: 2, text: total > 99 ? '99+' : String(total) })
    } else {
      uni.removeTabBarBadge({ index: 2 })
    }
  } catch (error) {
    // 未登录或后端未启动时不打扰用户正常浏览。
  }
}

export default {
  onLaunch() {
    console.log('App Launch')
  },
  onShow() {
    console.log('App Show')
    updateMessageBadge()
  },
  onHide() {
    console.log('App Hide')
  }
}
</script>

<style lang="scss">
@import './styles/trade.scss';

page {
  background: #f6f8f5;
  color: #24302b;
  font-size: 28rpx;
}

button::after {
  border: 0;
}

.empty-state {
  padding: 96rpx 24rpx;
  color: #89938f;
  text-align: center;
}
</style>
