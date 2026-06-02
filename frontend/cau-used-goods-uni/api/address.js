const ADDRESS_KEY = 'CAU_USED_GOODS_ADDRESS_LIST'

const readList = () => {
  return uni.getStorageSync(ADDRESS_KEY) || []
}

const writeList = (list) => {
  uni.setStorageSync(ADDRESS_KEY, list)
  return list
}

export const listAddresses = () => {
  return Promise.resolve(readList())
}

export const addAddress = (address) => {
  const list = readList()
  const next = {
    id: Date.now(),
    isDefault: list.length === 0,
    ...address
  }
  return Promise.resolve(writeList([next, ...list]))
}

export const removeAddress = (id) => {
  const list = readList().filter((item) => item.id !== id)
  return Promise.resolve(writeList(list))
}

export const setDefaultAddress = (id) => {
  const list = readList().map((item) => ({
    ...item,
    isDefault: item.id === id
  }))
  return Promise.resolve(writeList(list))
}
