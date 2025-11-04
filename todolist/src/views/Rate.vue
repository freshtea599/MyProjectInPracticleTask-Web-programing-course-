<template>
  <div class="rate-page text-center">
    <h2 class="mb-3">Оцените приложение</h2>
    <div class="mb-4">
      <i
        v-for="star in 5"
        :key="star"
        class="bi"
        :class="star <= rating ? 'bi-star-fill text-warning' : 'bi-star text-secondary'"
        style="font-size: 2rem; cursor: pointer;"
        @click="setRating(star)"
      ></i>
    </div>

    <p v-if="rating" class="lead">
      Ваша оценка: <strong>{{ rating }}</strong> {{ ratingWord }}
    </p>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'

const rating = ref(0)

function setRating(value) {
  rating.value = value
}

const ratingWord = computed(() => {
  switch (rating.value) {
    case 1: return 'звезда'
    case 2:
    case 3:
    case 4: return 'звезды'
    case 5: return 'звёзд'
    default: return ''
  }
})
</script>
