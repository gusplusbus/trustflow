# Stage 1: Build
FROM node:20-alpine AS builder

WORKDIR /app
COPY . .
RUN yarn install
RUN yarn build

# Stage 2: Serve static content
FROM nginx:alpine

COPY --from=builder /app/dist /usr/share/nginx/html

# Remove default nginx config and copy ours (optional)
# COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
