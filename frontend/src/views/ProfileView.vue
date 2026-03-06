<template>
  <div class="profile-canvas">
    <nav class="board-header">
      <div class="header-left">
        <h2>User Settings</h2>
      </div>
      <div class="header-right">
        <button class="btn-secondary" @click="goBack">Back to Dashboard</button>
      </div>
    </nav>

    <div class="profile-content">
      <div class="settings-card">
        
        <section class="settings-section">
          <h3>Profile Picture</h3>
          <div class="avatar-upload-area">
            <div class="avatar-preview">
              <span v-if="!avatarUrl">{{ auth.user?.username?.charAt(0).toUpperCase() || 'U' }}</span>
              <img v-else :src="avatarUrl" alt="Avatar" />
            </div>
            <div class="upload-controls">
              <input type="file" ref="fileInput" accept="image/*" style="display: none" @change="onFileSelected" />
              <button class="btn-secondary" @click="$refs.fileInput.click()">Choose File</button>
              <button class="btn-primary" :disabled="!selectedFile" @click="uploadAvatar">Upload</button>
              <p class="file-name" v-if="selectedFile">{{ selectedFile.name }}</p>
            </div>
          </div>
        </section>

        <section class="settings-section">
          <h3>Update Profile</h3>
          <form @submit.prevent="updateProfile" class="profile-form">
            <div class="form-group">
              <label>Email</label>
              <input type="email" :value="auth.user?.email" disabled class="disabled-input" />
              <small class="hint">Email cannot be changed</small>
            </div>
            <div class="form-group">
              <label>Username</label>
              <input type="text" v-model="profileForm.username" placeholder="New username" required />
            </div>
            <div class="form-group">
              <label>Current Password</label>
              <input type="password" v-model="profileForm.current_password" placeholder="Verify current password" required />
            </div>
            <div class="form-group">
              <label>New Password (Optional)</label>
              <input type="password" v-model="profileForm.new_password" placeholder="Leave blank to keep current" />
            </div>
            <button type="submit" class="btn-primary">Save Changes</button>
          </form>
        </section>

        <!-- Theme Toggling Section -->
        <section class="settings-section">
          <h3>Display Theme</h3>
          <p class="subtitle">Customize the look and feel of your Trello Clone workspace.</p>
          
          <div class="theme-options">
            <button 
              class="theme-btn" 
              :class="{ active: currentTheme === 'light' }" 
              @click="setTheme('light')"
            >
              ☀️ Light Mode
            </button>
            <button 
              class="theme-btn" 
              :class="{ active: currentTheme === 'dark' }" 
              @click="setTheme('dark')"
            >
              🌙 Dark Mode
            </button>
          </div>
        </section>


      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue';
import { useRouter } from 'vue-router';
import { useAuthStore } from '../stores/auth';
import { useToast } from 'vue-toastification';
import api from '../api';

const router = useRouter();
const auth = useAuthStore();
const toast = useToast();

// State
const selectedFile = ref(null);
const fileInput = ref(null);

const avatarUrl = computed(() => {
  if (!auth.user?.avatar_url) return '';
  if (auth.user.avatar_url.startsWith('http')) return auth.user.avatar_url;
  return `http://localhost:8080/${auth.user.avatar_url.replace(/\\/g, '/')}`;
});

const profileForm = ref({
  username: '',
  current_password: '',
  new_password: ''
});

const currentTheme = ref('light');

// Initialization
onMounted(async () => {
  if (!auth.user) {
    await auth.fetchUser();
  }
  
  if (auth.user) {
    profileForm.value.username = auth.user.username || '';
    
    // Load existing theme
    const savedTheme = localStorage.getItem('theme') || auth.user.theme || 'light';
    currentTheme.value = savedTheme;
    document.documentElement.setAttribute('data-theme', savedTheme);
  }
});

const goBack = () => {
  router.push('/');
};

// Avatar Handling
const onFileSelected = (event) => {
  const file = event.target.files[0];
  if (file) {
    selectedFile.value = file;
  }
};

const uploadAvatar = async () => {
  if (!selectedFile.value) return;
  
  const formData = new FormData();
  formData.append('avatar', selectedFile.value); // Backend expects 'avatar'

  try {
    const response = await api.post('/user/avatar', formData, {
      headers: { 'Content-Type': 'multipart/form-data' }
    });
    toast.success("Profile picture updated!");
    selectedFile.value = null;
    
    // Update local state by re-fetching user data
    await auth.fetchUser();
  } catch (error) {
    toast.error(error.response?.data?.error || "Failed to upload avatar");
  }
};

// Profile Update Handling
const updateProfile = async () => {
  try {
    await api.put('/profile', {
      username: profileForm.value.username,
      current_password: profileForm.value.current_password,
      new_password: profileForm.value.new_password
    });
    
    toast.success("Profile updated successfully");
    
    // Clear passwords after successful update
    profileForm.value.current_password = '';
    profileForm.value.new_password = '';
    
    // Refresh user data in store
    await auth.fetchUser();
  } catch (error) {
    toast.error(error.response?.data?.error || "Failed to update profile");
  }
};

// Theme Toggle Logic
const setTheme = async (theme) => {
  currentTheme.value = theme;
  document.documentElement.setAttribute('data-theme', theme);
  localStorage.setItem('theme', theme);

  try {
    await api.put('/user/theme', { theme });
    toast.success(`Theme updated to ${theme} mode`);
  } catch (err) {
    toast.error("Failed to save theme preference");
  }
};
</script>

<style scoped>
.profile-canvas {
  min-height: 100vh;
  background: var(--bg-gradient);
  display: flex;
  flex-direction: column;
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

.header-left h2 {
  color: var(--text-main);
  margin: 0;
  font-size: 24px;
  font-weight: 700;
  letter-spacing: -0.5px;
}

.profile-content {
  flex-grow: 1;
  display: flex;
  justify-content: center;
  align-items: flex-start;
  padding: 40px 20px;
  overflow-y: auto;
}

.settings-card {
  background: var(--surface-glass);
  backdrop-filter: blur(24px);
  -webkit-backdrop-filter: blur(24px);
  border: 1px solid var(--border-subtle);
  border-radius: 20px;
  width: 100%;
  max-width: 640px;
  padding: 40px;
  box-shadow: var(--shadow-strong);
  animation: fadeIn 0.5s cubic-bezier(0.16, 1, 0.3, 1) forwards;
}

.settings-section {
  background: var(--surface-secondary);
  border-radius: var(--border-radius-sm);
  padding: 24px;
  margin-bottom: 24px;
  border: 1px solid var(--border-subtle);
  box-shadow: var(--shadow-soft);
  transition: transform 0.2s, box-shadow 0.2s;
}

.settings-section:hover {
  transform: translateY(-2px);
  box-shadow: var(--shadow-strong);
  border-color: var(--brand-primary);
}

.settings-section h3 {
  margin-top: 0;
  margin-bottom: 20px;
  color: var(--text-main);
  border-bottom: 1px solid var(--border-subtle);
  padding-bottom: 12px;
  font-size: 18px;
  font-weight: 700;
}

.subtitle {
  color: var(--text-muted);
  font-size: 14px;
  margin-bottom: 20px;
  line-height: 1.5;
}

/* Avatar Upload Styles */
.avatar-upload-area {
  display: flex;
  align-items: center;
  gap: 24px;
}

.avatar-preview {
  width: 96px;
  height: 96px;
  border-radius: 50%;
  background: var(--brand-primary);
  display: flex;
  justify-content: center;
  align-items: center;
  font-size: 36px;
  font-weight: 700;
  color: var(--text-on-brand);
  overflow: hidden;
  box-shadow: var(--shadow-soft);
  flex-shrink: 0;
}

.avatar-preview img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.upload-controls {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.upload-actions {
  display: flex;
  gap: 12px;
}

.file-name {
  font-size: 13px;
  color: var(--text-muted);
  margin: 0;
  background: var(--surface-primary);
  padding: 6px 12px;
  border-radius: 4px;
  display: inline-block;
  border: 1px solid var(--border-subtle);
}

/* Form Styles */
.profile-form {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.form-group label {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-muted);
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.form-group input {
  padding: 12px 16px;
  border: 1px solid var(--border-subtle);
  border-radius: var(--border-radius-sm);
  font-size: 15px;
  background: var(--surface-primary);
  color: var(--text-main);
  transition: all 0.2s;
}

.form-group input:focus {
  border-color: var(--brand-primary);
  outline: none;
  background: var(--surface-primary);
  box-shadow: 0 0 0 3px rgba(99, 102, 241, 0.15);
}

.form-group input:disabled {
  background: var(--surface-secondary);
  color: var(--text-extramuted);
  cursor: not-allowed;
  border-color: var(--border-subtle);
}

.hint {
  font-size: 11px;
  color: var(--text-extramuted);
  margin-top: -12px;
}

.form-group input::placeholder {
  color: var(--text-extramuted);
}


/* Buttons */
.btn-primary {
  background: var(--brand-primary);
  color: var(--text-on-brand);
  border: none;
  padding: 10px 20px;
  border-radius: var(--border-radius-sm);
  cursor: pointer;
  font-weight: 600;
  transition: all 0.2s;
  align-self: flex-start;
}

.btn-primary:hover:not(:disabled) { 
  background: var(--brand-primary-hover);
  transform: translateY(-1px);
  box-shadow: var(--shadow-soft);
}

.btn-primary:disabled {
  background: var(--text-extramuted);
  cursor: not-allowed;
  transform: none;
  box-shadow: none;
}

.btn-secondary {
  background: var(--surface-secondary);
  color: var(--text-muted);
  border: 1px solid var(--border-subtle);
  padding: 10px 20px;
  border-radius: var(--border-radius-sm);
  cursor: pointer;
  font-weight: 600;
  transition: all 0.2s;
}

.btn-secondary:hover { 
  background: var(--surface-primary);
  transform: translateY(-1px);
  color: var(--text-main);
  border-color: var(--brand-primary);
}

.settings-section .btn-secondary {
  background: var(--surface-primary);
  color: var(--text-muted);
  border: 1px solid var(--border-subtle);
}

.settings-section .btn-secondary:hover { 
  background: var(--surface-secondary);
  color: var(--text-main);
}

/* Theme Option Styles */
.theme-options {
  display: flex;
  gap: 16px;
}

.theme-btn {
  flex: 1;
  padding: 16px;
  background: var(--surface-primary);
  border: 2px solid var(--border-subtle);
  border-radius: var(--border-radius-sm);
  font-size: 16px;
  font-weight: 600;
  color: var(--text-muted);
  cursor: pointer;
  transition: all 0.2s ease;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
}

.theme-btn:hover {
  background: var(--surface-secondary);
  border-color: var(--brand-primary);
  color: var(--text-main);
}

.theme-btn.active {
  background: rgba(99, 102, 241, 0.1);
  border-color: var(--brand-primary);
  color: var(--brand-primary);
}
</style>