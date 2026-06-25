# syntax=docker/dockerfile:1
FROM node:22-alpine AS build
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build

FROM nginx:alpine
COPY --from=build /app/dist /usr/share/nginx/html

# ponytail: nginx config inline — SPA routing needs try_files so react-router deep paths don't 404
COPY <<'EOF' /etc/nginx/conf.d/default.conf
server {
    listen 8080;
    root /usr/share/nginx/html;
    index index.html;
    location / {
        try_files $uri $uri/ /index.html;
    }
}
EOF

RUN <<EOF
chown -R nginx:nginx /usr/share/nginx/html /var/cache/nginx /var/log/nginx
chown nginx:nginx /etc/nginx/conf.d/default.conf
touch /var/run/nginx.pid
chown nginx:nginx /var/run/nginx.pid
EOF

USER nginx
EXPOSE 8080
# ponytail: generate env.js at container start so API_BASE_URL is runtime-configurable
CMD ["/bin/sh", "-c", \
    "echo \"globalThis._env_ = { API_BASE_URL: '${API_BASE_URL:-http://localhost:8080/api}' };\" \
        > /usr/share/nginx/html/env.js \
    && nginx -g 'daemon off;'"]
