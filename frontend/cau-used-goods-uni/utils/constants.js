export const ORDER_STATUS = {
  PENDING_CONFIRM: { label: '待卖家确认', tone: 'warning' },
  WAIT_MEET: { label: '待线下面交', tone: 'primary' },
  COMPLETED: { label: '已完成', tone: 'success' },
  CANCELED: { label: '已取消', tone: 'muted' },
  EXCEPTION_CLOSED: { label: '异常关闭', tone: 'danger' }
}

export const REPORT_STATUS = {
  PENDING: { label: '待处理', tone: 'warning' },
  RESOLVED: { label: '已处理', tone: 'success' },
  REJECTED: { label: '未采纳', tone: 'muted' }
}

export const MESSAGE_TYPE = {
  ORDER: '订单通知',
  REPORT: '举报反馈',
  SYSTEM: '系统消息'
}
