<template>
  <main class="container text-center mt-5">
    <img :src="person.avatar" class="avatar mb-3" />
    <h2><i :class="person.icon"></i> {{ person.title }}</h2>
    <p>{{ person.description }}</p>
    <p>Email: <a :href="`mailto:${person.email}`">{{ person.email }}</a></p>
    <router-link to="/about" class="btn btn-primary mt-3">
      <i class="bi bi-arrow-left"></i> Назад
    </router-link>
  </main>
</template>

<script setup>
import { useRoute } from 'vue-router'
import { computed } from 'vue'
import person1 from '../assets/image/person-1.jpg'
import person2 from '../assets/image/person-2.jpg'
import person3 from '../assets/image/person-3.jpg'

const PERSONS = {
  'person-1': {
    avatar: person1,
    icon: 'bi bi-cup-hot',
    title: 'Наш бариста',
    description: 'Раньше уничтожал вселенные, а теперь варит кофе. Его капучино способно оживить даже самый долгий созвон в Zoom.',
    email: 'barista@itcoffee.ru'
  },
  'person-2': {
    avatar: person2,
    icon: 'bi bi-lightbulb',
    title: 'Креативный директор',
    description: 'Он пьет столько кофе, что новые идеи появляются быстрее, чем успевает остыть чашка. Осторожно: может заразить энтузиазмом!',
    email: 'creative@itcoffee.ru'
  },
  'person-3': {
    avatar: person3,
    icon: 'bi bi-truck',
    title: 'Логист-бизон',
    description: 'Этот человек доставит кофе даже в тундру. Иногда прилетает на вертолете просто ради фана. Легенда среди клиентов!',
    email: 'logist@itcoffee.ru'
  }
}

const route = useRoute()
const personId = computed(() => route.params.personId)
const person = computed(() => PERSONS[personId.value] || {
  avatar: '',
  icon: 'bi bi-question-circle',
  title: 'Сотрудник',
  description: 'Информация о сотруднике не найдена.',
  email: ''
})
</script>

<style scoped>
.avatar {
  width: 180px;
  height: 180px;
  object-fit: cover;
  border-radius: 50%;
  display: block;
  margin: 0 auto;
  border: 4px solid #f0f0f0;
  box-shadow: 0 4px 12px rgba(0,0,0,0.15);
  transition: transform 0.3s ease;
}
</style>
