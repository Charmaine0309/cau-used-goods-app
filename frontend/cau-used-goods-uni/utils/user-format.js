const statusOf = (value) => String(value || '').toUpperCase()

export const isBannedUserStatus = (status) => {
  const value = statusOf(status)
  return value === 'BANNED' || value === 'PERM_BANNED' || value === 'PERMANENT_BANNED'
}

export const isCanceledUserStatus = (status) => {
  const value = statusOf(status)
  return value === 'CANCELED' || value === 'CANCELLED'
}

export const displayUserName = (user = {}, fallback = '微信用户') => {
  const status = user.accountStatus || user.account_status || user.status || user.userStatus || user.user_status
  if (isBannedUserStatus(status)) return '违规账号'
  if (isCanceledUserStatus(status)) return '已注销用户'
  return user.nickname || user.realName || user.name || fallback
}

export const displayRelatedUserName = (record = {}, role = 'user', fallback = '微信用户') => {
  const prefix = role ? `${role}` : ''
  const nickname = record[`${prefix}Nickname`] || record[`${prefix}_nickname`] || record[`${prefix}Name`] || record[`${prefix}_name`]
  const status = record[`${prefix}AccountStatus`] || record[`${prefix}_account_status`] || record[`${prefix}Status`] || record[`${prefix}_status`]
  return displayUserName({ nickname, status }, fallback)
}
