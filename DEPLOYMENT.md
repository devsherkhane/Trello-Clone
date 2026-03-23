# Deployment Guide for Drift

This guide explains how to deploy the Drift application using Docker Compose.

## Prerequisites

1.  **Docker & Docker Compose**: Ensure you have Docker installed on your server.
2.  **Domain/Public IP**: You'll need a way for the frontend to reach the backend.

## Step 1: Clone the Repository
```bash
git clone <your-repo-url>
cd Drift
```

## Step 2: Configure Environment Variables

### Backend (`backend/.env`)
Create a `.env` file in the `backend` directory based on `.env.example`:
```bash
DB_PASSWORD=your_secure_db_password
JWT_SECRET_KEY=your_secure_jwt_secret
```

### Frontend (`frontend/.env`)
Create a `.env` file in the `frontend` directory:
```bash
VITE_API_URL=http://<your-server-ip-or-domain>:8080/api
VITE_WS_URL=ws://<your-server-ip-or-domain>:8080/api/ws
```

## Step 3: Deployment with Docker Compose

Run the following command in the root directory:
```bash
docker-compose up -d --build
```

This will:
- Spin up a MySQL 8.0 database.
- Build and run the Go backend on port `8080`.
- Build the Vue frontend using Vite and serve it via Nginx on port `3000`.

## Accessing the Application
- **Frontend**: `http://<your-server-ip-or-domain>:3000`
- **Backend API**: `http://<your-server-ip-or-domain>:8080/api`
- **Swagger Documentation**: `http://<your-server-ip-or-domain>:8080/swagger/index.html`

---

## Alternative: Cloud Deployment (Render.com)

Render is a zero-config cloud platform that can deploy this entire stack using a **Blueprint**.

1. **Push to GitHub**: Ensure your code is in a public or private GitHub repository.
2. **Connect to Render**: Go to [dashboard.render.com](https://dashboard.render.com/) and create a "New Blueprint".
3. **Select Repository**: Pick your Drift repository.
4. **Deploy**: Render will read `render.yaml` and spin up:
    - **MySQL Database**
    - **Go Backend** (Web Service)
    - **Vue Frontend** (Static Site)

*Note: Render will automatically handle SSL/HTTPS for you.*

---

## Production Considerations
- **HTTPS**: It's highly recommended to use a reverse proxy like Nginx or Caddy with Let's Encrypt for SSL.
- **Backups**: Ensure you back up the `mysql_data` Docker volume regularly.
- **Monitoring**: Consider using tools like Prometheus/Grafana or basic logging services to monitor the backend performance.
