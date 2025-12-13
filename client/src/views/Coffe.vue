<template>
  <main class="container">
    <div class="text-center mb-5">
      <h1 class="fw-bold display-5">Айти кофейня - лучшее место для учебы в айти!</h1>
      <p class="lead text-muted">
        Добро пожаловать в лучшее место на планете земля! <br>
        Ведь у нас вы можете не только взбодриться кофейком, но и научиться парочке классных айти штучек, все наши посетители славятся своими hard-скиллами.
        К примеру, этот сайт сделал человек, купивший "Быстрый-Скилл-Капучино", и сверстал сайт за 2 минуты. Купи и ты, чтобы всё почувствовать на себе!
      </p>
    </div>

    <!-- Слайдер картинок -->
    <div class="carousel slide shadow rounded overflow-hidden mb-5">
      <!-- Индикаторы -->
      <div class="carousel-indicators mb-0">
        <button
          v-for="(slide, index) in slides"
          :key="index"
          :class="{ active: index === currentSlide }"
          @click="goToSlide(index)"
          type="button"
          style="width:10px;height:10px;border-radius:50%;margin:1px;"
        ></button>
      </div>
      <!-- Слайды -->
      <div class="carousel-inner">
        <div
          v-for="(slide, index) in slides"
          :key="index"
          :class="['carousel-item', { active: index === currentSlide }]"
          v-show="index === currentSlide"
        >
          <img :src="slide.img" class="d-block w-100" :alt="slide.alt" />
        </div>
      </div>
      <!-- Кнопки управления -->
      <button class="carousel-control-prev" @click="prevSlide" type="button" aria-label="Previous">
        <span class="carousel-control-prev-icon"></span>
      </button>
      <button class="carousel-control-next" @click="nextSlide" type="button" aria-label="Next">
        <span class="carousel-control-next-icon"></span>
      </button>
    </div>

    <!-- Единый блок отзывов -->
    <h2 class="mt-5 feedback text-center">Отзывы наших гостей</h2>

    <div class="reviews-carousel position-relative mb-5 text-center">
      <div
        v-if="allReviews.length === 0"
        class="p-4 text-muted"
      >
        Отзывов пока нет. Будьте первым!
      </div>

      <div
        v-else
        v-for="(review, idx) in allReviews"
        :key="review.id || `static-${idx}`"
        :class="['carousel-item p-4', { active: idx === currentReview }]"
        v-show="idx === currentReview"
      >
        <i class="bi bi-person-circle" style="font-size:2rem;"></i>
        <h5 class="mt-2">{{ review.username }}</h5>
        <p>
          <span class="text-warning">
            <i v-for="star in review.rating" :key="star" class="bi bi-star-fill"></i>
          </span>
        </p>
        <p class="lead fst-italic">"{{ review.comment }}"</p>
        <small v-if="review.created_at" class="text-muted d-block mt-2">
          {{ new Date(review.created_at).toLocaleDateString('ru-RU') }}
        </small>
      </div>

      <button
        v-if="allReviews.length > 1"
        @click="prevReview"
        type="button"
        class="carousel-control-prev"
        style="top:50%;left:0;transform:translateY(-50%);position:absolute;"
      >
        <span class="carousel-control-prev-icon" aria-hidden="true"></span>
      </button>
      <button
        v-if="allReviews.length > 1"
        @click="nextReview"
        type="button"
        class="carousel-control-next"
        style="top:50%;right:0;transform:translateY(-50%);position:absolute;"
      >
        <span class="carousel-control-next-icon" aria-hidden="true"></span>
      </button>
    </div>
  </main>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import api from '../axios'
import ingg1 from '../assets/image/slide-1.png'
import ingg2 from '../assets/image/slide-2.png'
import ingg3 from '../assets/image/slide-3.jpg'

// --- СЛАЙДЕР ---
const slides = [
  { img: ingg1, alt: 'Слайд 1' },
  { img: ingg2, alt: 'Слайд 2' },
  { img: ingg3, alt: 'Слайд 3' },
]
const currentSlide = ref(0)

function goToSlide(index) {
  currentSlide.value = index
}
function prevSlide() {
  currentSlide.value = (currentSlide.value - 1 + slides.length) % slides.length
}
function nextSlide() {
  currentSlide.value = (currentSlide.value + 1) % slides.length
}

// --- СТАТИЧНЫЕ ОТЗЫВЫ ---
const staticReviews = [
  {
    username: 'Алексей Петров',
    rating: 5,
    comment:
      "Выпил Латте 'Bug-Fix' — и баги исчезли сами собой! Теперь думаю, не открыть ли филиал этой кофейни прямо у себя в офисе?",
  },
  {
    username: 'Мария Иванова',
    rating: 4,
    comment:
      "Эспрессо 'Джун-На-Стероидах' помог понять, что я уже почти Middle! Правда, кот дома теперь требует кофеиновые апдейты каждое утро.",
  },
  {
    username: 'Игорь Смирнов',
    rating: 5,
    comment:
      "Выпил американо 'Git-Push' — и внезапно отправил все домашки за семестр. Теперь преподаватель зовет меня на кофе чаще, чем друзей.",
  },
]

// --- РЕАЛЬНЫЕ ОТЗЫВЫ ИЗ БД ---
const realReviews = ref([])
const currentReview = ref(0)

const allReviews = computed(() => {
  const mappedReal = realReviews.value.map((r) => ({
    id: r.id,
    username: r.username,
    rating: r.rating,
    comment: r.comment,
    created_at: r.created_at,
  }))
  
  return [...mappedReal, ...staticReviews]
})

async function loadRealReviews() {
  try {
    const res = await api.get('/api/reviews')
    realReviews.value = res.data || []
  } catch (e) {
    console.error('Failed to load real reviews', e)
  }
}

function prevReview() {
  if (allReviews.value.length === 0) return
  currentReview.value =
    (currentReview.value - 1 + allReviews.value.length) % allReviews.value.length
}
function nextReview() {
  if (allReviews.value.length === 0) return
  currentReview.value = (currentReview.value + 1) % allReviews.value.length
}

onMounted(() => {
  loadRealReviews()
})
</script>

<style scoped>
.carousel-item {
  display: none;
  transition: opacity 0.6s ease-in-out;
  opacity: 0;
}
.carousel-item.active {
  display: block;
  opacity: 1;
}
.carousel-indicators .active {
  background-color: #222;
}
.carousel-indicators button {
  background-color: #ddd;
  border: none;
}
.carousel-control-prev-icon,
.carousel-control-next-icon {
  background-color: #000;
  border-radius: 50%;
  padding: 10px;
}
</style>
