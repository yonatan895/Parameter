FROM node:20-alpine AS build
WORKDIR /app

# Cache dependencies
COPY package.json package-lock.json tsconfig.json webpack.config.js ./
COPY .eslintrc.js ./
RUN --mount=type=cache,target=/root/.npm \
    npm ci

# Copy rest of the source after installing dependencies
COPY src ./src
COPY index.html ./

RUN npm run lint --if-present
RUN npm run build

FROM nginx:alpine
COPY --from=build /app/index.html /usr/share/nginx/html/index.html
COPY --from=build /app/dist/bundle.js /usr/share/nginx/html/bundle.js
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
