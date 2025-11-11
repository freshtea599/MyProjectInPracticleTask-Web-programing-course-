<template>
  <main class="container">
    <div class="text-center mb-5">
      <h1 class="fw-bold display-5">Айти кофейня - лучшее место для учебы в айти!</h1>
      <p class="lead text-muted">
        Добро пожаловать в лучшее место на планете земля! <br>
        Ведь у нас вы можете не только взбодриться кофейком, но и научиться парочке классных айти штучек, все наши посетители славятся своими hard-скиллами.
        К примеру, этот сайт, сделал человек купивший "Быстрый-Скилл-Капучино" и сверстал сайт за 2 минуты. Купи и ты, чтобы всё почувствовать на себе!
      </p>
    </div>

    <!-- Картинки -->
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

    <h2 class="mt-5 feedback">Отзывы наших гостей</h2>
    <div class="reviews-carousel position-relative mb-5">
      <div
        v-for="(review, idx) in reviews"
        :key="idx"
        :class="['carousel-item text-center p-4', { active: idx === currentReview }]"
        v-show="idx === currentReview"
      >
        <i class="bi bi-person-circle" style="font-size:2rem;"></i>
        <h5 class="mt-2">{{ review.name }}</h5>
        <p>
          <span class="text-warning">
            <i v-for="star in review.stars" :key="star" class="bi bi-star-fill"></i>
          </span>
        </p>
        <p>{{ review.text }}</p>
      </div>
      <button
        @click="prevReview"
        type="button"
        class="carousel-control-prev"
        style="top:50%;left:16px;transform:translateY(-50%);position:absolute;"
        aria-label="Предыдущий"
      >
        <span class="carousel-control-prev-icon" aria-hidden="true"></span>
        <span class="visually-hidden">Предыдущий</span>
      </button>
      <button
        @click="nextReview"
        type="button"
        class="carousel-control-next"
        style="top:50%;right:16px;transform:translateY(-50%);position:absolute;"
        aria-label="Следующий"
      >
        <span class="carousel-control-next-icon" aria-hidden="true"></span>
        <span class="visually-hidden">Следующий</span>
      </button>
    </div>
  </main>
</template>

<script setup>
import { ref } from 'vue';
import ingg1 from '../assets/image/slide-1.png';
import ingg2 from '../assets/image/slide-2.png';
import ingg3 from '../assets/image/slide-3.jpg';

const slides = [
  { img: ingg1, alt: 'Слайд 1' },
  { img: ingg2, alt: 'Слайд 2' },
  { img: ingg3, alt: 'Слайд 3' }
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

const reviews = [
  {
    name: 'Алексей Петров',
    stars: 5,
    text: "Выпил Латте 'Bug-Fix' — и баги исчезли сами собой! Теперь думаю, не открыть ли филиал этой кофейни прямо у себя в офисе?"
  },
  {
    name: 'Мария Иванова',
    stars: 4,
    text: "Эспрессо 'Джун-На-Стероидах' помог понять, что я уже почти Middle! Правда, кот дома теперь требует кофеиновые апдейты каждое утро."
  },
  {
    name: 'Игорь Смирнов',
    stars: 5,
    text: "Выпил американо 'Git-Push' — и внезапно отправил все домашки за семестр. Теперь преподаватель зовет меня на кофе чаще, чем друзей."
  }
]
const currentReview = ref(0)

function prevReview() {
  currentReview.value = (currentReview.value - 1 + reviews.length) % reviews.length
}
function nextReview() {
  currentReview.value = (currentReview.value + 1) % reviews.length
}
</script>

<style scoped>
.carousel-item {
  /* Скрывает неактивные элементы, важно для Vue-карусели! */
  display: none;
}
.carousel-item.active {
  display: block;
}
.carousel-indicators .active {
  background-color: #222;
}
.carousel-indicators button {
  background-color: #ddd;
  border: none;
}
</style>
