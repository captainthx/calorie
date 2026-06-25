FROM node:22-alpine AS build
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build

FROM nginx:alpine
COPY --from=build /app/dist /usr/share/nginx/html
# ponytail: nginx config inline — SPA routing needs try_files so react-router deep paths don't 404
RUN printf 'server {\n    listen 80;\n    root /usr/share/nginx/html;\n    index index.html;\n    location / {\n        try_files $uri $uri/ /index.html;\n    }\n}\n' > /etc/nginx/conf.d/default.conf
EXPOSE 80
# ponytail: generate env.js at container start so API_BASE_URL is runtime-configurable
CMD ["/bin/sh", "-c", "echo \"window._env_ = { API_BASE_URL: '${API_BASE_URL:-http://localhost:8080/api}' };\" > /usr/share/nginx/html/env.js && nginx -g 'daemon off;'"]
