<template>
  <div class="board-canvas">
    <nav class="board-header">
      <div class="header-left">
        <h2 class="brand">Trello Clone</h2>
        <div class="divider"></div>
        <h2 class="board-title">{{ boardTitle }}</h2>
        <button class="btn-action" @click="isShareModalOpen = true">
          👥 Share
        </button>
        <button class="btn-action" @click="toggleMenu">
          ☰ Menu
        </button>
      </div>

      <div class="header-right">
        <div class="search-wrapper">
          <input 
            v-model="searchQuery" 
            @input="performSearch"
            placeholder="Search cards..." 
            class="search-input" 
          />
          
          <!-- Search Results Dropdown -->
          <div v-if="searchQuery.trim()" class="search-results-dropdown">
            <div v-if="isSearching" class="search-status">Searching...</div>
            <div v-else-if="searchResults.length === 0" class="search-status">No cards found.</div>
            <div 
              v-else 
              v-for="card in searchResults" 
              :key="card.id" 
              class="search-result-item"
              @click="openCardDetails(card, 'Search Result')"
            >
              <div class="search-card-title">{{ card.title }}</div>
              <div v-if="card.description" class="search-card-desc">
                {{ card.description.substring(0, 50) }}{{ card.description.length > 50 ? '...' : '' }}
              </div>
            </div>
          </div>
        </div>
        
        <router-link to="/profile" class="btn-icon-link" title="Profile Settings">
          👤
        </router-link>
        <button @click="logout" class="btn-logout" title="Logout">Logout</button>
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

    <div v-if="activeCard" class="modal-backdrop" @click.self="closeCardModal">
      <div class="card-detail-modal">
        <button class="modal-close" @click="closeCardModal">✕</button>

        <div class="modal-header">
          <div class="header-title">
            <h3>💳 {{ activeCard.title }}</h3>
            <p class="subtitle">in list <u>{{ activeCard.listName }}</u></p>
          </div>
          <button @click="deleteCard" class="btn-danger">Delete Card</button>
        </div>

        <div class="modal-body">
          <div class="main-column">
            <div class="modal-section">
              <h4>📝 Description</h4>
              <textarea 
                v-model="activeCard.description" 
                placeholder="Add a more detailed description..."
                @blur="updateCardDetails('description', activeCard.description)"
                class="description-input"
              ></textarea>
            </div>

            <div class="modal-section">
              <h4>📎 Attachments</h4>
              <div class="attachments-list">
                <div v-for="file in activeCard.attachments" :key="file.id" class="attachment-item">
                  <div class="file-preview">
                    <img v-if="isImage(file.filename)" :src="getAttachmentUrl(file.file_path)" alt="preview"/>
                    <span v-else class="file-icon">📄</span>
                  </div>
                  <div class="file-info">
                    <a :href="getAttachmentUrl(file.file_path)" target="_blank">{{ file.filename }}</a>
                    <span class="file-date">Added recently</span>
                  </div>
                </div>
              </div>
              <input type="file" ref="fileInput" style="display: none" @change="handleFileUpload" />
              <button class="btn-secondary" @click="$refs.fileInput.click()">Add Attachment</button>
            </div>

            <div class="modal-section comments-section">
              <h4>💬 Activity</h4>
              <div class="comment-input-area">
                <div class="comment-avatar">{{ authStore.user?.username?.charAt(0).toUpperCase() || 'U' }}</div>
                <div class="comment-box">
                  <textarea v-model="newCommentText" placeholder="Write a comment..."></textarea>
                  <button class="btn-primary" @click="submitComment" :disabled="!newCommentText.trim()">Save</button>
                </div>
              </div>
              
              <div class="comments-list">
                <div v-for="comment in activeCard.comments" :key="comment.id" class="comment-item">
                  <div class="comment-avatar">{{ comment.user_name?.charAt(0).toUpperCase() || 'U' }}</div>
                  <div class="comment-content">
                    <span class="comment-user">{{ comment.user_name }}</span>
                    <span class="comment-date">{{ formatDate(comment.created_at) }}</span>
                    <p>{{ comment.text }}</p>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <div class="sidebar-column">
            <div class="modal-section">
              <h4>📅 Due Date</h4>
              <input 
                type="date" 
                v-model="activeCard.due_date" 
                @change="updateCardDetails('due_date', activeCard.due_date)"
                class="date-picker"
              />
            </div>

            <div class="modal-section">
              <h4>🏷️ Labels</h4>
              <div class="labels-picker">
                <div 
                  v-for="color in ['#61bd4f', '#f2d600', '#ff9f1a', '#eb5a46', '#c377e0', '#0079bf']" 
                  :key="color"
                  class="label-swatch" 
                  :style="{ backgroundColor: color }"
                  :class="{ active: activeCard.label_color === color }" 
                  @click="updateCardDetails('label_color', color)"
                ></div>
                <button class="btn-clear-label" @click="updateCardDetails('label_color', null)">Clear</button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Right Sidebar Menu -->
    <div v-if="isMenuOpen" class="menu-overlay" @click="isMenuOpen = false"></div>
    <div class="board-menu" :class="{ 'menu-open': isMenuOpen }">
      <div class="menu-header">
        <h3>Menu</h3>
        <button class="btn-close" @click="isMenuOpen = false">✕</button>
      </div>
      <div class="menu-content">
        <section class="menu-section">
          <button class="btn-danger w-100" @click="archiveBoard">
            🗑️ Archive Board
          </button>
        </section>
        
        <section class="menu-section">
          <h4>Activity</h4>
          <div class="activity-feed">
            <div v-if="activityLogs.length === 0" class="empty-state">
              No recent activity.
            </div>
            <div v-for="(log, index) in activityLogs" :key="index" class="activity-item">
              <div class="activity-content">
                <p>{{ log.action }}</p>
                <span class="activity-time">{{ formatTimeAgo(log.created_at) }}</span>
              </div>
            </div>
          </div>
        </section>
      </div>
    </div>

    <!-- Share / Collaborate Modal -->
    <div v-if="isShareModalOpen" class="modal-backdrop" @click.self="isShareModalOpen = false">
      <div class="share-modal-content">
        <div class="modal-header" style="padding-right: 0;">
          <h3>Share Board</h3>
          <button class="btn-close" @click="isShareModalOpen = false">✕</button>
        </div>
        <div class="modal-body" style="flex-direction: column; gap: 15px;">
          <p class="subtitle" style="margin: 0;">Invite someone to collaborate on this board.</p>
          <div class="form-group">
            <input 
              v-model="collaboratorEmail" 
              type="email" 
              placeholder="Email address" 
              @keyup.enter="inviteCollaborator"
              class="share-input" 
            />
          </div>
          <button class="btn-primary w-100" style="padding: 12px;" @click="inviteCollaborator" :disabled="!collaboratorEmail.trim()">
            Send Invite
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import draggable from 'vuedraggable';
import api from '../api';
import { useWebsocket } from '../api/websocket';
import { useToast } from "vue-toastification";
import { useBoardStore } from '../stores/board';
import { useAuthStore } from '../stores/auth';

const toast = useToast();
const { connect, disconnect } = useWebsocket();
const route = useRoute();
const router = useRouter();

// Store Initialization
const boardStore = useBoardStore();
const authStore = useAuthStore();

// UI State
const searchQuery = ref('');
const searchResults = ref([]);
const isSearching = ref(false);
let searchTimeout = null;

const currentlyEditingList = ref(null);
const newCardTitle = ref('');
const isAddingList = ref(false);
const newListTitle = ref('');
const activeCard = ref(null);
const newCommentText = ref('');
const fileInput = ref(null);

// Menu State
const isMenuOpen = ref(false);
const activityLogs = ref([]);

// Share State
const isShareModalOpen = ref(false);
const collaboratorEmail = ref('');

// Computed properties mapping to Pinia state
const boardTitle = computed(() => boardStore.boardTitle);
const isLoading = computed(() => boardStore.isLoading);
const lists = computed(() => boardStore.lists);

// The actual lists displayed
const filteredLists = computed(() => lists.value);

// --- BACKEND ADVANCED SEARCH ---
const performSearch = () => {
  clearTimeout(searchTimeout);
  if (!searchQuery.value.trim()) {
    searchResults.value = [];
    isSearching.value = false;
    return;
  }
  
  isSearching.value = true;
  searchTimeout = setTimeout(async () => {
    try {
      // FIX: Add board_id to scope search to this board
      const response = await api.get(`/search/advanced?q=${encodeURIComponent(searchQuery.value)}&board_id=${route.params.id}`);
      searchResults.value = response.data || [];
    } catch (err) {
      console.error("Search failed:", err);
      searchResults.value = [];
    } finally {
      isSearching.value = false;
    }
  }, 300); // Debounce 300ms
};

// Logout Helper
const logout = () => {
  authStore.logout();
  router.push('/login');
};

// --- MODAL STATE MANAGEMENT ---
const closeCardModal = () => {
  activeCard.value = null;
};

// Generic function to update any card property instantly
const updateCardDetails = async (field, value) => {
  if (!activeCard.value) return;

  // Optimistic UI updates
  activeCard.value[field] = value;
  boardStore.updateCardLocally({ id: activeCard.value.id, [field]: value });

  try {
    // UNIFIED UPDATE: All fields now go through the main PUT /cards/:id endpoint
    await api.put(`/cards/${activeCard.value.id}`, {
      title: activeCard.value.title,
      description: activeCard.value.description,
      due_date: activeCard.value.due_date,
      label_color: activeCard.value.label_color
    });
    
    if (field === 'label_color') toast.success("Label updated");
  } catch (err) {
    toast.error(`Failed to update ${field}`);
    boardStore.fetchBoardDetails(route.params.id); // Revert on failure
  }
};

// --- FLAWLESS DRAG AND DROP ---
const onCardMove = async (event, newListId) => {
  const item = event.added || event.moved;
  if (!item) return;

  const card = item.element;
  const newIndex = item.newIndex;

  try {
    // FIX: Backend expects new_list_id and new_position
    await api.patch(`/cards/${card.id}/move`, {
      new_list_id: newListId,
      new_position: newIndex
    });
  } catch (err) {
    toast.error("Failed to sync card position");
    boardStore.fetchBoardDetails(route.params.id);
  }
};

// --- List Actions ---
const deleteList = async (listId) => {
  if (!confirm("Are you sure you want to delete this list and all its cards?")) return;
  try {
    await api.delete(`/lists/${listId}`);
    // Update Pinia store directly
    boardStore.lists = boardStore.lists.filter(l => l.id !== listId);
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

const submitList = async () => {
  if (!newListTitle.value.trim()) return;
  try {
    const response = await api.post(`/lists`, { board_id: parseInt(route.params.id), title: newListTitle.value.trim() });
    boardStore.lists.push({ ...response.data, cards: [] });
    newListTitle.value = '';
    isAddingList.value = false;
  } catch (err) { toast.error("List creation failed"); }
};

// --- Card Actions ---
const openAddCard = (listId) => {
  currentlyEditingList.value = listId;
  newCardTitle.value = '';
};

const submitCard = async (listId) => {
  if (!newCardTitle.value.trim()) return;
  try {
    const response = await api.post('/cards', { list_id: listId, title: newCardTitle.value.trim() });
    const targetList = boardStore.lists.find(l => l.id === listId);
    if (targetList) {
      if (!targetList.cards) targetList.cards = [];
      targetList.cards.push(response.data);
    }
    newCardTitle.value = '';
    currentlyEditingList.value = null;
  } catch (err) { toast.error("Card creation failed"); }
};

const deleteCard = async () => {
  if (!confirm("Are you sure you want to delete this card?")) return;
  try {
    await api.delete(`/cards/${activeCard.value.id}`);

    // Remove from UI (Pinia)
    const list = boardStore.lists.find(l => l.title === activeCard.value.listName);
    if (list) {
      list.cards = list.cards.filter(c => c.id !== activeCard.value.id);
    }

    activeCard.value = null;
    toast.success("Card deleted");
  } catch (err) { toast.error("Failed to delete card"); }
};

const openCardDetails = async (card, listName) => {
  // Parse existing date to YYYY-MM-DD for the HTML date input if it exists
  const parsedDate = card.due_date ? new Date(card.due_date).toISOString().split('T')[0] : null;

  activeCard.value = { 
    ...card, 
    listName,
    due_date: parsedDate,
    attachments: [],
    comments: []
  };

  // Fetch attachments and comments for this card
  try {
    const [attachmentsRes, commentsRes] = await Promise.all([
      api.get(`/cards/${card.id}/attachments`),
      api.get(`/cards/${card.id}/comments`)
    ]);
    activeCard.value.attachments = attachmentsRes.data || [];
    activeCard.value.comments = commentsRes.data || [];
  } catch (err) {
    console.error("Failed to load card secondary data", err);
  }
};

// --- Attachments ---
const getAttachmentUrl = (filePath) => {
  if (!filePath) return '';
  // Convert Windows path separators if necessary and build full URL
  return `http://localhost:8080/${filePath.replace(/\\/g, '/')}`;
};

const isImage = (filename) => {
  if (!filename) return false;
  return /\.(jpg|jpeg|png|gif|webp)$/i.test(filename);
};

const handleFileUpload = async (event) => {
  const file = event.target.files[0];
  if (!file) return;
  const formData = new FormData();
  formData.append('file', file); // Make sure backend expects 'file'
  formData.append('card_id', activeCard.value.id);
  
  try {
    const response = await api.post(`/attachments`, formData, {
      headers: { 'Content-Type': 'multipart/form-data' }
    });
    if (!activeCard.value.attachments) activeCard.value.attachments = [];
    activeCard.value.attachments.push(response.data);
    toast.success("File attached!");
  } catch (err) { 
    toast.error("Upload failed"); 
  } finally {
    event.target.value = ''; // Reset file input
  }
};

// --- Comments ---
const submitComment = async () => {
  if (!newCommentText.value.trim()) return;
  try {
    const response = await api.post(`/comments`, { card_id: activeCard.value.id, text: newCommentText.value });
    if (!activeCard.value.comments) activeCard.value.comments = [];
    // Unshift puts it at the top
    activeCard.value.comments.unshift(response.data);
    newCommentText.value = '';
  } catch (err) { toast.error("Comment failed"); }
};

const formatDate = (dateString) => {
  if (!dateString) return '';
  const date = new Date(dateString);
  return date.toLocaleString(undefined, { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' });
};

// --- Board Actions ---
const exportBoard = () => {
  window.open(`http://localhost:8080/api/boards/${route.params.id}/export?token=${localStorage.getItem('token')}`, '_blank');
};

// --- Board Menu Actions ---
const toggleMenu = async () => {
  isMenuOpen.value = !isMenuOpen.value;
  if (isMenuOpen.value) {
    try {
      const response = await api.get(`/boards/${route.params.id}/activity`);
      activityLogs.value = response.data || [];
    } catch (err) {
      toast.error("Failed to fetch activity logs");
    }
  }
};

const archiveBoard = async () => {
  if (!confirm("Are you sure you want to archive this board? It will no longer appear on your dashboard.")) return;
  try {
    await api.patch(`/boards/${route.params.id}/archive`);
    toast.success("Board archived");
    router.push('/');
  } catch (err) {
    toast.error("Failed to archive board");
  }
};

const formatTimeAgo = (dateString) => {
  if (!dateString) return '';
  const date = new Date(dateString);
  const now = new Date();
  const diffInSeconds = Math.floor((now - date) / 1000);
  
  if (diffInSeconds < 60) return `${diffInSeconds}s ago`;
  if (diffInSeconds < 3600) return `${Math.floor(diffInSeconds / 60)}m ago`;
  if (diffInSeconds < 86400) return `${Math.floor(diffInSeconds / 3600)}h ago`;
  return date.toLocaleDateString();
};

// --- Collaboration Actions ---
const inviteCollaborator = async () => {
  if (!collaboratorEmail.value.trim()) return;
  try {
    // FIX: Backend expects both email AND role
    await api.post(`/boards/${route.params.id}/collaborators`, { 
      email: collaboratorEmail.value.trim(),
      role: 'member' 
    });
    toast.success(`Invite sent to ${collaboratorEmail.value}`);
    collaboratorEmail.value = '';
    isShareModalOpen.value = false;
  } catch (err) {
    toast.error(err.response?.data?.error || "Failed to invite collaborator");
  }
};

onMounted(async () => {
  await boardStore.fetchBoardDetails(route.params.id);
  connect(route.params.id);
});

onUnmounted(() => disconnect());
</script>

<style scoped>
/* Main Layout */
.board-canvas {
  height: 100vh;
  background: var(--bg-gradient);
  display: flex;
  flex-direction: column;
}

.board-header {
  padding: 10px 20px;
  background: var(--surface-glass);
  backdrop-filter: blur(24px);
  -webkit-backdrop-filter: blur(24px);
  border-bottom: 1px solid var(--border-subtle);
  color: var(--text-main);
  display: flex;
  justify-content: space-between;
  align-items: center;
  position: sticky;
  top: 0;
  z-index: 1000;
  box-shadow: 0 4px 30px rgba(0, 0, 0, 0.05);
}

.header-left, .header-right {
  display: flex;
  align-items: center;
  gap: 12px;
}

.brand {
  font-size: 18px;
  font-weight: 800;
  letter-spacing: -0.5px;
  margin: 0;
  color: var(--brand-primary);
}

.divider {
  width: 1px;
  height: 24px;
  background: var(--border-subtle);
  margin: 0 10px;
}

.board-title {
  font-size: 16px;
  font-weight: 600;
  margin: 0;
  margin-right: 15px;
}

.btn-action {
  background: var(--surface-primary);
  border: 1px solid var(--border-subtle);
  color: var(--text-main);
  padding: 6px 12px;
  border-radius: 6px;
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
  box-shadow: var(--shadow-soft);
}

.btn-action:hover {
  background: var(--surface-secondary);
  transform: translateY(-1px);
}

.search-wrapper {
  position: relative;
  margin-right: 10px;
}

.btn-icon-link {
  font-size: 18px;
  text-decoration: none;
  background: var(--surface-primary);
  width: 34px;
  height: 34px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  transition: all 0.2s;
  color: var(--text-main);
  border: 1px solid var(--border-subtle);
}

.btn-icon-link:hover {
  background: var(--surface-secondary);
}

.btn-logout {
  background: rgba(239, 68, 68, 0.2);
  border: 1px solid rgba(239, 68, 68, 0.3);
  color: #fca5a5;
  padding: 6px 12px;
  border-radius: 6px;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-logout:hover {
  background: rgba(239, 68, 68, 0.3);
  color: white;
}

.search-input {
  background: var(--surface-primary);
  border: 1px solid var(--border-subtle);
  border-radius: var(--border-radius-sm);
  padding: 8px 16px;
  color: var(--text-main);
  width: 250px;
  font-size: 14px;
  transition: all 0.2s ease;
}

.search-input::placeholder {
  color: var(--text-extramuted);
}

.search-input:focus {
  background: var(--surface-primary);
  border-color: var(--brand-primary);
  outline: none;
  box-shadow: 0 0 0 3px rgba(99, 102, 241, 0.15);
}

.search-results-dropdown {
  position: absolute;
  top: calc(100% + 10px);
  right: 0;
  width: 320px;
  background: white;
  border-radius: 12px;
  box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.1), 0 10px 10px -5px rgba(0, 0, 0, 0.04);
  overflow-y: auto;
  z-index: 1100;
  color: var(--text-primary);
  border: 1px solid #e2e8f0;
}

[data-theme="dark"] .search-results-dropdown {
  background: var(--surface-color);
  border: 1px solid rgba(255, 255, 255, 0.1);
}

.search-status {
  padding: 16px;
  text-align: center;
  color: var(--text-secondary);
  font-size: 14px;
}

.search-result-item {
  padding: 12px 16px;
  border-bottom: 1px solid #e2e8f0;
  cursor: pointer;
  transition: background 0.2s;
}

[data-theme="dark"] .search-result-item {
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
}

.search-result-item:last-child {
  border-bottom: none;
}

.search-result-item:hover {
  background: #f8fafc;
}

[data-theme="dark"] .search-result-item:hover {
  background: rgba(255, 255, 255, 0.05);
}

.search-card-title {
  font-weight: 600;
  font-size: 14px;
  margin-bottom: 4px;
}

.search-card-desc {
  font-size: 12px;
  color: var(--text-secondary);
}

.lists-container {
  display: flex;
  align-items: flex-start;
  padding: 24px;
  gap: 16px;
  overflow-x: auto;
  flex-grow: 1;
  /* Scrollbar styling */
  scrollbar-width: thin;
  scrollbar-color: rgba(255,255,255,0.3) transparent;
}

.lists-container::-webkit-scrollbar {
  height: 12px;
}
.lists-container::-webkit-scrollbar-track {
  background: rgba(255, 255, 255, 0.1);
  border-radius: 6px;
}
.lists-container::-webkit-scrollbar-thumb {
  background-color: rgba(255, 255, 255, 0.3);
  border-radius: 6px;
  border: 3px solid transparent;
  background-clip: padding-box;
}

.list-wrapper {
  width: 300px;
  flex-shrink: 0;
}

.list-content {
  background: var(--trello-list);
  backdrop-filter: blur(12px);
  -webkit-backdrop-filter: blur(12px);
  border: 1px solid var(--border-subtle);
  border-radius: var(--border-radius);
  padding: 12px;
  display: flex;
  flex-direction: column;
  max-height: calc(100vh - 120px);
  box-shadow: var(--shadow-soft);
}

/* Card Styling */
.card {
  background: var(--trello-card);
  padding: 12px;
  margin-bottom: 8px;
  border-radius: var(--border-radius-sm);
  box-shadow: var(--card-shadow);
  cursor: grab;
  font-size: 14px;
  position: relative;
  transition: transform 0.2s cubic-bezier(0.175, 0.885, 0.32, 1.275), box-shadow 0.2s ease;
  border: 1px solid var(--border-subtle);
  color: var(--text-main);
}

.card:hover {
  transform: scale(1.02) translateY(-2px);
  box-shadow: 0 8px 16px rgba(15, 23, 42, 0.1);
  border-color: var(--brand-primary);
}

.card:hover {
  transform: translateY(-2px);
  box-shadow: var(--shadow-md);
}

.card:active {
  cursor: grabbing;
}

.card-label-bar {
  height: 6px;
  width: 48px;
  border-radius: 3px;
  margin-bottom: 8px;
}

.btn-add {
  background: var(--brand-primary);
  color: var(--text-on-brand);
  border: none;
  padding: 8px 16px;
  border-radius: var(--border-radius-sm);
  font-weight: 600;
  cursor: pointer;
  transition: background 0.2s;
}

.btn-add:hover {
  background: var(--brand-primary-hover);
}

.btn-close {
  background: none;
  border: none;
  font-size: 20px;
  cursor: pointer;
  color: var(--text-secondary);
  padding: 4px 8px;
  border-radius: var(--border-radius-sm);
}

.btn-close:hover {
  background: var(--surface-secondary);
  color: var(--text-main);
}

.add-card-btn, .add-list-placeholder {
  background: transparent;
  border: none;
  padding: 10px 12px;
  border-radius: var(--border-radius-sm);
  color: var(--text-secondary);
  font-weight: 500;
  cursor: pointer;
  text-align: left;
  transition: all 0.2s ease;
  width: 100%;
}

.add-list-placeholder {
  background: rgba(255, 255, 255, 0.2);
  color: white;
  backdrop-filter: blur(4px);
}

.add-card-btn:hover {
  background: var(--surface-secondary);
  color: var(--text-main);
}

.add-list-placeholder:hover {
  background: var(--surface-primary);
  color: var(--brand-primary);
  border: 1px solid var(--brand-primary);
}

.add-card-form textarea, .add-list-form input {
  width: 100%;
  border: 1px solid rgba(99, 102, 241, 0.2);
  border-radius: var(--border-radius-sm);
  padding: 12px;
  margin-bottom: 8px;
  box-sizing: border-box;
  font-family: inherit;
  font-size: 14px;
  resize: vertical;
  min-height: 40px;
}

.add-card-form textarea:focus, .add-list-form input:focus {
  outline: none;
  border-color: var(--brand-primary);
  box-shadow: 0 0 0 3px rgba(99, 102, 241, 0.15);
}

/* --- ENHANCED MODAL STYLES --- */
.modal-backdrop {
  position: fixed;
  top: 0; left: 0; width: 100%; height: 100%;
  background: rgba(15, 23, 42, 0.6);
  backdrop-filter: blur(4px);
  -webkit-backdrop-filter: blur(4px);
  display: flex;
  justify-content: center;
  align-items: flex-start;
  z-index: 100;
  overflow-y: auto;
  padding: 50px 0;
  animation: fadeIn 0.2s ease-out;
}

.card-detail-modal {
  background: var(--surface-primary);
  width: 768px;
  padding: 32px;
  border-radius: 20px;
  position: relative;
  box-shadow: var(--shadow-strong);
  border: 1px solid var(--border-subtle);
  transform: translateY(0);
  animation: modalSlideIn 0.4s cubic-bezier(0.16, 1, 0.3, 1);
}

@keyframes modalSlideIn {
  from { opacity: 0; transform: translateY(20px) scale(0.95); }
  to { opacity: 1; transform: translateY(0) scale(1); }
}

.modal-close {
  position: absolute;
  top: 20px;
  right: 20px;
  background: #f1f5f9;
  border: none;
  font-size: 18px;
  width: 36px;
  height: 36px;
  border-radius: 50%;
  cursor: pointer;
  color: var(--text-secondary);
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s;
}

.modal-close:hover {
  background: #e2e8f0;
  color: var(--text-primary);
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 24px;
  padding-right: 40px;
}

.modal-header h3 {
  margin: 0;
  font-size: 24px;
  font-weight: 700;
  color: var(--text-main);
  letter-spacing: -0.5px;
}

.modal-header .subtitle {
  margin: 4px 0 0 32px;
  color: var(--text-muted);
  font-size: 14px;
}

.btn-danger {
  background: #fef2f2;
  color: #ef4444;
  border: 1px solid #fee2e2;
  padding: 8px 16px;
  border-radius: var(--border-radius-sm);
  cursor: pointer;
  font-weight: 600;
  transition: all 0.2s;
}

.btn-danger:hover {
  background: #fee2e2;
  border-color: #fca5a5;
}

.btn-primary {
  background: var(--brand-primary);
  color: var(--text-on-brand);
  border: none;
  padding: 8px 16px;
  border-radius: var(--border-radius-sm);
  cursor: pointer;
  font-weight: 600;
  transition: all 0.2s;
}

.btn-primary:hover {
  background: var(--brand-primary-hover);
  transform: translateY(-1px);
}

.btn-primary:disabled {
  background: #cbd5e1;
  cursor: not-allowed;
  transform: none;
}

.modal-body {
  display: flex;
  gap: 32px;
}

.main-column {
  flex: 3;
}

.sidebar-column {
  flex: 1;
  min-width: 180px;
}

.modal-section {
  margin-bottom: 32px;
}

.modal-section h4 {
  margin: 0 0 16px 0;
  color: var(--text-primary);
  font-size: 16px;
  font-weight: 600;
  display: flex;
  align-items: center;
  gap: 8px;
}

/* Description */
.description-input {
  width: 100%;
  min-height: 120px;
  padding: 16px;
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: var(--border-radius-sm);
  resize: vertical;
  font-family: inherit;
  font-size: 14px;
  box-sizing: border-box;
  transition: all 0.2s;
}

.description-input:focus {
  background: #fff;
  border-color: var(--primary-color);
  box-shadow: 0 0 0 3px rgba(99, 102, 241, 0.15);
  outline: none;
}

/* Attachments */
.attachment-item {
  display: flex;
  gap: 16px;
  margin-bottom: 16px;
  padding: 16px;
  border: 1px solid #e2e8f0;
  border-radius: var(--border-radius-sm);
  background: #f8fafc;
  transition: all 0.2s;
}

.attachment-item:hover {
  border-color: #cbd5e1;
  background: #fff;
}

.file-preview {
  width: 80px;
  height: 60px;
  background: #e2e8f0;
  border-radius: var(--border-radius-sm);
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
}

.file-preview img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.file-icon {
  font-size: 24px;
}

.file-info {
  display: flex;
  flex-direction: column;
  justify-content: center;
  gap: 4px;
}

.file-info a {
  color: var(--text-primary);
  font-weight: 600;
  font-size: 14px;
  text-decoration: none;
  word-break: break-all;
}

.file-info a:hover {
  color: var(--primary-color);
  text-decoration: underline;
}

.file-date {
  font-size: 12px;
  color: var(--text-secondary);
}

/* Comments */
.comment-input-area {
  display: flex;
  gap: 16px;
  margin-bottom: 24px;
}

.comment-avatar {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  background: var(--bg-gradient);
  color: white;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 600;
  font-size: 14px;
  flex-shrink: 0;
  box-shadow: var(--shadow-sm);
}

.comment-box {
  flex-grow: 1;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.comment-box textarea {
  width: 100%;
  height: 80px;
  padding: 12px;
  border: 1px solid #e2e8f0;
  border-radius: var(--border-radius-sm);
  background: #f8fafc;
  font-family: inherit;
  font-size: 14px;
  resize: vertical;
  box-sizing: border-box;
  transition: all 0.2s;
}

.comment-box textarea:focus {
  background: #fff;
  border-color: var(--primary-color);
  box-shadow: 0 0 0 3px rgba(99, 102, 241, 0.15);
  outline: none;
}

.comment-box .btn-primary {
  align-self: flex-start;
}

.comment-item {
  display: flex;
  gap: 16px;
  margin-bottom: 20px;
}

.comment-content {
  background: #fff;
  border: 1px solid #e2e8f0;
  padding: 16px;
  border-radius: var(--border-radius-sm);
  width: 100%;
}

.comment-header {
  display: flex;
  align-items: baseline;
  gap: 12px;
  margin-bottom: 8px;
}

.comment-user {
  font-weight: 600;
  color: var(--text-primary);
}

.comment-date {
  font-size: 12px;
  color: var(--text-secondary);
}

.comment-content p {
  margin: 0;
  color: var(--text-primary);
  font-size: 14px;
  line-height: 1.5;
  white-space: pre-wrap;
}

/* Sidebar Elements (Date & Labels) */
.date-picker {
  width: 100%;
  padding: 10px;
  border: 1px solid #e2e8f0;
  border-radius: var(--border-radius-sm);
  background: #f8fafc;
  font-family: inherit;
  color: var(--text-primary);
  box-sizing: border-box;
  transition: all 0.2s;
}

.date-picker:focus {
  outline: none;
  border-color: var(--primary-color);
  box-shadow: 0 0 0 3px rgba(99, 102, 241, 0.15);
  background: #fff;
}

.labels-picker {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
  margin-top: 12px;
}

.label-swatch {
  width: 44px;
  height: 32px;
  border-radius: var(--border-radius-sm);
  cursor: pointer;
  transition: transform 0.2s;
}

.label-swatch:hover {
  transform: scale(1.05);
}

.label-swatch.active {
  border: 2px solid var(--text-primary);
  box-shadow: 0 0 0 2px white inset;
}

.btn-clear-label {
  background: #f1f5f9;
  border: none;
  padding: 8px 12px;
  border-radius: var(--border-radius-sm);
  cursor: pointer;
  margin-top: 8px;
  width: 100%;
  text-align: center;
  font-weight: 600;
  color: var(--text-secondary);
  transition: all 0.2s;
}

.btn-clear-label:hover {
  background: #e2e8f0;
  color: var(--text-primary);
}

.btn-secondary {
  background: rgba(255, 255, 255, 0.2);
  border: 1px solid rgba(255, 255, 255, 0.3);
  padding: 8px 16px;
  border-radius: var(--border-radius-sm);
  cursor: pointer;
  color: white;
  font-weight: 600;
  transition: all 0.2s;
}

.btn-secondary:hover {
  background: rgba(255, 255, 255, 0.3);
  transform: translateY(-1px);
}

/* Skeletons */
.skeleton-list {
  height: 400px;
  background: rgba(248, 250, 252, 0.5);
  border: 1px solid rgba(255, 255, 255, 0.2);
}

.skeleton-header {
  height: 24px;
  width: 60%;
  background: rgba(0,0,0,0.05);
  margin-bottom: 24px;
  border-radius: 4px;
}

.skeleton-card {
  height: 48px;
  background: rgba(255,255,255,0.8);
  margin-bottom: 12px;
  border-radius: var(--border-radius-sm);
  box-shadow: 0 1px 2px rgba(0,0,0,0.02);
}

.skeleton-wrapper {
  animation: pulse 2s infinite ease-in-out;
}

@keyframes pulse {
  0% { opacity: 0.8; }
  50% { opacity: 0.5; }
  100% { opacity: 0.8; }
}

/* --- BOARD MENU SIDEBAR --- */
.menu-overlay {
  position: fixed;
  top: 0; left: 0; right: 0; bottom: 0;
  background: rgba(15, 23, 42, 0.3);
  z-index: 80;
}

.board-menu {
  position: fixed;
  top: 0; right: 0;
  width: 320px;
  height: 100vh;
  background: white;
  box-shadow: -4px 0 15px rgba(0,0,0,0.1);
  z-index: 90;
  transform: translateX(100%);
  transition: transform 0.3s cubic-bezier(0.16, 1, 0.3, 1);
  display: flex;
  flex-direction: column;
}

.board-menu.menu-open {
  transform: translateX(0);
}

.menu-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px;
  border-bottom: 1px solid #e2e8f0;
}

.menu-header h3 {
  margin: 0;
  color: var(--text-primary);
}

.menu-content {
  flex: 1;
  overflow-y: auto;
  padding: 20px;
  background: #f8fafc;
}

.menu-section {
  margin-bottom: 30px;
}

.menu-section h4 {
  margin: 0 0 15px 0;
  color: var(--text-secondary);
  text-transform: uppercase;
  font-size: 13px;
  letter-spacing: 0.5px;
}

.w-100 {
  width: 100%;
}

.activity-feed {
  display: flex;
  flex-direction: column;
  gap: 15px;
}

.activity-item {
  display: flex;
  gap: 12px;
  padding-bottom: 15px;
  border-bottom: 1px solid #e2e8f0;
}

.activity-content p {
  margin: 0 0 4px 0;
  font-size: 14px;
  color: var(--text-primary);
  line-height: 1.4;
}

.activity-time {
  font-size: 12px;
  color: var(--text-secondary);
}

.empty-state {
  text-align: center;
  color: var(--text-secondary);
  font-size: 14px;
  padding: 20px 0;
}
/* Share Modal */
.share-modal-content {
  background: white;
  width: 400px;
  padding: 24px;
  border-radius: var(--border-radius);
  box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.1);
  animation: modalSlideIn 0.2s ease-out;
}

.share-input {
  width: 100%;
  padding: 12px;
  border: 1px solid #e2e8f0;
  border-radius: var(--border-radius-sm);
  font-family: inherit;
  font-size: 14px;
  box-sizing: border-box;
  transition: all 0.2s;
}

.share-input:focus {
  outline: none;
  border-color: var(--primary-color);
  box-shadow: 0 0 0 3px rgba(99, 102, 241, 0.15);
}
</style>