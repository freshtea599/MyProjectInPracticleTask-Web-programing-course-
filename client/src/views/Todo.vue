<template>
  <div class="row py-3">
    <!-- Панель поиска -->
    <div class="col-12 mb-3 d-flex gap-2">
      <input v-model="searchQuery" class="form-control" placeholder="Поиск по группам и задачам..." />
      <select v-model="sortOption" class="form-select" style="max-width: 200px;">
        <option value="date">По дате (новые сверху)</option>
        <option value="alpha">По алфавиту</option>
        <option value="unfinished">Невыполненные сверху</option>
      </select>
    </div>
    <!-- Группы -->
    <div class="col-4 border-end">
      <h5 class="d-flex justify-content-between">
        Группы
        <button class="btn btn-sm btn-primary" @click="addGroupVisible = !addGroupVisible">
          <i class="bi bi-plus-lg"></i>
        </button>
      </h5>
      <div v-if="addGroupVisible" class="input-group mb-2">
        <input v-model="newGroupTitle" type="text" class="form-control" placeholder="Название группы" @keyup.enter="addGroup" />
        <button class="btn btn-success" @click="addGroup"><i class="bi bi-check-lg"></i></button>
      </div>
      <ul class="list-group">
        <li
          v-for="group in filteredAndSortedGroups"
          :key="group.id"
          @click="selectGroup(group.id)"
          class="list-group-item d-flex justify-content-between align-items-center"
          :class="{ active: group.id === selectedGroupId }"
          style="cursor: pointer;"
        >
          <span>{{ group.title }}</span>
          <div>
            <i class="bi bi-pencil me-2 text-primary" @click.stop="editGroup(group)"></i>
            <i class="bi bi-trash text-danger" @click.stop="deleteGroup(group.id)"></i>
          </div>
        </li>
      </ul>
    </div>
    <!-- Задачи -->
    <div class="col-8">
      <template v-if="selectedGroupId">
        <h5 class="d-flex justify-content-between">
          {{ selectedGroup ? selectedGroup.title : "-" }}
          <button class="btn btn-sm btn-primary" @click="addTaskVisible = !addTaskVisible">
            <i class="bi bi-plus-lg"></i>
          </button>
        </h5>
        <div v-if="addTaskVisible" class="input-group mb-2">
          <input v-model="newTaskTitle" type="text" class="form-control" placeholder="Название задачи" @keyup.enter="addTask" />
          <button class="btn btn-success" @click="addTask"><i class="bi bi-check-lg"></i></button>
        </div>
        <ul class="list-group">
          <li
            v-for="task in filteredAndSortedTasks"
            :key="task.id"
            class="list-group-item d-flex justify-content-between align-items-center"
          >
            <div @click="toggleTask(task.id)" style="cursor: pointer;">
              <i :class="task.done ? 'bi bi-check-circle text-success' : 'bi bi-circle text-secondary'"></i>
              <span :class="{ 'text-decoration-line-through text-muted': task.done }" class="ms-2">{{ task.title }}</span>
            </div>
            <div>
              <i class="bi bi-pencil me-2 text-primary" @click.stop="editTask(task)"></i>
              <i class="bi bi-trash text-danger" @click.stop="deleteTask(task.id)"></i>
            </div>
          </li>
        </ul>
      </template>
    </div>
  </div>
</template>

<script>
import api from '../axios';

export default {
  data() {
    return {
      groups: [],
      tasks: [],
      selectedGroupId: null,
      searchQuery: "",
      sortOption: "date",
      addGroupVisible: false,
      addTaskVisible: false,
      newGroupTitle: "",
      newTaskTitle: ""
    };
  },
  computed: {
    selectedGroup() {
      return this.groups.find(g => g.id === this.selectedGroupId) || null;
    },
    filteredAndSortedGroups() {
      let output = [...this.groups];
      if (this.searchQuery.trim()) {
        const q = this.searchQuery.toLowerCase();
        output = output.filter(g => g.title.toLowerCase().includes(q));
      }
      return this.applySort(output, "group");
    },
    filteredAndSortedTasks() {
      let output = [...this.tasks];
      if (this.searchQuery.trim()) {
        const q = this.searchQuery.toLowerCase();
        output = output.filter(t => t.title.toLowerCase().includes(q));
      }
      return this.applySort(output, "task");
    }
  },
  methods: {
    async fetchGroups() {
      const res = await api.get("/api/groups");
      this.groups = res.data;
      if (!this.selectedGroupId && this.groups.length > 0) {
        this.selectGroup(this.groups[0].id);
      }
    },
    async selectGroup(id) {
      this.selectedGroupId = id;
      await this.fetchTasks(id);
    },
    async fetchTasks(groupId) {
      const res = await api.get(`/api/groups/${groupId}/tasks`);
      this.tasks = res.data;
    },
    async addGroup() {
      if (!this.newGroupTitle.trim()) return;
      await api.post("/api/groups", { title: this.newGroupTitle });
      this.newGroupTitle = "";
      this.addGroupVisible = false;
      await this.fetchGroups();
    },
    async editGroup(group) {
      const newTitle = prompt("Введите новое название группы", group.title);
      if (newTitle && newTitle.trim()) {
        await api.put(`/api/groups/${group.id}`, { title: newTitle.trim() });
        await this.fetchGroups();
      }
    },
    async deleteGroup(id) {
      if (confirm("Удалить группу?")) {
        await api.delete(`/api/groups/${id}`);
        if (this.selectedGroupId === id) this.selectedGroupId = null;
        await this.fetchGroups();
      }
    },
    async addTask() {
      if (!this.selectedGroupId || !this.newTaskTitle.trim()) return;
      await api.post(`/api/groups/${this.selectedGroupId}/tasks`, { title: this.newTaskTitle });
      this.newTaskTitle = "";
      this.addTaskVisible = false;
      await this.fetchTasks(this.selectedGroupId);
    },
    async editTask(task) {
      const newTitle = prompt("Введите новое название задачи", task.title);
      if (newTitle && newTitle.trim()) {
        await api.put(`/api/tasks/${task.id}`, { title: newTitle.trim() });
        await this.fetchTasks(this.selectedGroupId);
      }
    },
    async toggleTask(id) {
      const task = this.tasks.find(t => t.id === id);
      if (task) {
        await api.put(`/api/tasks/${id}`, { done: !task.done });
        await this.fetchTasks(this.selectedGroupId);
      }
    },
    async deleteTask(id) {
      if (confirm("Удалить задачу?")) {
        await api.delete(`/api/tasks/${id}`);
        await this.fetchTasks(this.selectedGroupId);
      }
    },
    applySort(list, type) {
      if (this.sortOption === "alpha") {
        return list.sort((a, b) => a.title.localeCompare(b.title));
      }
      if (this.sortOption === "unfinished" && type === "task") {
        return list.sort((a, b) => (a.done === b.done ? 0 : a.done ? 1 : -1));
      }
      return list.sort((a, b) => b.id - a.id);
    }
  },
  async mounted() {
    await this.fetchGroups();
  }
};
</script>

<style scoped>
.list-group-item.active {
  background-color: #0d6dfd2a;
  border-color: #0d6dfd38;
  color: #fff;
}
</style>
