<template>
  <div>
    <h2>Оставить отзыв</h2>

    <p v-if="!loggedIn" class="alert alert-info">
      <router-link to="/login">Войдите</router-link>, чтобы оставить отзыв
    </p>

    <div v-else>
      <p v-if="error" class="text-danger alert alert-danger">{{ error }}</p>
      <p v-if="success" class="text-success alert alert-success">{{ success }}</p>

      <form @submit.prevent="submitReview" class="card p-4">
        <div class="mb-3">
          <label class="form-label">Ваша оценка</label>
          <div class="d-flex gap-2">
            <button
              v-for="n in 5"
              :key="n"
              type="button"
              class="btn"
              :class="rating === n ? 'btn-warning' : 'btn-outline-warning'"
              @click="rating = n"
            >
              {{ '⭐'.repeat(n) }}
            </button>
          </div>
          <small v-if="!rating" class="text-danger">Выберите оценку</small>
        </div>

        <div class="mb-3">
          <label for="comment" class="form-label">Ваш комментарий</label>
          <textarea
            id="comment"
            v-model="comment"
            class="form-control"
            rows="5"
            placeholder="Поделитесь своим мнением..."
            required
          ></textarea>
          <small class="text-muted">{{ comment.length }} / 500 символов</small>
        </div>

        <button type="submit" class="btn btn-primary" :disabled="loading">
          <span v-if="loading" class="spinner-border spinner-border-sm me-2"></span>
          {{ loading ? 'Отправка...' : 'Отправить отзыв' }}
        </button>
      </form>

      <hr />

      <div class="mt-4">
        <h5>Последние одобренные отзывы</h5>
        <div v-if="recentReviews.length === 0" class="text-muted">
          Отзывов пока нет
        </div>
        <div v-else class="row">
          <div v-for="rev in recentReviews" :key="rev.id" class="col-md-6 mb-2">
            <div class="card">
              <div class="card-body p-3">
                <div class="d-flex justify-content-between">
                  <strong>{{ rev.username }}</strong>
                  <span class="badge bg-warning">{{ rev.rating }}⭐</span>
                </div>
                <p class="mb-1">{{ rev.comment }}</p>
                <small class="text-muted">
                  {{ formatDate(rev.created_at) }}
                </small>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import api from '../axios'

const router = useRouter()
const loggedIn = computed(() => !!localStorage.getItem('token'))

const rating = ref(0)
const comment = ref('')
const error = ref('')
const success = ref('')
const loading = ref(false)
const recentReviews = ref([])

function formatDate(dateStr) {
  return new Date(dateStr).toLocaleDateString('ru-RU')
}

async function submitReview() {
  error.value = ''
  success.value = ''

  if (!rating.value) {
    error.value = 'Выберите оценку'
    return
  }

  if (!comment.value.trim()) {
    error.value = 'Напишите комментарий'
    return
  }

  if (comment.value.length > 500) {
    error.value = 'Комментарий не должен превышать 500 символов'
    return
  }

  loading.value = true
  try {
    await api.post('/api/reviews', {
      rating: rating.value,
      comment: comment.value,
    })
    success.value = 'Отзыв отправлен на модерацию! Спасибо!'
    rating.value = 0
    comment.value = ''
    await loadRecentReviews()
  } catch (e) {
    error.value = 'Не удалось отправить отзыв'
    console.error(e)
  } finally {
    loading.value = false
  }
}

async function loadRecentReviews() {
  try {
    const res = await api.get('/api/reviews')
    recentReviews.value = (res.data || []).slice(0, 4)
  } catch (e) {
    console.error('Failed to load recent reviews', e)
  }
}

onMounted(loadRecentReviews)
</script>
