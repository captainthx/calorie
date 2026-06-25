# syntax=docker/dockerfile:1
FROM node:22-alpine AS build
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build

FROM nginx:alpine
COPY --from=build /app/dist /usr/share/nginx/html

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

COPY <<'EOF' /entrypoint.sh
#!/bin/sh
set -eu

cat > /usr/share/nginx/html/env.js <<ENVEOF
window._env_ = {
  API_BASE_URL: "${API_BASE_URL:-http://localhost:8080/api}"
};
ENVEOF

exec nginx -g 'daemon off;'
EOF

RUN <<EOF
chown -R nginx:nginx /usr/share/nginx/html /var/cache/nginx /var/log/nginx
chown nginx:nginx /etc/nginx/conf.d/default.conf
chown nginx:nginx /entrypoint.sh
chmod 755 /entrypoint.sh
touch /var/run/nginx.pid
chown nginx:nginx /var/run/nginx.pid
EOF

USER nginx
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
  CMD wget -q -O /dev/null http://127.0.0.1:8080/ || exit 1

CMD ["/entrypoint.sh"]
