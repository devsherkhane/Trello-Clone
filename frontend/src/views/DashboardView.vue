<template>
  <main class="content">
    <nav class="board-header">
      <div class="header-left">
        <h2>Trello Clone</h2>
      </div>
      <div class="header-right" style="display: flex; gap: 10px;">
        <router-link to="/profile" class="btn-secondary" style="text-decoration: none; color: white;">
          Profile Settings
        </router-link>
        <button @click="logout" class="btn-secondary">Logout</button>
      </div>
    </nav>
    <h2>Your Boards</h2>
    <div class="board-grid">
      <div class="board-tile create-tile" @click="isModalOpen = true">
        <span>Create new board</span>
      </div>

      <div v-for="board in boards" :key="board.id" class="board-tile">
        <h3 @click="openBoard(board.id)">{{ board.title }}</h3>
        <div class="board-actions">
          <button @click.stop="editBoard(board)">✏️</button>
          <button @click.stop="deleteBoard(board.id)">🗑️</button>
        </div>
      </div>
    </div>
  </main>

  <div v-if="isModalOpen" class="modal-backdrop" @click.self="closeModal">
    <div class="modal-content">
      <h3>{{ editingBoard ? 'Edit Board' : 'Create Board' }}</h3>
      <input v-model="boardFormTitle" placeholder="Enter board title" @keyup.enter="submitBoardForm" ref="modalInput" />
      <div class="modal-actions">
        <button class="btn-cancel" @click="closeModal">Cancel</button>
        <button class="btn-create" :disabled="!boardFormTitle.trim()" @click="submitBoardForm">
          {{ editingBoard ? 'Save' : 'Create' }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue';
import { useAuthStore } from '../stores/auth';
import { useRouter } from 'vue-router';
import api from '../api';
import { useToast } from "vue-toastification";

const auth = useAuthStore();
const router = useRouter();
const toast = useToast();
const boards = ref([]);

const isModalOpen = ref(false);
const boardFormTitle = ref('');
const editingBoard = ref(null);

const fetchBoards = async () => {
  try {
    const response = await api.get('/boards');
    // If response.data is null, default to an empty array []
    boards.value = response.data || []; 
  } catch (error) { 
    toast.error("Failed to fetch boards"); 
  }
};

const openBoard = (id) => router.push(`/board/${id}`);

const editBoard = (board) => {
  editingBoard.value = board;
  boardFormTitle.value = board.title;
  isModalOpen.value = true;
};

const closeModal = () => {
  isModalOpen.value = false;
  editingBoard.value = null;
  boardFormTitle.value = '';
};

const submitBoardForm = async () => {
  if (!boardFormTitle.value.trim()) return;

  try {
    if (editingBoard.value) {
      await api.put(`/boards/${editingBoard.value.id}`, { title: boardFormTitle.value });
      const index = boards.value.findIndex(b => b.id === editingBoard.value.id);
      boards.value[index].title = boardFormTitle.value;
      toast.success("Board updated");
    } else {
      const response = await api.post('/boards', { title: boardFormTitle.value });
      // Ensure boards.value is an array before pushing
      if (!boards.value) boards.value = []; 
      boards.value.push(response.data);
      toast.success("Board created");
    }
    closeModal();
  } catch (err) { toast.error("Error saving board"); }
};

const deleteBoard = async (id) => {
  if (!confirm("Are you sure you want to delete this board?")) return;
  try {
    await api.delete(`/boards/${id}`);
    boards.value = boards.value.filter(b => b.id !== id);
    toast.success("Board deleted");
  } catch (err) { toast.error("Error deleting board"); }
};

onMounted(fetchBoards);
</script>

<style scoped>
/* Add to existing styles */
.board-tile {
  position: relative;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
}

.board-tile h3 {
  flex-grow: 1;
}

.board-actions {
  display: flex;
  justify-content: flex-end;
  gap: 5px;
}

.board-actions button {
  background: transparent;
  border: none;
  cursor: pointer;
  padding: 4px;
}

.board-actions button:hover {
  background: rgba(255, 255, 255, 0.2);
  border-radius: 4px;
}
</style>