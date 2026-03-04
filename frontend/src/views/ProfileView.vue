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

        <section class="settings-section">
          <h3>Two-Factor Authentication (2FA)</h3>
          <p class="subtitle">Secure your account by requiring a code from an authenticator app when you log in.</p>
          
          <div v-if="!qrCodeUrl" class="two-fa-action">
            <button class="btn-primary" @click="setup2FA">Setup 2FA</button>
          </div>
          
          <div v-else class="qr-code-container">
            <p class="success-text">Scan this QR code with Google Authenticator or Authy:</p>
            <img :src="qrCodeUrl" alt="2FA QR Code" class="qr-image"/>
            <p class="warning-text">Make sure to save this setup in your app before leaving this page!</p>
          </div>
        </section>

      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import { useAuthStore } from '../stores/auth';
import { useToast } from 'vue-toastification';
import api from '../api';

const router = useRouter();
const auth = useAuthStore();
const toast = useToast();

// State
const avatarUrl = ref('');
const selectedFile = ref(null);
const fileInput = ref(null);
const qrCodeUrl = ref('');

const profileForm = ref({
  username: '',
  current_password: '',
  new_password: ''
});

// Initialization
onMounted(() => {
  if (auth.user) {
    profileForm.value.username = auth.user.username || '';
    // If your backend returns the avatar URL in the user object, set it here
    avatarUrl.value = auth.user.avatar_url || '';
  }
});

const goBack = () => {
  router.push('/dashboard');
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
  formData.append('file', selectedFile.value); // Backend usually expects 'file' or 'avatar'

  try {
    const response = await api.post('/user/avatar', formData, {
      headers: { 'Content-Type': 'multipart/form-data' }
    });
    toast.success("Profile picture updated!");
    selectedFile.value = null;
    
    // Update local state if backend returns the new URL
    if (response.data && response.data.avatar_url) {
      avatarUrl.value = response.data.avatar_url;
    }
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
    
    // Update the username in Pinia store
    if (auth.user) {
      auth.user.username = profileForm.value.username;
    }
  } catch (error) {
    toast.error(error.response?.data?.error || "Failed to update profile");
  }
};

// 2FA Handling
const setup2FA = async () => {
  try {
    const response = await api.post('/2fa/setup');
    // Adjust '.qr_url' based on the exact JSON key your backend returns
    qrCodeUrl.value = response.data.qr_url || response.data.url || response.data; 
    toast.success("2FA generated successfully");
  } catch (error) {
    toast.error("Failed to generate 2FA setup");
  }
};
</script>

<style scoped>
.profile-canvas {
  height: 100vh;
  background-color: var(--trello-blue, #0079bf);
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

.profile-content {
  flex-grow: 1;
  display: flex;
  justify-content: center;
  align-items: flex-start;
  padding: 40px 20px;
  overflow-y: auto;
}

.settings-card {
  background: #f4f5f7;
  border-radius: 3px;
  width: 100%;
  max-width: 600px;
  padding: 30px;
  box-shadow: 0 1px 3px rgba(0,0,0,0.3);
}

.settings-section {
  background: white;
  border-radius: 3px;
  padding: 20px;
  margin-bottom: 20px;
  border: 1px solid #dfe1e6;
}

.settings-section h3 {
  margin-top: 0;
  margin-bottom: 15px;
  color: #172b4d;
  border-bottom: 1px solid #dfe1e6;
  padding-bottom: 10px;
}

.subtitle {
  color: #5e6c84;
  font-size: 14px;
  margin-bottom: 15px;
}

/* Avatar Upload Styles */
.avatar-upload-area {
  display: flex;
  align-items: center;
  gap: 20px;
}

.avatar-preview {
  width: 80px;
  height: 80px;
  border-radius: 50%;
  background: #dfe1e6;
  display: flex;
  justify-content: center;
  align-items: center;
  font-size: 32px;
  font-weight: bold;
  color: #172b4d;
  overflow: hidden;
}

.avatar-preview img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.upload-controls {
  display: flex;
  align-items: center;
  gap: 10px;
}

.file-name {
  font-size: 12px;
  color: #5e6c84;
  margin: 0;
}

/* Form Styles */
.profile-form {
  display: flex;
  flex-direction: column;
  gap: 15px;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 5px;
}

.form-group label {
  font-size: 14px;
  font-weight: 600;
  color: #172b4d;
}

.form-group input {
  padding: 8px 12px;
  border: 2px solid #dfe1e6;
  border-radius: 3px;
  font-size: 14px;
  background: #fafbfc;
}

.form-group input:focus {
  border-color: #4c9aff;
  outline: none;
  background: white;
}

/* 2FA Styles */
.qr-code-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  background: #fafbfc;
  padding: 20px;
  border-radius: 3px;
  border: 1px solid #dfe1e6;
}

.qr-image {
  max-width: 200px;
  margin: 15px 0;
  border: 10px solid white;
  border-radius: 5px;
}

.success-text {
  color: #006644;
  font-weight: bold;
  margin: 0;
}

.warning-text {
  color: #ff991f;
  font-size: 12px;
  margin: 0;
}

/* Buttons */
.btn-primary {
  background: #0079bf;
  color: white;
  border: none;
  padding: 8px 16px;
  border-radius: 3px;
  cursor: pointer;
  font-weight: bold;
}
.btn-primary:hover { background: #026aa7; }
.btn-primary:disabled {
  background: #ebecf0;
  color: #a5adba;
  cursor: not-allowed;
}

.btn-secondary {
  background: rgba(255, 255, 255, 0.2);
  color: white;
  border: none;
  padding: 8px 16px;
  border-radius: 3px;
  cursor: pointer;
}
.btn-secondary:hover { background: rgba(255, 255, 255, 0.3); }

.settings-section .btn-secondary {
  background: #ebecf0;
  color: #172b4d;
}
.settings-section .btn-secondary:hover { background: #dfe1e6; }
</style>