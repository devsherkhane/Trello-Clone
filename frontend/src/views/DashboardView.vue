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
    <div v-if="pendingInvitations.length > 0" class="dashboard-section invitations-section">
      <h2 class="section-title">📬 Pending Invitations</h2>
      <div class="invitation-grid">
        <div v-for="invite in pendingInvitations" :key="invite.id" class="invitation-card">
          <div class="invite-info">
            <span class="invite-label">Board Invite</span>
            <h3 class="invite-title">{{ invite.title }}</h3>
          </div>
          <div class="invite-actions">
            <button class="btn-invite accept" @click="respondToInvite(invite.id, 'accept')">Accept</button>
            <button class="btn-invite decline" @click="respondToInvite(invite.id, 'decline')">Decline</button>
          </div>
        </div>
      </div>
    </div>

    <div class="dashboard-section">
      <h2 class="section-title">My Boards ({{ myBoards.length }})</h2>
      <div class="board-grid">
        <div class="board-tile create-tile" @click="isModalOpen = true">
          <div class="tile-icon">+</div>
          <span>Create new board</span>
        </div>

        <div v-for="board in myBoards" :key="board.id" class="board-tile" @click="openBoard(board.id)">
          <div class="board-tile-content">
            <h3 class="board-title-text">{{ board.title }}</h3>
          </div>
          <div class="board-actions">
            <button @click.stop="editBoard(board)" title="Edit Board">✏️</button>
            <button @click.stop="deleteBoard(board.id)" title="Delete Board">🗑️</button>
          </div>
        </div>
      </div>
    </div>

    <div v-if="sharedBoards.length > 0" class="dashboard-section">
      <h2 class="section-title">Shared with Me</h2>
      <div class="board-grid">
        <div v-for="board in sharedBoards" :key="board.id" class="board-tile shared-tile" @click="openBoard(board.id)">
          <div class="board-tile-content">
            <div class="shared-badge">Shared</div>
            <h3 class="board-title-text">{{ board.title }}</h3>
          </div>
          <!-- No edit/delete for shared boards -->
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
import { ref, onMounted, computed } from 'vue';
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

const myBoards = computed(() => boards.value.filter(b => b.is_owner));
const sharedBoards = computed(() => boards.value.filter(b => !b.is_owner && b.status === 'accepted'));
const pendingInvitations = computed(() => boards.value.filter(b => !b.is_owner && b.status === 'pending'));

const fetchBoards = async () => {
  try {
    const response = await api.get('/boards');
    // If response.data is null, default to an empty array []
    boards.value = response.data || []; 
  } catch (error) { 
    toast.error("Failed to fetch boards"); 
  }
};

const logout = () => {
  auth.logout();
};

const openBoard = (id) => router.push(`/board/${id}`);

const respondToInvite = async (boardId, action) => {
  try {
    await api.patch(`/boards/${boardId}/invitation`, { action });
    toast.success(`Invitation ${action}ed`);
    await fetchBoards(); // Refresh list
  } catch (err) {
    toast.error(`Failed to ${action} invitation`);
  }
};

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
.content {
  padding: 0;
  min-height: 100vh;
  background: var(--bg-gradient);
}

.board-header {
  padding: 16px 32px;
  background: var(--surface-glass);
  backdrop-filter: blur(16px);
  -webkit-backdrop-filter: blur(16px);
  border-bottom: 1px solid var(--border-subtle);
  display: flex;
  justify-content: space-between;
  align-items: center;
  position: sticky;
  top: 0;
  z-index: 10;
}

[data-theme="dark"] .board-header {
  background: rgba(15, 23, 42, 0.4);
  border-bottom: 1px solid rgba(255, 255, 255, 0.05);
}

.header-left h2 {
  color: var(--text-main);
  margin: 0;
  font-size: 24px;
  font-weight: 700;
  letter-spacing: -0.5px;
}

.btn-secondary {
  background: rgba(255, 255, 255, 0.2);
  color: var(--text-light);
  border: 1px solid rgba(255, 255, 255, 0.3);
  padding: 8px 16px;
  border-radius: var(--border-radius-sm);
  font-weight: 600;
  font-size: 14px;
  cursor: pointer;
  transition: all 0.2s ease;
}

.btn-secondary:hover {
  background: rgba(255, 255, 255, 0.3);
  transform: translateY(-1px);
}

[data-theme="dark"] .btn-secondary {
  background: rgba(255, 255, 255, 0.1);
  border: 1px solid rgba(255, 255, 255, 0.15);
}

[data-theme="dark"] .btn-secondary:hover {
  background: rgba(255, 255, 255, 0.2);
}

h2 {
  color: var(--text-main);
  margin: 40px 32px 20px;
  font-size: 28px;
  font-weight: 700;
}

.invitations-section {
  background: linear-gradient(135deg, rgba(20, 184, 166, 0.05), rgba(99, 102, 241, 0.05));
  padding: 24px 32px !important;
  border-radius: var(--border-radius);
  margin: 20px 32px 48px !important;
  border: 1px solid var(--border-subtle);
  box-shadow: var(--shadow-strong);
  backdrop-filter: blur(8px);
}

.invitation-grid {
  display: flex;
  flex-wrap: wrap;
  gap: 16px;
}

.invitation-card {
  background: var(--surface-primary);
  padding: 16px 20px;
  border-radius: 14px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 32px;
  min-width: 400px;
  border: 1px solid var(--border-subtle);
  box-shadow: var(--shadow-soft);
  transition: transform 0.2s, box-shadow 0.2s;
}

.invitation-card:hover {
  transform: translateY(-2px);
  box-shadow: var(--shadow-strong);
}

[data-theme="dark"] .invitation-card {
  background: var(--surface-primary);
  border: 1px solid var(--border-subtle);
}

.invite-info {
  display: flex;
  flex-direction: column;
}

.invite-label {
  font-size: 11px;
  text-transform: uppercase;
  font-weight: 800;
  color: var(--brand-primary);
  letter-spacing: 0.5px;
}

.invite-title {
  margin: 4px 0 0;
  font-size: 16px;
  color: var(--text-main);
}

.invite-actions {
  display: flex;
  gap: 8px;
}

.btn-invite {
  padding: 8px 16px;
  border-radius: 8px;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-invite.accept {
  background: var(--brand-primary);
  color: var(--text-on-brand);
  border: none;
}

.btn-invite.accept:hover {
  background: var(--brand-primary-hover);
  transform: translateY(-1px);
}

.btn-invite.decline {
  background: transparent;
  color: var(--text-muted);
  border: 1px solid var(--border-subtle);
}

.btn-invite.decline:hover {
  background: #fee2e2;
  color: #ef4444;
  border-color: #fca5a5;
}

.dashboard-section {
  margin-bottom: 48px;
  padding: 0 32px;
}

.section-title {
  font-size: 20px;
  font-weight: 700;
  color: var(--text-main);
  margin-bottom: 20px;
  display: flex;
  align-items: center;
  gap: 10px;
}

.board-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 24px;
}

.board-tile {
  background: var(--surface-primary);
  border: 1px solid var(--border-subtle);
  border-radius: var(--border-radius);
  padding: 24px;
  height: 160px;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  cursor: pointer;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  color: var(--text-main);
  box-shadow: var(--shadow-soft);
  position: relative;
  overflow: hidden;
  background-image: radial-gradient(circle at top right, rgba(20, 184, 166, 0.05), transparent);
}

[data-theme="dark"] .board-tile {
  background: var(--surface-primary);
  border: 1px solid var(--border-subtle);
}

.board-tile:hover {
  transform: translateY(-4px) scale(1.02);
  background: rgba(255, 255, 255, 0.25);
  box-shadow: 0 12px 20px rgba(0,0,0,0.1);
  border-color: rgba(255, 255, 255, 0.5);
}

.board-tile-content {
  flex-grow: 1;
  z-index: 1;
}

.board-title-text {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
}

.shared-tile {
  border-left: 4px solid var(--brand-primary);
}

.shared-badge {
  display: inline-block;
  font-size: 10px;
  text-transform: uppercase;
  font-weight: 800;
  color: var(--text-on-brand);
  background: var(--brand-primary);
  padding: 4px 10px;
  border-radius: 12px;
  margin-bottom: 12px;
}

.create-tile {
  background: rgba(20, 184, 166, 0.03);
  border: 2px dashed var(--brand-primary);
  align-items: center;
  justify-content: center;
  text-align: center;
  color: var(--brand-primary);
}

.tile-icon {
  font-size: 32px;
  margin-bottom: 8px;
  z-index: 1;
}

.board-actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  opacity: 0;
  transition: opacity 0.2s ease;
  z-index: 2;
}

.board-tile:hover .board-actions {
  opacity: 1;
}

.board-actions button {
  background: rgba(255, 255, 255, 0.2);
  border: none;
  cursor: pointer;
  padding: 8px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: background 0.2s;
  width: 32px;
  height: 32px;
  font-size: 14px;
}

[data-theme="dark"] .board-actions button {
  background: rgba(255, 255, 255, 0.1);
  color: var(--text-light);
}

.board-actions button:hover {
  background: var(--text-light);
}

[data-theme="dark"] .board-actions button:hover {
  background: rgba(255, 255, 255, 0.2);
}

/* Modal Styles */
.modal-backdrop {
  position: fixed;
  top: 0; left: 0; right: 0; bottom: 0;
  background: rgba(15, 23, 42, 0.6);
  backdrop-filter: blur(4px);
  -webkit-backdrop-filter: blur(4px);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 100;
  animation: fadeIn 0.2s ease-out;
}

.modal-content {
  background: var(--surface-color);
  padding: 32px;
  border-radius: var(--border-radius);
  width: 100%;
  max-width: 400px;
  box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.1), 0 10px 10px -5px rgba(0, 0, 0, 0.04);
  transform: translateY(0);
  animation: modalSlideIn 0.3s cubic-bezier(0.16, 1, 0.3, 1);
}

@keyframes modalSlideIn {
  from { opacity: 0; transform: translateY(20px) scale(0.95); }
  to { opacity: 1; transform: translateY(0) scale(1); }
}

.modal-content h3 {
  margin: 0 0 20px;
  color: var(--text-primary);
  font-size: 20px;
  font-weight: 700;
}

.modal-content input {
  width: 100%;
  padding: 12px 16px;
  border: 1px solid rgba(99, 102, 241, 0.2);
  border-radius: var(--border-radius-sm);
  background: transparent;
  color: var(--text-primary);
  font-size: 16px;
  margin-bottom: 24px;
  transition: all 0.2s ease;
}

.modal-content input:focus {
  outline: none;
  border-color: var(--primary-color);
  box-shadow: 0 0 0 3px rgba(99, 102, 241, 0.15);
}

.modal-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}

.btn-cancel {
  background: rgba(100, 116, 139, 0.1);
  color: var(--text-secondary);
  border: none;
  padding: 10px 20px;
  border-radius: var(--border-radius-sm);
  font-weight: 600;
  cursor: pointer;
  transition: background 0.2s;
}

.btn-cancel:hover {
  background: rgba(100, 116, 139, 0.2);
  color: var(--text-primary);
}

.btn-create {
  background: var(--primary-color);
  color: white;
  border: none;
  padding: 10px 20px;
  border-radius: var(--border-radius-sm);
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
  box-shadow: var(--shadow-sm);
}

.btn-create:hover:not(:disabled) {
  background: var(--primary-hover);
  transform: translateY(-1px);
  box-shadow: var(--shadow-md);
}

.btn-create:disabled {
  background: #cbd5e1;
  cursor: not-allowed;
  transform: none;
  box-shadow: none;
}
</style>