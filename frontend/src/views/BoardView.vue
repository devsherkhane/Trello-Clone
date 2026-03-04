<template>
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
  <div class="board-canvas">
    <nav class="board-header">
      <div class="header-left" style="display: flex; align-items: center;">
        <h2 style="margin: 0;">{{ boardTitle }}</h2>
        <button class="btn-secondary" style="margin-left: 15px; padding: 6px 10px;" @click="exportBoard">⬇️ Export
          CSV</button>
      </div>
      <div class="search-wrapper">
        <input v-model="searchQuery" placeholder="Filter cards..." class="search-input" />
      </div>
    </nav>

    <div v-if="isLoading" class="lists-container">
      <div v-for="i in 3" :key="i" class="list-wrapper skeleton-wrapper">
        <div class="list-content skeleton-list">
          <div class="skeleton-header"></div>
          <div v-for="j in 4" :key="j" class="skeleton-card"></div>
        </div>
      </div>
    </div>

    <div v-else class="lists-container">
      <div v-for="list in filteredLists" :key="list.id" class="list-wrapper">
        <div class="list-content">

          <div class="list-header"
            style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 10px;">
            <input v-model="list.title" @change="updateListTitle(list.id, list.title)"
              style="background: transparent; border: none; font-weight: bold; width: 80%; font-size: 16px; color: #172b4d;" />
            <button @click="deleteList(list.id)"
              style="background: transparent; border: none; cursor: pointer; font-size: 14px;">🗑️</button>
          </div>

          <draggable v-model="list.cards" group="cards" item-key="id" class="cards-list" ghost-class="ghost-card"
            @change="onCardMove($event, list.id)">
            <template #item="{ element }">
              <div class="card" @click="openCardDetails(element, list.title)">
                <div v-if="element.label_color" class="card-label-bar"
                  :style="{ backgroundColor: element.label_color }"></div>
                {{ element.title }}
              </div>
            </template>
          </draggable>

          <div v-if="currentlyEditingList === list.id" class="add-card-form">
            <textarea v-model="newCardTitle" placeholder="Enter a title..."
              @keyup.enter="submitCard(list.id)"></textarea>
            <div class="form-actions">
              <button class="btn-add" @click="submitCard(list.id)">Add card</button>
              <button class="btn-close" @click="currentlyEditingList = null">✕</button>
            </div>
          </div>
          <button v-else class="add-card-btn" @click="openAddCard(list.id)">+ Add a card</button>
        </div>
      </div>

      <div class="list-wrapper">
        <div v-if="!isAddingList" class="add-list-placeholder" @click="isAddingList = true">+ Add another list</div>
        <div v-else class="add-list-form">
          <input v-model="newListTitle" placeholder="Enter list title..." @keyup.enter="submitList" autofocus />
          <div class="form-actions">
            <button class="btn-add" @click="submitList">Add list</button>
            <button class="btn-close" @click="isAddingList = false">✕</button>
          </div>
        </div>
      </div>
    </div>

    <div v-if="activeCard" class="modal-backdrop" @click.self="activeCard = null">
      <div class="card-detail-modal">
        <button class="modal-close" @click="activeCard = null">✕</button>

        <div class="modal-section" style="display: flex; justify-content: space-between; align-items: flex-start;">
          <div>
            <h3 style="margin-top: 0;">💳 {{ activeCard.title }}</h3>
            <p class="subtitle" style="margin-top: 5px;">in list <u>{{ activeCard.listName }}</u></p>
          </div>
          <button @click="deleteCard" class="btn-cancel"
            style="color: #c92a2a; border: 1px solid #c92a2a; padding: 6px 12px; border-radius: 4px; background: transparent; cursor: pointer; font-weight: bold;">
            Delete Card
          </button>
        </div>

        <div class="modal-section">
          <h4>Labels</h4>
          <div class="labels-picker">
            <div v-for="color in ['#61bd4f', '#f2d600', '#ff9f1a', '#eb5a46', '#c377e0', '#0079bf']" :key="color"
              class="label-swatch" :style="{ backgroundColor: color }"
              :class="{ active: activeCard.label_color === color }" @click="updateCardLabel(color)"></div>
            <button class="btn-clear-label" @click="updateCardLabel(null)">None</button>
          </div>
        </div>

        <div class="modal-section">
          <h4>Description</h4>
          <textarea v-model="activeCard.description" placeholder="Add a description..."
            @blur="updateCardDescription"></textarea>
        </div>

        <div class="modal-section">
          <h4>Attachments</h4>
          <div class="attachments-list">
            <div v-for="file in activeCard.attachments" :key="file.id" class="attachment-item">
              <span class="file-icon">📄</span>
              <div class="file-info">
                <a :href="file.url" target="_blank">{{ file.filename }}</a>
              </div>
            </div>
          </div>
          <input type="file" ref="fileInput" style="display: none" @change="handleFileUpload" />
          <button class="btn-secondary" @click="$refs.fileInput.click()">📎 Attach a file</button>
        </div>

        <div class="modal-section comments-section">
          <h4>Activity</h4>
          <div class="comment-input-area">
            <textarea v-model="newCommentText" placeholder="Write a comment..."></textarea>
            <button class="btn-add" @click="submitComment" :disabled="!newCommentText.trim()">Save</button>
          </div>
          <div class="comments-list">
            <div v-for="comment in activeCard.comments" :key="comment.id" class="comment-item">
              <div class="comment-avatar">{{ comment.user_name?.charAt(0) || 'U' }}</div>
              <div class="comment-content">
                <span class="comment-user">{{ comment.user_name }}</span>
                <p>{{ comment.text }}</p>
                <span class="comment-date">{{ new Date(comment.created_at).toLocaleString() }}</span>
              </div>
            </div>
          </div>
        </div>

      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue';
import { useRoute } from 'vue-router';
import draggable from 'vuedraggable';
import api from '../api';
import { useWebsocket } from '../api/websocket';
import { useToast } from "vue-toastification";

const toast = useToast();
const { connect, disconnect } = useWebsocket();
const route = useRoute();

// Board State
const boardTitle = ref('Loading...');
const lists = ref([]);
const searchQuery = ref('');
const isLoading = ref(true);

// UI State
const currentlyEditingList = ref(null);
const newCardTitle = ref('');
const isAddingList = ref(false);
const newListTitle = ref('');
const activeCard = ref(null);
const newCommentText = ref('');


// --- List Actions ---
const deleteList = async (listId) => {
  if (!confirm("Are you sure you want to delete this list and all its cards?")) return;
  try {
    await api.delete(`/lists/${listId}`);
    lists.value = lists.value.filter(l => l.id !== listId);
    toast.success("List deleted");
  } catch (err) { toast.error("Failed to delete list"); }
};

const updateListTitle = async (listId, newTitle) => {
  if (!newTitle.trim()) return;
  try {
    await api.put(`/lists/${listId}`, { title: newTitle });
    toast.success("List title updated");
  } catch (err) { toast.error("Failed to update list"); }
};

// --- Card Actions ---
const deleteCard = async () => {
  if (!confirm("Are you sure you want to delete this card?")) return;
  try {
    await api.delete(`/cards/${activeCard.value.id}`);

    // Remove from UI
    const list = lists.value.find(l => l.title === activeCard.value.listName);
    if (list) {
      list.cards = list.cards.filter(c => c.id !== activeCard.value.id);
    }

    activeCard.value = null;
    toast.success("Card deleted");
  } catch (err) { toast.error("Failed to delete card"); }
};

// Fetch attachments and open card
const openCardDetails = async (card, listName) => {
  activeCard.value = { ...card, listName };

  // Fetch attachments for this card
  try {
    const res = await api.get(`/cards/${card.id}/attachments`);
    activeCard.value.attachments = res.data || [];
  } catch (err) {
    console.error("Failed to load attachments");
  }
};

// --- Board Actions ---
const exportBoard = () => {
  window.open(`http://localhost:8080/api/boards/${route.params.id}/export?token=${localStorage.getItem('token')}`, '_blank');
};

const filteredLists = computed(() => {
  if (!searchQuery.value.trim()) return lists.value;
  const query = searchQuery.value.toLowerCase();
  return lists.value.map(list => ({
    ...list,
    cards: list.cards?.filter(card =>
      card.title.toLowerCase().includes(query) ||
      card.description?.toLowerCase().includes(query)
    )
  }));
});

// API Actions
const fetchBoardDetails = async () => {
  isLoading.value = true;
  try {
    const response = await api.get(`/boards/${route.params.id}`);
    boardTitle.value = response.data.title;
    lists.value = response.data.lists || [];
  } catch (err) {
    toast.error("Failed to load board");
  } finally {
    isLoading.value = false;
  }
};

const handleFileUpload = async (event) => {
  const file = event.target.files[0];
  if (!file) return;
  const formData = new FormData();
  formData.append('file', file);
  try {
    const response = await api.post(`/cards/${activeCard.value.id}/attachments`, formData);
    if (!activeCard.value.attachments) activeCard.value.attachments = [];
    activeCard.value.attachments.push(response.data);
    toast.success("File attached!");
  } catch (err) { toast.error("Upload failed"); }
};

const updateCardLabel = async (color) => {
  activeCard.value.label_color = color;
  try {
    await api.patch(`/cards/${activeCard.value.id}`, { label_color: color });
  } catch (err) { toast.error("Label update failed"); }
};

const submitComment = async () => {
  try {
    const response = await api.post(`/cards/${activeCard.value.id}/comments`, { text: newCommentText.value });
    if (!activeCard.value.comments) activeCard.value.comments = [];
    activeCard.value.comments.unshift(response.data);
    newCommentText.value = '';
  } catch (err) { toast.error("Comment failed"); }
};

const submitList = async () => {
  if (!newListTitle.value.trim()) return;
  try {
    const response = await api.post(`/boards/${route.params.id}/lists`, { title: newListTitle.value.trim() });
    lists.value.push({ ...response.data, cards: [] });
    newListTitle.value = '';
    isAddingList.value = false;
  } catch (err) { toast.error("List creation failed"); }
};

const openAddCard = (listId) => {
  currentlyEditingList.value = listId;
  newCardTitle.value = '';
};

const submitCard = async (listId) => {
  if (!newCardTitle.value.trim()) return;
  try {
    const response = await api.post('/cards', { list_id: listId, title: newCardTitle.value.trim() });
    const targetList = lists.value.find(l => l.id === listId);
    if (!targetList.cards) targetList.cards = [];
    targetList.cards.push(response.data);
    newCardTitle.value = '';
    currentlyEditingList.value = null;
  } catch (err) { toast.error("Card creation failed"); }
};

const onCardMove = async (event, newListId) => {
  if (event.added || event.moved) {
    const card = event.added ? event.added.element : event.moved.element;
    const newIndex = event.added ? event.added.newIndex : event.moved.newIndex;
    try {
      await api.patch(`/cards/${card.id}/move`, { list_id: newListId, position: newIndex });
    } catch (err) {
      toast.error("Sync failed");
      fetchBoardDetails();
    }
  }
};

const updateCardDescription = async () => {
  try {
    await api.patch(`/cards/${activeCard.value.id}`, { description: activeCard.value.description });
  } catch (err) { toast.error("Update failed"); }
};

onMounted(() => {
  fetchBoardDetails();
  connect(route.params.id);
});
onUnmounted(() => disconnect());
</script>

<style scoped>
/* Main Layout */
.board-canvas {
  height: 100vh;
  background-color: var(--trello-blue);
  display: flex;
  flex-direction: column;
}

.board-header {
  padding: 10px 20px;
  background: rgba(0, 0, 0, 0.15);
  color: white;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.search-input {
  background: rgba(255, 255, 255, 0.2);
  border: none;
  border-radius: 3px;
  padding: 6px 12px;
  color: white;
  width: 250px;
}

.search-input:focus {
  background: white;
  color: #172b4d;
  outline: none;
}

.lists-container {
  display: flex;
  align-items: flex-start;
  padding: 10px;
  gap: 12px;
  overflow-x: auto;
  flex-grow: 1;
}

.list-wrapper {
  width: 272px;
  flex-shrink: 0;
}

.list-content {
  background: var(--trello-gray);
  border-radius: 3px;
  padding: 10px;
  display: flex;
  flex-direction: column;
  max-height: 100%;
}

/* Card Styling */
.card {
  background: white;
  padding: 10px;
  margin-bottom: 8px;
  border-radius: 3px;
  box-shadow: 0 1px 0 rgba(9, 30, 66, 0.25);
  cursor: grab;
  font-size: 14px;
  position: relative;
}

.card-label-bar {
  height: 4px;
  width: 40px;
  border-radius: 2px;
  margin-bottom: 4px;
}

/* Modal & UI Elements */
.modal-backdrop {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background: rgba(0, 0, 0, 0.6);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 100;
}

.card-detail-modal {
  background: #f4f5f7;
  width: 600px;
  padding: 30px;
  border-radius: 3px;
  position: relative;
  max-height: 90vh;
  overflow-y: auto;
}

.modal-close {
  position: absolute;
  top: 15px;
  right: 15px;
  background: transparent;
  border: none;
  font-size: 18px;
  cursor: pointer;
  color: #6b778c;
}

.modal-section {
  margin-bottom: 24px;
}

.labels-picker {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
  margin-top: 10px;
}

.label-swatch {
  width: 40px;
  height: 32px;
  border-radius: 3px;
  cursor: pointer;
}

.label-swatch.active {
  border: 2px solid #172b4d;
}

.attachment-item {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 8px;
  background: #fff;
  padding: 8px;
  border-radius: 3px;
}

.btn-secondary {
  background: #ebecf0;
  border: none;
  padding: 8px 12px;
  border-radius: 3px;
  cursor: pointer;
  margin-top: 10px;
  color: #172b4d;
}

.btn-secondary:hover {
  background: #dfe1e6;
}

.btn-clear-label {
  background: #ebecf0;
  border: none;
  padding: 0 12px;
  border-radius: 3px;
  cursor: pointer;
}

/* Skeleton Animation */
.skeleton-list {
  height: 400px;
  background: #edeff0;
}

.skeleton-header {
  height: 20px;
  width: 60%;
  background: #ddd;
  margin-bottom: 20px;
  border-radius: 3px;
}

.skeleton-card {
  height: 40px;
  background: #fff;
  margin-bottom: 8px;
  border-radius: 3px;
}

.skeleton-wrapper {
  animation: pulse 1.5s infinite ease-in-out;
}

@keyframes pulse {
  0% {
    opacity: 1;
  }

  50% {
    opacity: 0.4;
  }

  100% {
    opacity: 1;
  }
}
</style>