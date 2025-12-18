<template>
  <main class="container">
    <div class="text-center mb-5">
      <h1 class="fw-bold display-5">Айти кофейня - лучшее место для учебы в айти!</h1>
      <p class="lead text-muted">
        Заказ или предзаказ, вот в чём вопрос... Только будьте аккуратны, наш кофе вызывает привыкание!
      </p>
    </div>

    <div v-if="loading" class="text-center mt-5">
      <div class="spinner-border" role="status">
        <span class="visually-hidden">Загрузка...</span>
      </div>
    </div>
    <div v-else-if="products.length === 0" class="alert alert-info text-center">
      В данный момент товаров нет в наличии. Загляните позже!
    </div>

    <div v-else>
      <div
        v-for="product in paginatedProducts"
        :key="product.id"
        class="card shadow rounded my-3"
      >
        <div class="card-body">
          <div class="row">
            <div class="col-md-2 text-center d-flex align-items-center justify-content-center">
              <img
                v-if="product.image_url"
                :src="product.image_url"
                class="product-image rounded"
                width="90"
                style="max-height: 90px; object-fit: cover;"
                alt="Product Image"
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

    <p v-if="error" class="text-danger text-center mt-3">{{ error }}</p>
  </main>
</template>

<script>
import api from '../axios'

export default {
  data() {
    return {
      products: [],
      loading: false,
      error: '',
      currentPage: 1,
      pageSize: 4,
    }
  },
  computed: {
    totalPages() {
      return Math.ceil(this.products.length / this.pageSize) || 1
    },
    paginatedProducts() {
      const start = (this.currentPage - 1) * this.pageSize
      return this.products.slice(start, start + this.pageSize)
    },
  },
  async mounted() {
    await this.loadProducts()
  },
  methods: {
    async loadProducts() {
      this.loading = true
      this.error = ''
      try {
        const res = await api.get('/api/products')
        this.products = res.data || []
      } catch (e) {
        this.error = 'Не удалось загрузить товары. Попробуйте позже.'
        console.error('Ошибка загрузки товаров:', e)
      } finally {
        this.loading = false
      }
    },
    async addToCart(product) {
      this.error = ''
      try {
        await api.post('/api/cart', {
          product_id: product.id,
        })
        alert('Товар успешно добавлен в корзину!')
      } catch (e) {
        this.error = 'Не удалось добавить товар в корзину.'
        console.error('Ошибка добавления в корзину:', e)
        if (e.response && e.response.data && e.response.data.error) {
             this.error = `Ошибка: ${e.response.data.error}`
        }
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
  cursor: pointer;
}

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
