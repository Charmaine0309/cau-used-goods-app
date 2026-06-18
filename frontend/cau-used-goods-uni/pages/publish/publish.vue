<template>
  <view class="page">
    <view v-if="!canPublish" class="locked-card">
      <view class="locked-title">先完成认证，再发布闲置</view>
      <view class="locked-desc">你的学生认证尚未通过，发布表单已暂时关闭。认证通过后即可发布商品。</view>
      <button class="locked-button" @click="goStudentAuth">去学生认证</button>
    </view>

    <view v-else>
      <view class="hero">
        <text class="eyebrow">CAU CAMPUS MARKET</text>
        <text class="title">{{ editMode ? '编辑商品信息' : '发布你的闲置好物' }}</text>
        <text class="hero-copy">{{ editMode ? '修改商品信息和图片后保存，买家会看到最新内容' : '真实描述物品情况，更容易遇到合适的新主人' }}</text>
      </view>

      <view class="card">
        <view class="field">
          <view class="field-head">
            <text>商品图片</text>
            <text class="muted">{{ form.images.length }}/9</text>
          </view>

          <view class="upload-row">
            <view v-if="form.images.length < 9" class="upload" @click="chooseImages">
              +
              <text>添加图片</text>
            </view>
            <view class="upload-copy">
              <text>选择图片后可拖动、缩放，自主裁剪 4:3 商品图。</text>
              <text>确认裁剪后在下方预览。</text>
            </view>
          </view>

          <view v-if="previewImages.length" class="preview-block">
            <view class="preview-head">
              <text>裁剪预览</text>
              <text class="muted">商品卡片展示效果</text>
            </view>
            <view class="image-grid">
              <view v-for="(image, index) in previewImages" :key="image + index" class="image-item" @click="previewImage(index)">
                <image :src="image" mode="aspectFill" @click.stop="previewImage(index)" @tap.stop="previewImage(index)" />
                <text class="ratio-tag">4:3</text>
                <text class="remove" @click.stop="removeImage(index)">x</text>
              </view>
            </view>
          </view>
        </view>

        <view class="field">
          <view class="field-head">
            <text>商品标题</text>
            <text class="ai" @click="optimizeTitle">AI 优化标题</text>
          </view>
          <input v-model.trim="form.title" maxlength="100" placeholder="例如：九成新小米台灯，宿舍自用" />
        </view>

        <view class="field inline">
          <text>一级分类</text>
          <picker :range="primaryCategories" range-key="name" @change="changePrimaryCategory">
            <view class="picker">{{ primaryName || '请选择一级分类' }} ›</view>
          </picker>
        </view>

        <view v-if="childCategories.length" class="field inline">
          <text>二级分类</text>
          <picker :range="childCategories" range-key="name" @change="changeChildCategory">
            <view class="picker">{{ categoryName || '请选择二级分类' }} ›</view>
          </picker>
        </view>

        <view class="field inline">
          <text>成色</text>
          <picker :range="conditions" @change="changeCondition">
            <view class="picker">{{ form.conditionLevel || '请选择成色' }} ›</view>
          </picker>
        </view>

        <view class="price-row">
          <view class="field">
            <text>原价</text>
            <input v-model="form.originalPrice" type="digit" placeholder="选填" />
          </view>
          <view class="field">
            <text>售价</text>
            <input v-model="form.price" type="digit" placeholder="必填" />
          </view>
        </view>

        <view class="field">
          <view class="field-head">
            <text>商品描述</text>
            <text class="muted">{{ form.description.length }}/{{ DESCRIPTION_LIMIT }}</text>
          </view>
          <textarea v-model.trim="form.description" :maxlength="DESCRIPTION_LIMIT" placeholder="说明使用情况、外观瑕疵、配件等信息" />
          <view class="field-foot">
            <text class="ai" @click="generateDescription">AI 生成描述</text>
          </view>
        </view>

        <view class="field last-field">
          <text>建议面交地点</text>
          <input v-model.trim="form.meetLocation" maxlength="100" placeholder="例如：东区图书馆门口" />
        </view>
      </view>

      <button class="submit" :loading="submitting" @click="submit">{{ editMode ? '保存修改' : '确认发布' }}</button>
    </view>

    <view v-if="cropperVisible" class="crop-mask">
      <view class="crop-panel">
        <view class="crop-head">
          <text>裁剪商品图</text>
          <text class="crop-close" @click="cancelCrop">取消</text>
        </view>
        <view
          class="crop-frame"
          :style="{ width: cropBox.width + 'px', height: cropBox.height + 'px' }"
          @touchstart.stop="startDrag"
          @touchmove.stop.prevent="moveDrag"
          @touchend.stop="endDrag"
        >
          <image
            class="crop-image"
            :src="cropSource"
            mode="aspectFit"
            :style="cropImageStyle"
            draggable="false"
          />
          <view class="crop-border"></view>
        </view>
        <view class="zoom-row">
          <text>缩放</text>
          <slider
            class="zoom-slider"
            :value="cropState.zoomPercent"
            min="100"
            max="300"
            block-size="18"
            activeColor="#23734f"
            backgroundColor="#dbe5df"
            @changing="changeZoom"
            @change="changeZoom"
          />
        </view>
        <view class="crop-actions">
          <button class="crop-secondary" @click="resetCropPosition">居中</button>
          <button class="crop-primary" @click="confirmCrop">确认裁剪</button>
        </view>
      </view>
    </view>

    <canvas
      canvas-id="cropCanvas"
      class="crop-canvas"
      :style="{ width: canvasSize.width + 'px', height: canvasSize.height + 'px' }"
    ></canvas>
  </view>
</template>

<script setup>
import { computed, getCurrentInstance, reactive, ref } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import {
  addProductImages,
  createProduct,
  generateProductDescription,
  listCategories,
  optimizeProductTitle,
  updateProduct,
  uploadProductImage
} from '../../api/product'
import { getCurrentUser } from '../../api/auth'
import { getToken, setUser } from '../../utils/auth'
import { normalizeImage } from '../../utils/product-format'

const MAX_SIZE = 5 * 1024 * 1024
const CANVAS_WIDTH = 800
const CANVAS_HEIGHT = 600
const DESCRIPTION_LIMIT = 50
const conditions = ['全新', '九成新', '八成新', '七成新', '有明显使用痕迹']
const categories = ref([])
const selectedPrimaryId = ref('')
const selectedChildName = ref('')
const submitting = ref(false)
const canPublish = ref(false)
const previewImages = ref([])
const editProductId = ref('')
const originalImages = ref([])
const newImages = ref([])
const cropperVisible = ref(false)
const cropSource = ref('')
const cropResolve = ref(null)
const instance = getCurrentInstance()
const form = reactive({
  images: [],
  title: '',
  categoryId: '',
  conditionLevel: '',
  originalPrice: '',
  price: '',
  description: '',
  meetLocation: ''
})
const canvasSize = reactive({ width: CANVAS_WIDTH, height: CANVAS_HEIGHT })
const cropBox = reactive({ width: 320, height: 240 })
const cropState = reactive({
  imageWidth: 0,
  imageHeight: 0,
  baseScale: 1,
  zoomPercent: 100,
  offsetX: 0,
  offsetY: 0,
  startX: 0,
  startY: 0,
  startOffsetX: 0,
  startOffsetY: 0,
  dragging: false
})

const parentIdOf = (item) => Number(item?.parentId || item?.parent_id || item?.pid || 0)
const fallbackChildrenMap = {
  教材资料: ['公共课教材', '专业课教材', '考研资料', '考证资料', '外语学习', '课程笔记', '课外读物', '打印复印资料'],
  电子产品: ['手机通讯', '电脑笔记本', '平板设备', '耳机音响', '相机摄影', '键盘鼠标', '充电器线材', '数码配件'],
  生活用品: ['宿舍用品', '学习文具', '收纳整理', '台灯照明', '小家电', '厨具餐具', '清洁用品', '床上用品'],
  服饰鞋包: ['男装', '女装', '鞋靴', '箱包', '帽子围巾', '手表饰品', '运动服饰', '其他配饰'],
  运动户外: ['球类用品', '健身器材', '户外装备', '骑行装备', '运动护具', '运动鞋服', '露营旅行'],
  其他: ['票券卡券', '虚拟资料', '乐器器材', '宠物用品', '手工艺品', '模型玩具', '活动周边', '其他闲置']
}
const normalizeCategories = (list = [], parentId = 0) => {
  const source = Array.isArray(list) ? list : (list.items || list.list || list.data || [])
  const result = []
  ;(source || []).forEach((item) => {
    const children = item.children || item.childList || item.subCategories || item.sub_categories || []
    const normalized = {
      ...item,
      id: item.id || item.categoryId || item.category_id,
      name: item.name || item.categoryName || item.category_name,
      parentId: item.parentId || item.parent_id || parentId || 0
    }
    if (Number(normalized.id) !== 0) result.push(normalized)
    if (Array.isArray(children) && children.length) {
      result.push(...normalizeCategories(children, normalized.id))
    }
  })
  return result.filter((item) => item.id && item.name)
}
const primaryCategories = computed(() => categories.value.filter((item) => parentIdOf(item) === 0))
const childCategories = computed(() => {
  if (!selectedPrimaryId.value) return []
  const realChildren = categories.value.filter((item) => parentIdOf(item) === Number(selectedPrimaryId.value) && Number(item.id) !== Number(selectedPrimaryId.value))
  if (realChildren.length) return realChildren
  const primary = primaryCategories.value.find((item) => Number(item.id) === Number(selectedPrimaryId.value))
  return (fallbackChildrenMap[primary?.name] || []).map((name) => ({
    id: selectedPrimaryId.value,
    name,
    fallback: true
  }))
})
const primaryName = computed(() => primaryCategories.value.find((item) => Number(item.id) === Number(selectedPrimaryId.value))?.name || '')
const categoryName = computed(() => selectedChildName.value || categories.value.find((item) => Number(item.id) === Number(form.categoryId))?.name || '')
const toast = (title) => uni.showToast({ title, icon: 'none' })
const zoom = computed(() => cropState.zoomPercent / 100)
const displayWidth = computed(() => cropState.imageWidth * cropState.baseScale * zoom.value)
const displayHeight = computed(() => cropState.imageHeight * cropState.baseScale * zoom.value)
const cropImageStyle = computed(() => ({
  width: `${displayWidth.value}px`,
  height: `${displayHeight.value}px`,
  transform: `translate(${cropState.offsetX}px, ${cropState.offsetY}px)`
}))
const editMode = computed(() => !!editProductId.value)

const resetForm = () => {
  editProductId.value = ''
  originalImages.value = []
  newImages.value = []
  form.images = []
  previewImages.value = []
  form.title = ''
  selectedPrimaryId.value = ''
  selectedChildName.value = ''
  form.categoryId = ''
  form.conditionLevel = ''
  form.originalPrice = ''
  form.price = ''
  form.description = ''
  form.meetLocation = ''
}

const fillEditForm = (product = {}) => {
  editProductId.value = String(product.id || '')
  form.images = Array.isArray(product.images) ? product.images.slice() : []
  originalImages.value = form.images.slice()
  newImages.value = []
  previewImages.value = form.images.map((image) => normalizeImage(image)).filter(Boolean)
  form.title = product.title || ''
  form.categoryId = product.categoryId || product.category_id || ''
  form.conditionLevel = product.conditionLevel || product.condition_level || ''
  form.originalPrice = product.originalPrice ?? product.original_price ?? ''
  form.price = product.price ?? ''
  form.description = product.description || ''
  form.meetLocation = product.meetLocation || product.meet_location || ''

  const category = categories.value.find((item) => Number(item.id) === Number(form.categoryId))
  if (category && parentIdOf(category)) {
    selectedPrimaryId.value = category.parentId
    selectedChildName.value = category.name
  } else if (category) {
    selectedPrimaryId.value = category.id
    selectedChildName.value = ''
  }
}

const applyEditState = () => {
  const id = uni.getStorageSync('PUBLISH_EDIT_PRODUCT_ID')
  const product = uni.getStorageSync('PUBLISH_EDIT_PRODUCT_DATA')
  if (!id || !product?.id) {
    if (editProductId.value) resetForm()
    return
  }
  if (String(id) === String(editProductId.value)) return
  fillEditForm(product)
}

const isVerified = (user) => {
  const status = user?.authStatus || user?.auth_status || ''
  return status === 'VERIFIED'
}

const loadPublishState = async () => {
  if (!getToken()) {
    canPublish.value = false
    uni.showToast({ title: '请先登录', icon: 'none' })
    return
  }

  try {
    const current = await getCurrentUser()
    setUser(current)
    canPublish.value = isVerified(current)
    if (!canPublish.value) {
      uni.showToast({ title: '未完成学生认证', icon: 'none' })
      return
    }
    categories.value = normalizeCategories(await listCategories())
    applyEditState()
  } catch (error) {
    canPublish.value = false
    toast(error.message || '加载失败')
  }
}

onShow(loadPublishState)

const goStudentAuth = () => uni.navigateTo({ url: '/pages/student-auth/student-auth' })
const changePrimaryCategory = ({ detail }) => {
  const primary = primaryCategories.value[detail.value]
  selectedPrimaryId.value = primary?.id || ''
  selectedChildName.value = ''
  form.categoryId = childCategories.value.length ? '' : selectedPrimaryId.value
}
const changeChildCategory = ({ detail }) => {
  const child = childCategories.value[detail.value]
  form.categoryId = child?.id || ''
  selectedChildName.value = child?.name || ''
}
const changeCondition = ({ detail }) => { form.conditionLevel = conditions[detail.value] }
const removeImage = (index) => {
  const image = form.images[index]
  if (editMode.value && originalImages.value.includes(image)) {
    toast('当前后端暂未提供删除已上传图片接口')
    return
  }
  if (newImages.value.includes(image)) {
    newImages.value = newImages.value.filter((item) => item !== image)
  }
  form.images.splice(index, 1)
  previewImages.value.splice(index, 1)
}
const previewImage = (index) => {
  if (!previewImages.value.length) return
  const remoteUrls = form.images.map((image) => normalizeImage(image)).filter(Boolean)
  const localUrls = previewImages.value.filter(Boolean)
  const urls = remoteUrls.length === previewImages.value.length ? remoteUrls : localUrls
  const current = remoteUrls[index] || localUrls[index]
  if (!current) return
  uni.previewImage({
    urls,
    current
  })
}

const clampOffset = () => {
  const width = displayWidth.value
  const height = displayHeight.value
  cropState.offsetX = width <= cropBox.width
    ? (cropBox.width - width) / 2
    : Math.min(0, Math.max(cropBox.width - width, cropState.offsetX))
  cropState.offsetY = height <= cropBox.height
    ? (cropBox.height - height) / 2
    : Math.min(0, Math.max(cropBox.height - height, cropState.offsetY))
}

const resetCropPosition = () => {
  cropState.zoomPercent = 100
  cropState.offsetX = (cropBox.width - displayWidth.value) / 2
  cropState.offsetY = (cropBox.height - displayHeight.value) / 2
  clampOffset()
}

const openCropper = (src) => {
  return new Promise((resolve) => {
    uni.getImageInfo({
      src,
      success: (info) => {
        const system = uni.getSystemInfoSync()
        cropBox.width = Math.min(system.windowWidth - 56, 360)
        cropBox.height = Math.round(cropBox.width * 3 / 4)
        cropState.imageWidth = info.width
        cropState.imageHeight = info.height
        cropState.baseScale = Math.max(cropBox.width / info.width, cropBox.height / info.height)
        cropState.zoomPercent = 100
        cropSource.value = src
        cropResolve.value = resolve
        cropperVisible.value = true
        setTimeout(resetCropPosition, 20)
      },
      fail: () => {
        toast('图片读取失败')
        resolve('')
      }
    })
  })
}

const startDrag = (event) => {
  const touch = event.touches?.[0]
  if (!touch) return
  cropState.dragging = true
  cropState.startX = touch.clientX
  cropState.startY = touch.clientY
  cropState.startOffsetX = cropState.offsetX
  cropState.startOffsetY = cropState.offsetY
}

const moveDrag = (event) => {
  if (!cropState.dragging) return
  const touch = event.touches?.[0]
  if (!touch) return
  cropState.offsetX = cropState.startOffsetX + touch.clientX - cropState.startX
  cropState.offsetY = cropState.startOffsetY + touch.clientY - cropState.startY
  clampOffset()
}

const endDrag = () => {
  cropState.dragging = false
}

const changeZoom = ({ detail }) => {
  const oldWidth = displayWidth.value
  const oldHeight = displayHeight.value
  const centerX = cropBox.width / 2
  const centerY = cropBox.height / 2
  const ratioX = oldWidth ? (centerX - cropState.offsetX) / oldWidth : 0.5
  const ratioY = oldHeight ? (centerY - cropState.offsetY) / oldHeight : 0.5
  cropState.zoomPercent = detail.value
  cropState.offsetX = centerX - displayWidth.value * ratioX
  cropState.offsetY = centerY - displayHeight.value * ratioY
  clampOffset()
}

const finishCrop = (path) => {
  const resolve = cropResolve.value
  cropperVisible.value = false
  cropResolve.value = null
  cropSource.value = ''
  if (resolve) resolve(path)
}

const cancelCrop = () => {
  finishCrop('')
}

const confirmCrop = () => {
  const ctx = uni.createCanvasContext('cropCanvas', instance?.proxy)
  const scaleX = CANVAS_WIDTH / cropBox.width
  const scaleY = CANVAS_HEIGHT / cropBox.height
  ctx.clearRect(0, 0, CANVAS_WIDTH, CANVAS_HEIGHT)
  ctx.drawImage(
    cropSource.value,
    cropState.offsetX * scaleX,
    cropState.offsetY * scaleY,
    displayWidth.value * scaleX,
    displayHeight.value * scaleY
  )
  ctx.draw(false, () => {
    uni.canvasToTempFilePath({
      canvasId: 'cropCanvas',
      x: 0,
      y: 0,
      width: CANVAS_WIDTH,
      height: CANVAS_HEIGHT,
      destWidth: CANVAS_WIDTH,
      destHeight: CANVAS_HEIGHT,
      fileType: 'jpg',
      quality: 0.9,
      success: (res) => finishCrop(res.tempFilePath),
      fail: () => {
        toast('裁剪失败，请重新选择图片')
        finishCrop('')
      }
    }, instance?.proxy)
  })
}

const chooseImages = async () => {
  if (!canPublish.value) return toast('未完成学生认证')

  try {
    const result = await uni.chooseImage({
      count: 9 - form.images.length,
      sizeType: ['compressed']
    })
    const files = result.tempFiles.filter((file) => {
      if (file.size > MAX_SIZE) {
        toast('单张图片不能超过 5 MB')
        return false
      }
      return true
    })
    if (!files.length) return

    for (const file of files) {
      const croppedPath = await openCropper(file.path)
      if (!croppedPath) continue
      uni.showLoading({ title: '上传裁剪图中' })
      const uploaded = await uploadProductImage(croppedPath)
      form.images.push(uploaded.imageUrl)
      if (editMode.value) newImages.value.push(uploaded.imageUrl)
      previewImages.value.push(croppedPath)
      uni.hideLoading()
    }
  } catch (error) {
    if (!error?.errMsg?.includes('cancel')) toast(error.message || '图片选择或上传失败')
  } finally {
    uni.hideLoading()
  }
}

const optimizeTitle = async () => {
  if (!canPublish.value) return toast('未完成学生认证')
  if (!form.title) return toast('请先填写一个基础标题')
  try {
    uni.showLoading({ title: 'AI 正在优化' })
    const { titles } = await optimizeProductTitle({
      title: form.title,
      categoryName: categoryName.value,
      conditionLevel: form.conditionLevel
    })
    uni.showActionSheet({
      itemList: titles,
      success: ({ tapIndex }) => { form.title = titles[tapIndex] }
    })
  } catch (error) {
    toast('AI 服务暂时不可用')
  } finally {
    uni.hideLoading()
  }
}

const generateDescription = async () => {
  if (!canPublish.value) return toast('未完成学生认证')
  if (!form.title) return toast('请先填写标题')
  try {
    uni.showLoading({ title: 'AI 正在生成' })
    const result = await generateProductDescription({
      title: form.title,
      categoryName: categoryName.value,
      conditionLevel: form.conditionLevel,
      meetLocation: form.meetLocation
    })
    form.description = String(result.description || '').slice(0, DESCRIPTION_LIMIT)
  } catch (error) {
    toast('AI 服务暂时不可用')
  } finally {
    uni.hideLoading()
  }
}

const validate = () => {
  if (!canPublish.value) return '未完成学生认证'
  if (!form.images.length) return '请至少上传一张商品图片'
  if (!form.title) return '请填写商品标题'
  if (!selectedPrimaryId.value) return '请选择一级分类'
  if (childCategories.value.length && !form.categoryId) return '请选择二级分类'
  if (!form.categoryId) form.categoryId = selectedPrimaryId.value
  if (!form.conditionLevel) return '请选择商品成色'
  if (!form.price || Number(form.price) <= 0) return '请填写正确的商品售价'
  if (!form.description) return '请填写商品描述'
  if (form.description.length > DESCRIPTION_LIMIT) return `商品描述不能超过 ${DESCRIPTION_LIMIT} 字`
  return ''
}

const submit = async () => {
  const message = validate()
  if (message) return toast(message)
  submitting.value = true
  try {
    const payload = {
      ...form,
      price: Number(form.price),
      originalPrice: Number(form.originalPrice || 0)
    }
    if (editMode.value) {
      const { images, ...productPayload } = payload
      await updateProduct(editProductId.value, productPayload)
      if (newImages.value.length) await addProductImages(editProductId.value, newImages.value)
      uni.removeStorageSync('PUBLISH_EDIT_PRODUCT_ID')
      uni.removeStorageSync('PUBLISH_EDIT_PRODUCT_DATA')
      resetForm()
      uni.setStorageSync('PRODUCT_LIST_DIRTY', true)
      uni.showToast({ title: '修改成功', icon: 'success' })
      setTimeout(() => uni.navigateTo({ url: '/pages/my-products/my-products' }), 600)
      return
    }
    await createProduct(payload)
    resetForm()
    uni.setStorageSync('PRODUCT_LIST_DIRTY', true)
    uni.showToast({ title: '发布成功', icon: 'success' })
    setTimeout(() => uni.switchTab({ url: '/pages/home/home' }), 600)
  } catch (error) {
    toast(error.message || '发布失败')
  } finally {
    submitting.value = false
  }
}
</script>

<style scoped>
.page { min-height: 100vh; padding-bottom: 36rpx; background: #f5f8f6; box-sizing: border-box; }
.hero { margin: 24rpx 28rpx 0; padding: 34rpx 30rpx; border-radius: 24rpx; background: #2d855f; color: #fff; }
.eyebrow, .title, .hero-copy, .muted { display: block; }
.eyebrow { color: rgba(255,255,255,.72); font-size: 20rpx; letter-spacing: 3rpx; }
.title { margin-top: 12rpx; font-size: 38rpx; font-weight: 700; }
.hero-copy { margin-top: 10rpx; color: rgba(255,255,255,.78); font-size: 24rpx; }
.locked-card, .card { margin: 28rpx; padding: 28rpx; border-radius: 18rpx; background: #fff; }
.locked-title { color: #26342f; font-size: 34rpx; font-weight: 700; }
.locked-desc { margin-top: 16rpx; color: #7d8984; font-size: 26rpx; line-height: 1.6; }
.locked-button, .submit { margin-top: 30rpx; border-radius: 999rpx; background: #23734f; color: #fff; }
.field { margin-bottom: 26rpx; }
.last-field { margin-bottom: 0; }
.field-head, .inline, .price-row, .preview-head, .field-foot { display: flex; align-items: center; justify-content: space-between; gap: 18rpx; }
.field text, .field-head { color: #26342f; font-size: 27rpx; font-weight: 600; }
.muted { color: #89938f; font-size: 22rpx; font-weight: 400; }
.ai { color: #23734f !important; font-size: 24rpx !important; font-weight: 600 !important; }
input, textarea, .picker { width: 100%; margin-top: 14rpx; padding: 18rpx 22rpx; border-radius: 16rpx; background: #f3f6f4; color: #26342f; font-size: 27rpx; box-sizing: border-box; }
input, .picker { min-height: 72rpx; }
textarea { height: 180rpx; }
.field-foot { margin-top: 10rpx; justify-content: flex-end; }
.inline .picker { min-width: 360rpx; margin-top: 0; text-align: right; }
.price-row .field { flex: 1; margin-bottom: 0; }
.upload-row { display: flex; align-items: center; gap: 18rpx; margin-top: 16rpx; }
.upload { display: flex; width: 176rpx; height: 132rpx; flex-shrink: 0; flex-direction: column; align-items: center; justify-content: center; border-radius: 14rpx; background: #edf4f1; color: #23734f; font-size: 42rpx; }
.upload text { margin-top: 4rpx; color: #23734f; font-size: 22rpx; }
.upload-copy { display: flex; min-width: 0; flex: 1; flex-direction: column; gap: 8rpx; }
.upload-copy text { color: #7d8984; font-size: 23rpx; font-weight: 400; line-height: 1.35; }
.preview-block { margin-top: 20rpx; }
.preview-head { margin-bottom: 12rpx; }
.image-grid { display: grid; grid-template-columns: repeat(3, 1fr); gap: 16rpx; }
.image-item { position: relative; aspect-ratio: 4 / 3; overflow: hidden; border-radius: 14rpx; background: #edf4f1; }
.image-item image { width: 100%; height: 100%; }
.ratio-tag, .remove { position: absolute; top: 8rpx; border-radius: 999rpx; color: #fff !important; font-size: 20rpx !important; font-weight: 700 !important; line-height: 34rpx; text-align: center; }
.ratio-tag { left: 8rpx; padding: 0 10rpx; background: rgba(35,115,79,.82); }
.remove { right: 8rpx; width: 34rpx; height: 34rpx; background: rgba(0,0,0,.55); }
.submit { margin: 0 28rpx; height: 88rpx; line-height: 88rpx; font-size: 30rpx; }
.crop-mask { position: fixed; z-index: 99; top: 0; right: 0; bottom: 0; left: 0; display: flex; align-items: center; justify-content: center; padding: 28rpx; background: rgba(0,0,0,.58); box-sizing: border-box; }
.crop-panel { width: 100%; padding: 28rpx; border-radius: 20rpx; background: #fff; box-sizing: border-box; }
.crop-head, .zoom-row, .crop-actions { display: flex; align-items: center; justify-content: space-between; gap: 18rpx; }
.crop-head { margin-bottom: 20rpx; color: #26342f; font-size: 30rpx; font-weight: 700; }
.crop-close { color: #89938f; font-size: 25rpx; font-weight: 500; }
.crop-frame { position: relative; overflow: hidden; margin: 0 auto; background: #111; touch-action: none; }
.crop-image { position: absolute; top: 0; left: 0; will-change: transform; }
.crop-border { position: absolute; top: 0; right: 0; bottom: 0; left: 0; border: 4rpx solid rgba(255,255,255,.96); box-shadow: inset 0 0 0 2rpx rgba(35,115,79,.7); pointer-events: none; }
.zoom-row { margin-top: 22rpx; color: #26342f; font-size: 25rpx; }
.zoom-slider { flex: 1; }
.crop-actions { margin-top: 18rpx; }
.crop-secondary, .crop-primary { flex: 1; height: 78rpx; border-radius: 999rpx; font-size: 27rpx; line-height: 78rpx; }
.crop-secondary { background: #edf4f1; color: #23734f; }
.crop-primary { background: #23734f; color: #fff; }
.crop-canvas { position: fixed; left: -9999px; top: -9999px; opacity: 0; pointer-events: none; }
</style>
