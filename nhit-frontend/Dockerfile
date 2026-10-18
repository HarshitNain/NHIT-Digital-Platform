# ==============================================================================
# NHIT FRONTEND - MULTI-STAGE PRODUCTION DOCKERFILE
# Stage 1: Build React 18 / Vite / TypeScript SPA
# Stage 2: Serve optimized static bundle via Nginx Alpine
# ==============================================================================

# ---- Build Stage ----
FROM node:22-alpine AS builder

WORKDIR /app

# Cache package dependencies
COPY package*.json ./
RUN npm ci

# Copy sources and compile
COPY . .
RUN npm run build

# ---- Production Runtime Stage ----
FROM nginx:alpine

# Copy built dist files
COPY --from=builder /app/dist /usr/share/nginx/html

# Copy custom nginx configuration for SPA routing
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
