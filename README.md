#  Trello Clone 

An intuitive, real-time project management tool built with a high-performance **Go (Gin)** backend and a reactive **Vue 3** frontend. It features a modern  UI with professional glassmorphism, seamless drag-and-drop task organization, and live WebSocket synchronization to keep teams perfectly in sync.



## ✨ Key Features

- **⚡ Real-Time Sync:** Instant updates across all connected clients using WebSockets. When a teammate moves a card, you see it instantly.
- **🖐️ Intuitive Drag & Drop:** Fluid card and list reordering for effortless, visual task management.
- **🔐 Secure Authentication:** Full auth suite featuring JWT-based login, signup, password resets, and secure session management.
- **🎨 Premium UI/UX:** Custom "Vibrant Aurora" dark theme and a pristine "Clean Minimalist" white theme, featuring deep slate cards, teal accents, and smooth hover micro-animations.
- **🤝 Collaboration Hub:** Share boards with teammates, manage access levels, and respond to pending board invitations directly from your dashboard.
- **🔍 Advanced Search:** Rapidly filter and locate cards across all your workspaces.
- **👤 Profile Management:** Personalized user profiles with avatar uploads and account settings.

## 🛠️ Tech Stack

**Frontend:**
- **Vue 3** (Composition API, `<script setup>`)
- **Vite** (Lightning-fast build tool)
- **Vue Router** (Client-side routing)
- **Vanilla CSS** (Custom design tokens, CSS variables, and flex/grid layouts)

**Backend:**
- **Go (Golang)** (High concurrency and performance)
- **Gin Web Framework** (Robust HTTP routing)
- **GORM** (Developer-friendly ORM)
- **Gorilla WebSockets** (Real-time communication)
- **MySQL** (Relational data storage)

---

## 🚀 Getting Started

### Prerequisites
- [Go 1.18+](https://golang.org/doc/install)
- [Node.js v16+](https://nodejs.org/)
- MySQL Server

### 1. Database Setup
Create a new MySQL database named `trello_clone`:
```sql
CREATE DATABASE trello_clone;
```

### 2. Backend Initialization
```bash
# Navigate to the backend directory
cd backend

# Create your .env file
cp .env.example .env 
# (Make sure to configure your DB_DSN and JWT_SECRET inside .env)

# Download Go modules
go mod tidy

# Start the Gin server (runs on :8080 by default)
go run .
```

### 3. Frontend Initialization
```bash
# Open a new terminal and navigate to the frontend directory
cd frontend

# Install dependencies
npm install

# Start the Vite development server
npm run dev
```

Your application should now be running locally! Navigate to `http://localhost:5173` (or the port Vite provides) in your browser.

---

