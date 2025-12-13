<template>
  <main class="container mt-5" style="max-width: 400px;">
    <h2>Вход</h2>
    <div class="mb-3">
      <input v-model="username" class="form-control" placeholder="Имя пользователя">
    </div>
    <div class="mb-3">
      <input v-model="password" class="form-control" type="password" placeholder="Пароль">
    </div>
    <div class="mb-3">
      <button class="btn btn-primary w-100" @click="submit">Войти</button>
    </div>
    <div v-if="msg" class="alert alert-danger">{{ msg }}</div>
    <router-link to="/register">Нет аккаунта? Зарегистрируйтесь</router-link>
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
    const res = await api.post('/api/login', {
      username: username.value,
      password: password.value,
    })
    localStorage.setItem('token', res.data.token)
    localStorage.setItem('role', res.data.role)
    router.push('/profile')
  } catch (e) {
    msg.value = 'Неверный логин или пароль'
  }
}
</script>
