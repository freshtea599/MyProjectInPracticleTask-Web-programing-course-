<template>
  <main class="container mt-5" style="max-width: 400px;">
    <h2 class="mb-4 text-center">Регистрация</h2>
    
    <form @submit.prevent="submit">
      <div class="mb-3">
        <label class="form-label">Имя пользователя</label>
        <input 
          v-model="username" 
          class="form-control" 
          placeholder="Придумайте логин" 
          required
        >
      </div>
      
      <div class="mb-3">
        <label class="form-label">Пароль</label>
        <input 
          v-model="password" 
          class="form-control" 
          type="password" 
          placeholder="Введите пароль" 
          required
        >
      </div>

      <div class="mb-3">
        <label class="form-label">Подтвердите пароль</label>
        <input 
          v-model="confirmPassword" 
          class="form-control" 
          type="password" 
          placeholder="Повторите пароль" 
          required
        >
      </div>

      <div class="mb-3">
        <button class="btn btn-success w-100" type="submit">Зарегистрироваться</button>
      </div>
    </form>

    <div v-if="msg" class="alert alert-danger text-center">{{ msg }}</div>
    
    <div class="text-center mt-3">
      <router-link to="/login" class="text-decoration-none">Уже есть аккаунт? Войти</router-link>
    </div>
  </main>
</template>

<script setup>
import { ref } from 'vue'
import api from '../axios'
import { useRouter } from 'vue-router'

const username = ref('')
const password = ref('')
const confirmPassword = ref('')
const msg = ref('')
const router = useRouter()

async function submit() {
  msg.value = ''

  if (password.value !== confirmPassword.value) {
    msg.value = 'Пароли не совпадают!'
    return
  }

  if (password.value.length < 8) {
    msg.value = 'Пароль должен быть не менее 8 символов'
    return
  }

  try {

    await api.post('/api/register', {
      username: username.value,
      password: password.value,
    })
    
    router.push('/login')
  } catch (e) {

    msg.value = e.response?.data?.error || 'Ошибка регистрации. Попробуйте другой логин.'
  }
}
</script>
