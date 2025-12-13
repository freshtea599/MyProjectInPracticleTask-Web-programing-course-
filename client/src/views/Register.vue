<template>
  <main class="container mt-5" style="max-width: 400px;">
    <h2>Регистрация</h2>
    <div class="mb-3">
      <input v-model="username" class="form-control" placeholder="Имя пользователя">
    </div>
    <div class="mb-3">
      <input v-model="password" class="form-control" type="password" placeholder="Пароль">
    </div>
    <div class="mb-3">
      <button class="btn btn-success w-100" @click="submit">Регистрация</button>
    </div>
    <div v-if="msg" class="alert alert-danger">{{ msg }}</div>
    <router-link to="/login">Уже есть аккаунт?</router-link>
  </main>
</template>
<script setup>
import { ref } from 'vue'
import api from '../axios'
import { useRouter } from 'vue-router'

const username = ref('')
const password = ref('')
const msg = ref('')
const router = useRouter()

async function submit() {
  msg.value = ''
  try {
    await api.post('/api/register', {
      username: username.value,
      password: password.value,
    })
    router.push('/login')
  } catch (e) {
    msg.value = 'Ошибка регистрации'
  }
}
</script>

