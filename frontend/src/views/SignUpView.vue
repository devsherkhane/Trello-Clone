<template>
  <div class="auth-wrapper">
    <div class="auth-card">
      <div class="auth-header">
        <h1>Trello Clone</h1>
        <p>Create your account</p>
      </div>
      
      <form @submit.prevent="handleSignup" class="auth-form">
        <div class="form-group">
          <label>FULL NAME</label>
          <input v-model="name" type="text" placeholder="e.g. Dev Sherkhane" required />
        </div>

        <div class="form-group">
          <label>EMAIL</label>
          <input v-model="email" type="email" placeholder="Enter email" required />
        </div>
        
        <div class="form-group">
          <label>PASSWORD</label>
          <input v-model="password" type="password" placeholder="Create password" required />
        </div>

        <button type="submit" :disabled="auth.loading" class="btn-primary">
          {{ auth.loading ? 'Creating account...' : 'Sign Up' }}
        </button>
      </form>
      
      <div class="auth-footer">
        <p>Already have an account? <router-link to="/login">Log In</router-link></p>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue';
import { useAuthStore } from '../stores/auth';
import { useRouter } from 'vue-router';
import { useToast } from "vue-toastification";

const auth = useAuthStore();
const router = useRouter();
const toast = useToast();

const name = ref('');
const email = ref('');
const password = ref('');

const handleSignup = async () => {
  const result = await auth.register({ 
    username: name.value, 
    email: email.value, 
    password: password.value 
  });
  
  if (result.success) {
    toast.success("Welcome to Trello Clone!");
    router.push('/');
  } else {
    toast.error(result.message);
  }
};
</script>

<style scoped>
.auth-wrapper {
    display: flex;
    justify-content: center;
    align-items: center;
    min-height: 100vh;
    background: linear-gradient(135deg, #7c3aed, #6d28d9 50%, #4c1d95);
    padding: 20px;
}

.auth-card {
    background: #ffffff;
    padding: 60px 40px;
    border-radius: 24px;
    box-shadow: 0 10px 40px rgba(0, 0, 0, 0.04);
    width: 100%;
    max-width: 440px;
    border: 1px solid #f1f5f9;
}

.auth-header {
    text-align: center;
    margin-bottom: 32px;
}

.auth-header h1 {
    color: #4c1d95; /* Violet 900 */
    font-size: 38px;
    font-weight: 900;
    letter-spacing: -1.5px;
    margin-bottom: 8px;
}

.auth-header p {
  color: #64748b;
  font-size: 16px;
  font-weight: 500;
}

.form-group {
    margin-bottom: 24px;
}

.form-group label {
    display: block;
    font-size: 12px;
    font-weight: 700;
    color: #64748b;
    margin-bottom: 12px;
    text-transform: uppercase;
    letter-spacing: 1.5px;
}

input {
    width: 100%;
    padding: 14px 20px;
    border: 1px solid #f1f5f9;
    border-radius: 12px;
    background-color: #fafafa;
    font-size: 15px;
    color: #0f172a;
    transition: all 0.2s ease;
}

input:focus {
    border-color: #7c3aed;
    outline: none;
    box-shadow: 0 0 0 4px rgba(124, 58, 237, 0.12);
    background: white;
}

.btn-primary {
    width: 100%;
    margin-top: 30px;
    padding: 16px;
    font-size: 16px;
    background: linear-gradient(135deg, #7c3aed, #6d28d9);
    color: white;
    border: none;
    border-radius: 12px;
    font-weight: 700;
    cursor: pointer;
    box-shadow: 0 4px 14px rgba(124, 58, 237, 0.35);
    transition: all 0.2s;
}

.btn-primary:hover {
    background: linear-gradient(135deg, #6d28d9, #5b21b6);
    transform: translateY(-1px);
}

.auth-footer {
  margin-top: 32px;
  text-align: center;
  font-size: 14px;
  color: #64748b;
}

.auth-footer a {
  color: #7c3aed;
  font-weight: 700;
  text-decoration: none;
  transition: all 0.2s;
}

.auth-footer a:hover {
  color: #6d28d9;
  text-decoration: underline;
}
</style>