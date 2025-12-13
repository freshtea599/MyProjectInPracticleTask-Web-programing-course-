<template>
  <main class="container">
    <div class="text-center mb-5">
      <h1 class="fw-bold display-5">Айти кофейня - лучшее место для учебы в айти!</h1>
      <p class="lead text-muted">
        Заказ или предзаказ, вот в чём вопрос... Только будьте аккуратны, наш кофе вызывает привыкание!
      </p>
    </div>

    <!-- Загрузка -->
    <div v-if="allProducts.length === 0" class="text-center mt-5">
      <div class="spinner-border" role="status">
        <span class="visually-hidden">Загрузка...</span>
      </div>
    </div>

    <!-- Список товаров (только текущая страница) -->
    <div v-else>
      <div
        v-for="product in paginatedProducts"
        :key="product.uniqueKey"
        class="card shadow rounded my-3"
      >
        <div class="card-body">
          <div class="row">
            <div
              class="col-md-2 text-center d-flex align-items-center justify-content-center"
            >
              <img
                v-if="product.image"
                :src="product.image"
                class="product-image rounded"
                width="90"
                style="max-height: 90px; object-fit: cover;"
              />
              <div
                v-else
                class="bg-light rounded d-flex align-items-center justify-content-center"
                style="width:90px; height:90px;"
              >
                <i class="bi bi-cup-hot text-muted" style="font-size: 2rem;"></i>
              </div>
            </div>
            <div class="col-md-10 d-flex flex-column">
              <div class="product-name fw-bold fs-5">{{ product.name }}</div>
              <div class="product-description text-muted mb-2">
                {{ product.description }}
              </div>
              <div class="flex-grow-1"></div>
              <div class="d-flex justify-content-end align-items-center mt-2">
                <div class="product-price me-3 fs-5">
                  Цена: <strong>{{ product.price }} р.</strong>
                </div>
                <button class="btn btn-warning" @click="addToCart(product)">
                  <i class="bi bi-cart"></i>
                  <span class="d-none d-md-inline ms-1">В корзину</span>
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Пагинация -->
      <nav
        v-if="totalPages > 1"
        class="d-flex justify-content-center mt-4"
        aria-label="Навигация по товарам"
      >
        <ul class="pagination">
          <li :class="['page-item', { disabled: currentPage === 1 }]">
            <button class="page-link" @click="goToPage(currentPage - 1)">
              Назад
            </button>
          </li>

          <li
            v-for="page in totalPages"
            :key="page"
            :class="['page-item', { active: page === currentPage }]"
          >
            <button class="page-link" @click="goToPage(page)">
              {{ page }}
            </button>
          </li>

          <li :class="['page-item', { disabled: currentPage === totalPages }]">
            <button class="page-link" @click="goToPage(currentPage + 1)">
              Вперёд
            </button>
          </li>
        </ul>
      </nav>
    </div>

    <p v-if="error" class="text-danger text-center">{{ error }}</p>
  </main>
</template>

<script>
import api from '../axios'
import ingg1 from '../assets/image/coffe-capuccino.jpg'
import ingg2 from '../assets/image/coffe-americano.png'
import ingg3 from '../assets/image/coffe-esspreso.jpg'
import ingg4 from '../assets/image/coffe-latte.jpg'
import ingg5 from '../assets/image/coffe-lavanda.jpg'
import ingg6 from '../assets/image/coffe-snickers.jpg'

export default {
  data() {
    return {
      error: '',
      // Статичные товары
      staticProducts: [
        {
          id: 101,
          name: 'Быстрый-Скилл-Капучино',
          image: ingg1,
          price: 450,
          description:
            'Хочешь быстро вкачать какой-либо навык в айти? Бери этот божественный напиток и укажи в комментарие к заказу свое пожелание и уже с первого глотка ты прозреешь!',
        },
        {
          id: 102,
          name: 'Git-Push',
          image: ingg2,
          price: 350,
          description:
            'На тебе давно весят не выполненные таски? Не беда! Бери данный напиток и вся твоя работа будет доделана и запушена на гит. ПРЕДУПРЕЖДЕНИЕ!!! Git-Push американо действует раз в сутки, вторая кружка может вызвать не предвиденные баги в вашем коде!',
        },
        {
          id: 103,
          name: 'Джун-На-Стероидах',
          image: ingg3,
          price: 1050,
          description:
            'Это горькое месиво пьют только самые отчаянные сорви головы! Не мудрено, ведь именно этот напиток помогает встать на путь истинный... Ну и немного прибавляет к харизме!',
        },
        {
          id: 104,
          name: 'Bug-Fix',
          image: ingg4,
          price: 550,
          description:
            'Замучали баги? Есть решение! Смело бери Латте Bug-Fix и смотри как твориться магия! Твой код сам берет и исправляет все баги, тебе лишь остается закусить печенькой...',
        },
        {
          id: 105,
          name: 'Лаванда-Де-Вайбкодинг',
          image: ingg5,
          price: 10,
          description:
            'Что может быть хуже лавандового рафа на кокосовом молоке? Правильно, ничего! Ведь только самые отбитые программисты могут пить эту гадость, ну а мы что? Добавим немного чистого кода и код начнет писать сам себя... всё равно фу? Да, мы в курсе...',
        },
        {
          id: 106,
          name: 'Кайф-На-Удаленке',
          image: ingg6,
          price: 199,
          description:
            'Устал от работы в офисе? Всё проще чем кажется, три глотка раф сникерс и твой начальник уже сегодня тебя отпустить балдеть на удаленке!',
        },
      ],
      // Товары из БД
      dbProducts: [],
      // Пагинация
      currentPage: 1,
      pageSize: 4,
    }
  },
  computed: {
    allProducts() {
      const staticList = this.staticProducts.map((p) => ({
        ...p,
        uniqueKey: `static-${p.id}`,
      }))
      const dbList = this.dbProducts.map((p) => ({
        id: p.id,
        name: p.name,
        description: p.description,
        price: p.price,
        image: p.image_url,
        uniqueKey: `db-${p.id}`,
      }))
      return [...staticList, ...dbList]
    },
    totalPages() {
      return Math.ceil(this.allProducts.length / this.pageSize) || 1
    },
    paginatedProducts() {
      const start = (this.currentPage - 1) * this.pageSize
      return this.allProducts.slice(start, start + this.pageSize)
    },
  },
  async mounted() {
    await this.loadDbProducts()
  },
  methods: {
    async loadDbProducts() {
      try {
        const res = await api.get('/api/products')
        this.dbProducts = res.data || []
      } catch (e) {
        console.error('Ошибка загрузки товаров из БД', e)
      }
    },
    async addToCart(product) {
      this.error = ''
      try {
        await api.post('/api/cart', {
          id: product.id,
          name: product.name,
          image: product.image || '',
          price: product.price,
        })
        alert('Товар добавлен в корзину')
      } catch (e) {
        this.error = 'Не удалось добавить в корзину'
        console.error(e)
      }
    },
    goToPage(page) {
      if (page < 1 || page > this.totalPages) return
      this.currentPage = page
      window.scrollTo({ top: 0, behavior: 'smooth' })
    },
  },
}
</script>

<style scoped>
.product-image {
  width: 100%;
  border-radius: 0.5rem;
  border: 1px solid rgb(210, 210, 210);
}

.product-name {
  font-size: 1.2rem;
}

.product-description {
  font-size: 0.95rem;
  color: rgb(60, 60, 60);
}

.product-price {
  padding: 0.5rem;
  border-radius: 0.5rem;
  border: 1px solid rgb(210, 210, 210);
}

.pagination {
  flex-wrap: wrap;
  gap: 0.25rem;
}

.pagination .page-link {
  padding: 0.4rem 0.6rem;
  font-size: 0.9rem;
}

/* Мобильные (до 576px) */
@media (max-width: 575.98px) {
  .card {
    margin: 1rem 0;
  }

  .card-body {
    padding: 0.75rem;
  }

  .row {
    flex-direction: column;
  }

  .col-md-2 {
    width: 100%;
    margin-bottom: 0.5rem;
  }

  .col-md-10 {
    width: 100%;
  }

  .product-name {
    font-size: 1rem;
  }

  .product-description {
    font-size: 0.85rem;
  }

  .product-price {
    font-size: 0.9rem;
  }

  .btn {
    padding: 0.375rem 0.75rem;
    font-size: 0.85rem;
  }

  .pagination .page-link {
    padding: 0.3rem 0.5rem;
    font-size: 0.75rem;
  }

  .pagination {
    margin-top: 1.5rem;
  }
}

/* Планшеты (576px - 768px) */
@media (min-width: 576px) and (max-width: 767.98px) {
  .product-name {
    font-size: 1.1rem;
  }

  .product-description {
    font-size: 0.9rem;
  }

  .pagination .page-link {
    padding: 0.35rem 0.55rem;
    font-size: 0.8rem;
  }
}

/* Десктоп (768px и выше) */
@media (min-width: 768px) {
  .product-name {
    font-size: 1.2rem;
  }

  .pagination .page-link {
    padding: 0.4rem 0.6rem;
    font-size: 0.9rem;
  }
}
</style>
