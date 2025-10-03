# ---- Build stage ----
FROM node:20 AS builder

WORKDIR /app

# Install dependencies
COPY frontend/package.json frontend/frontend/package-lock.json ./
RUN npm install

# Copy source and build
COPY frontend/frontend/ .
RUN npm run build --prod

# ---- Serve with Nginx ----
FROM nginx:1.25

COPY --from=builder /app/dist/frontend/browser /usr/share/nginx/html

# Copy custom nginx config (optional)
COPY deploy/nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
