# Stage 1: Build
FROM node:18-alpine as build

WORKDIR /app

# Копирование package.json и установка зависимостей
COPY package.json ./
RUN npm install

# Копирование исходного кода и сборка
COPY . ./
RUN npm run build

# Stage 2: Serve
FROM nginx:1.21-alpine

# Копирование собранного приложения
COPY --from=build /app/build /usr/share/nginx/html

# Копирование конфигурации nginx
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]