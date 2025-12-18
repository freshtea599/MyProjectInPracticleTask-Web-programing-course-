<template>
  <main class="container mt-5" style="max-width:500px;">
    <h2>Профиль</h2>
    <div v-if="msg" class="alert" :class="msgType">{{ msg }}</div>
    <div class="mb-2">
      <label>Имя аккаунта / тег (по умолчанию):</label>
      <div class="form-control mb-2" style="background:#f8f9fa;" disabled>
        {{ displayName }}
      </div>
    </div>
    <div class="mb-2">
      <label>Имя</label>
      <input v-model="profile.first_name" class="form-control" placeholder="Имя" />
    </div>
    <div class="mb-2">
      <label>Фамилия</label>
      <input v-model="profile.last_name" class="form-control" placeholder="Фамилия" />
    </div>
    <div class="mb-2">
      <label>Дата рождения</label>
      <input type="date" v-model="birthdateInput" class="form-control" />
    </div>
    <div class="mb-2">
      <label>Пол</label>
      <select v-model="profile.gender" class="form-select">
        <option value="">—</option>
        <option value="M">Мужской</option>
        <option value="F">Женский</option>
      </select>
    </div>
    <div class="mb-3">
      <button class="btn btn-primary" @click="save">Сохранить</button>
    </div>
  </main>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import api from '../axios'

const profile = ref({
  username: "",
  first_name: "",
  last_name: "",
  birthdate: "",
  is_male: null,
  gender: "",
  profile_tag: ""
})

const msg = ref("")
const msgType = ref("alert-danger")

const birthdateInput = computed({
  get() {
    return profile.value.birthdate ? profile.value.birthdate : ""
  },
  set(val) {
    profile.value.birthdate = val
  }
})

const displayName = computed(() => {
  if ((profile.value.first_name && profile.value.first_name.trim()) || (profile.value.last_name && profile.value.last_name.trim())) {
    return [profile.value.first_name, profile.value.last_name].filter(Boolean).join(" ")
  }
  return profile.value.profile_tag || profile.value.username
})

async function loadProfile() {
  try {
    const resp = await api.get('/api/profile')
    profile.value.first_name  = resp.data.first_name  || ""
    profile.value.last_name   = resp.data.last_name   || ""
    profile.value.birthdate   = resp.data.birthdate   || ""
    profile.value.profile_tag = resp.data.profile_tag || ""
    profile.value.username    = resp.data.username    || ""
    
    profile.value.is_male = resp.data.is_male
    
    if (resp.data.is_male === true) {
        profile.value.gender = "M"
    } else if (resp.data.is_male === false) {
        profile.value.gender = "F"
    } else {
        profile.value.gender = ""
    }

    msg.value = ""
    msgType.value = "alert-success"
  } catch(e) {
    msg.value = "Ошибка загрузки профиля"
    msgType.value = "alert-danger"
  }
}

async function save() {
  try {
    await api.put('/api/profile', {
      first_name: profile.value.first_name,
      last_name: profile.value.last_name,
      birthdate: profile.value.birthdate,
      gender: profile.value.gender
    })
    msg.value = "Профиль обновлён!"
    msgType.value = "alert-success"
    
    loadProfile()
  } catch(e) {
    msg.value = "Ошибка сохранения: " + (e.response?.data?.error || e.message)
    msgType.value = "alert-danger"
  }
}

onMounted(loadProfile)
</script>
