<template>
  <div>
    <h2>📋 Управление отзывами</h2>

    <p v-if="error" class="alert alert-danger">{{ error }}</p>
    <p v-if="success" class="alert alert-success">{{ success }}</p>

    <div class="mb-3">
      <button
        :class="['btn', activeFilter === 'pending' ? 'btn-primary' : 'btn-outline-primary']"
        @click="setFilter('pending')"
      >
        На модерации ({{ counts.pending }})
      </button>
      <button
        :class="['btn ms-2', activeFilter === 'approved' ? 'btn-success' : 'btn-outline-success']"
        @click="setFilter('approved')"
      >
        Одобрено ({{ counts.approved }})
      </button>
      <button
        :class="['btn ms-2', activeFilter === 'rejected' ? 'btn-danger' : 'btn-outline-danger']"
        @click="setFilter('rejected')"
      >
        Отклонено ({{ counts.rejected }})
      </button>
    </div>

    <div v-if="loading" class="text-center">
      <div class="spinner-border" role="status">
        <span class="visually-hidden">Загрузка...</span>
      </div>
    </div>

    <div v-else-if="reviews.length === 0" class="alert alert-info">
      Отзывов не найдено
    </div>

    <div v-else class="row">
      <div
        v-for="review in reviews"
        :key="review.id"
        class="col-md-6 mb-3"
      >
        <div class="card h-100">
          <div class="card-header bg-light">
            <div class="d-flex justify-content-between align-items-center">
              <div>
                <strong>{{ review.username }}</strong><br />
                <small class="text-muted">
                  {{ formatDate(review.created_at) }}
                </small>
              </div>
              <span class="badge bg-warning text-dark">
                {{ review.rating }} ⭐
              </span>
            </div>
          </div>
            <div class="card-body">
              <p>{{ review.comment }}</p>
              <p class="mb-0">
                <span
                  v-if="review.status === 'approved'"
                  class="text-success"
                >
                  ✓ Одобрено
                </span>
                <span
                  v-else-if="review.status === 'rejected'"
                  class="text-danger"
                >
                  ✗ Отклонено
                </span>
                <span
                  v-else
                  class="text-warning"
                >
                  ⏳ Ожидает модерации
                </span>
              </p>
            </div>
              <div class="card-footer bg-light d-flex justify-content-between align-items-center">
              <!-- Кнопки для PENDING -->
              <div v-if="review.status === 'pending'">
                <button
                  class="btn btn-sm btn-success"
                  @click="approveReview(review.id)"
                  :disabled="processing"
                >
                  ✓
                </button>
                <button
                  class="btn btn-sm btn-danger ms-2"
                  @click="rejectReview(review.id)"
                  :disabled="processing"
                >
                  ✗
                </button>
                <button
                  class="btn btn-sm btn-outline-secondary ms-2"
                  @click="deleteReview(review.id)"
                  :disabled="processing"
                  title="Удалить спам"
                >
                  🗑
                </button>
              </div>

              <!-- Кнопки для APPROVED / REJECTED -->
              <div v-else class="ms-auto">
                <button
                  class="btn btn-sm btn-outline-danger"
                  @click="deleteReview(review.id)"
                  :disabled="processing"
                >
                  🗑 Удалить
                </button>
              </div>
            </div>
            </div>
          </div>
        </div>
      </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import api from '../axios'

const reviews = ref([])
const error = ref('')
const success = ref('')
const loading = ref(false)
const processing = ref(false)
const activeFilter = ref('pending')
const counts = ref({
  pending: 0,
  approved: 0,
  rejected: 0,
})

function formatDate(dateStr) {
  return new Date(dateStr).toLocaleString('ru-RU', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}

async function loadReviews() {
  loading.value = true
  error.value = ''
  try {
    const res = await api.get(`/api/admin/reviews?status=${activeFilter.value}`)
    reviews.value = res.data || []
  } catch (e) {
    error.value = 'Не удалось загрузить отзывы'
    console.error(e)
  } finally {
    loading.value = false
  }
}

async function loadCounts() {
  try {
    const [p, a, r] = await Promise.all([
      api.get('/api/admin/reviews?status=pending'),
      api.get('/api/admin/reviews?status=approved'),
      api.get('/api/admin/reviews?status=rejected'),
    ])
    counts.value.pending = (p.data || []).length
    counts.value.approved = (a.data || []).length
    counts.value.rejected = (r.data || []).length
  } catch (e) {
    console.error('Failed to load review counts', e)
  }
}

async function approveReview(id) {
  processing.value = true
  error.value = ''
  success.value = ''
  try {
    await api.post(`/api/admin/reviews/${id}/approve`)
    success.value = 'Отзыв одобрен'
    await Promise.all([loadReviews(), loadCounts()])
  } catch (e) {
    error.value = 'Не удалось одобрить отзыв'
    console.error(e)
  } finally {
    processing.value = false
  }
}

async function rejectReview(id) {
  processing.value = true
  error.value = ''
  success.value = ''
  try {
    await api.post(`/api/admin/reviews/${id}/reject`)
    success.value = 'Отзыв отклонён'
    await Promise.all([loadReviews(), loadCounts()])
  } catch (e) {
    error.value = 'Не удалось отклонить отзыв'
    console.error(e)
  } finally {
    processing.value = false
  }
}

async function deleteReview(id) {
  if (!confirm('Удалить отзыв навсегда?')) return
  processing.value = true
  error.value = ''
  success.value = ''
  try {
    await api.delete(`/api/admin/reviews/${id}`)
    success.value = 'Отзыв удалён'
    await Promise.all([loadReviews(), loadCounts()])
  } catch (e) {
    error.value = 'Не удалось удалить отзыв'
    console.error(e)
  } finally {
    processing.value = false
  }
}

function setFilter(status) {
  activeFilter.value = status
  success.value = ''
  loadReviews()
}

onMounted(async () => {
  await loadReviews()
  await loadCounts()
})
</script>
