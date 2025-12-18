<template>
  <div>
    <h2>🛍️ Управление магазином</h2>

    <p v-if="error" class="text-danger alert alert-danger">{{ error }}</p>
    <p v-if="success" class="text-success alert alert-success">{{ success }}</p>

    <button class="btn btn-primary mb-3" @click="toggleForm">
      {{ showForm ? '✕ Отмена' : '+ Добавить товар' }}
    </button>

    <!-- Форма добавления/редактирования -->
    <div v-if="showForm" class="card mb-3 p-3">
      <h5>{{ editingId ? 'Редактировать товар' : 'Новый товар' }}</h5>
      <form @submit.prevent="saveProduct">
        <div class="mb-2">
          <label class="form-label">Название</label>
          <input
            v-model="formData.name"
            type="text"
            class="form-control"
            required
          />
        </div>
        <div class="mb-2">
          <label class="form-label">Описание</label>
          <textarea
            v-model="formData.description"
            class="form-control"
            rows="3"
          ></textarea>
        </div>
        <div class="mb-2">
          <label class="form-label">Цена (₽)</label>
          <input
            v-model.number="formData.price"
            type="number"
            class="form-control"
            min="0"
            required
          />
        </div>
        <div class="mb-2">
          <label class="form-label">URL картинки</label>
          <input
            v-model="formData.image_url"
            type="text"
            class="form-control"
            placeholder="https://..."
          />
        </div>
        <!-- Чекбокс активности -->
        <div class="mb-3 form-check">
          <input
            v-model="formData.is_active"
            type="checkbox"
            class="form-check-input"
            id="activeCheck"
          />
          <label class="form-check-label" for="activeCheck">Товар активен (виден в каталоге)</label>
        </div>
        
        <button type="submit" class="btn btn-success" :disabled="loading">
          {{ loading ? 'Сохранение...' : 'Сохранить' }}
        </button>
      </form>
    </div>

    <!-- Список товаров -->
    <div v-if="loading && products.length === 0" class="text-center">
      <div class="spinner-border" role="status">
        <span class="visually-hidden">Загрузка...</span>
      </div>
    </div>

    <div v-else-if="products.length === 0" class="alert alert-info">
      Товаров нет. Добавьте первый товар!
    </div>

    <table v-else class="table table-hover">
      <thead class="table-light">
        <tr>
          <th>Название</th>
          <th>Цена</th>
          <th>Статус</th>
          <th>Действия</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="product in products" :key="product.id">
          <td>
            <div class="d-flex align-items-center">
              <img 
                v-if="product.image_url" 
                :src="product.image_url" 
                alt="" 
                style="width: 40px; height: 40px; object-fit: cover; margin-right: 10px; border-radius: 4px;"
              >
              <div>
                <strong>{{ product.name }}</strong>
                <br />
                <small class="text-muted">{{ product.description }}</small>
              </div>
            </div>
          </td>
          <td>{{ product.price }} ₽</td>
          <td>
            <span
              v-if="product.is_active"
              class="badge bg-success"
            >
              Активен
            </span>
            <span v-else class="badge bg-secondary">Неактивен</span>
          </td>
          <td>
            <button
              class="btn btn-sm btn-warning"
              @click="editProduct(product)"
            >
              Изменить
            </button>
            <button
              class="btn btn-sm btn-danger ms-1"
              @click="deleteProduct(product.id)"
              :disabled="loading"
            >
              Удалить
            </button>
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import api from '../axios'

const products = ref([])
const error = ref('')
const success = ref('')
const loading = ref(false)
const showForm = ref(false)
const editingId = ref(null)

const initialForm = {
  name: '',
  description: '',
  price: 0,
  image_url: '',
  is_active: true,
}

const formData = ref({ ...initialForm })

async function loadProducts() {
  loading.value = true
  error.value = ''
  try {
    const res = await api.get('/api/admin/products')
    products.value = res.data || []
  } catch (e) {
    error.value = 'Не удалось загрузить товары'
    console.error(e)
  } finally {
    loading.value = false
  }
}

function toggleForm() {
    showForm.value = !showForm.value
    if (!showForm.value) resetForm()
}

function resetForm() {
  formData.value = { ...initialForm }
  editingId.value = null
  showForm.value = false
  error.value = ''
  success.value = ''
}

function editProduct(product) {
  formData.value = { 
      ...product,
      image_url: product.image_url || '' 
  }
  editingId.value = product.id
  showForm.value = true
  window.scrollTo({ top: 0, behavior: 'smooth' })
}

async function saveProduct() {
  error.value = ''
  success.value = ''

  if (!formData.value.name.trim()) {
    error.value = 'Введите название'
    return
  }

  if (formData.value.price < 0) {
    error.value = 'Цена не может быть отрицательной'
    return
  }

  loading.value = true
  try {
    if (editingId.value) {
      await api.put(`/api/admin/products/${editingId.value}`, formData.value)
      success.value = 'Товар обновлен'
    } else {
      await api.post('/api/admin/products', formData.value)
      success.value = 'Товар добавлен'
    }
    await loadProducts()
    resetForm()
  } catch (e) {
    error.value = 'Не удалось сохранить товар: ' + (e.response?.data?.error || e.message)
    console.error(e)
  } finally {
    loading.value = false
  }
}

async function deleteProduct(productId) {
  if (!confirm('Вы уверены? Это действие необратимо.')) return

  error.value = ''
  loading.value = true
  try {
    await api.delete(`/api/admin/products/${productId}`)
    success.value = 'Товар удален'
    await loadProducts()
  } catch (e) {
    error.value = 'Не удалось удалить товар'
    console.error(e)
  } finally {
    loading.value = false
  }
}

onMounted(loadProducts)
</script>
