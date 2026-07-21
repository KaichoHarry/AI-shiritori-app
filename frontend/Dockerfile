FROM node:22-alpine AS builder
WORKDIR /app
COPY package.json package-lock.json* ./
RUN npm ci
COPY . .
# publicディレクトリは中身が空の場合Gitに追跡されず存在しないことがあるため、
# 後段のCOPY --from=builderが失敗しないよう保証しておく。
RUN mkdir -p public
# NEXT_PUBLIC_*はNext.jsのビルド時に静的に埋め込まれる値のため、Dockerfile内で
# ARGとして明示的に受け取りビルド前にENV化する必要がある(Railway等のDockerfile
# ベースのビルドでは、ARG宣言のない環境変数はビルドステージに渡されないため)。
ARG NEXT_PUBLIC_API_BASE_URL
ENV NEXT_PUBLIC_API_BASE_URL=$NEXT_PUBLIC_API_BASE_URL
RUN npm run build

FROM node:22-alpine
WORKDIR /app
ENV NODE_ENV=production
COPY --from=builder /app/public ./public
COPY --from=builder /app/.next ./.next
COPY --from=builder /app/node_modules ./node_modules
COPY --from=builder /app/package.json ./package.json
EXPOSE 3000
CMD ["npm", "run", "start"]
