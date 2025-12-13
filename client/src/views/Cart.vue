<template>
  <main class="container mt-4 main-content">
  <div>
    <h2>Корзина</h2>

    <p v-if="error" class="text-danger">{{ error }}</p>

    <div v-if="items.length === 0" class="alert alert-info">
      Ваша корзина пуста.
    </div>

    <div v-else>
      <table class="table">
        <thead>
          <tr>
            <th>Товар</th>
            <th>Цена</th>
            <th>Картинка</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="item in items" :key="item.id">
            <td>{{ item.name }}</td>
            <td>{{ item.price }} ₽</td>
            <td>
              <img
                :src="item.image"
                alt=""
                style="height: 50px"
              />
            </td>
          </tr>
        </tbody>
      </table>

      <p class="fw-bold">
        Итого: {{ total }} ₽
      </p>

      <button class="btn btn-danger" @click="clearCart">
        Очистить корзину
      </button>
    </div>
  </div>
    <RouterView />
</main>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import api from '../axios'

const items = ref([])
const error = ref('')

const total = computed(() =>
  items.value.reduce((sum, item) => sum + item.price, 0),
)

async function loadCart() {
  error.value = ''
  try {
    const res = await api.get('/api/cart')
    items.value = res.data
  } catch (e) {
    error.value = 'Не удалось загрузить корзину'
    console.error(e)
  }
}

async function clearCart() {
  error.value = ''
  try {
    await api.delete('/api/cart')
    items.value = []
  } catch (e) {
    error.value = 'Не удалось очистить корзину'
    console.error(e)
  }
}

onMounted(loadCart)
</script>
