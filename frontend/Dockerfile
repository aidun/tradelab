FROM node:24-alpine AS builder

WORKDIR /app

COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci

COPY frontend ./

RUN npm run build

FROM node:24-alpine

RUN addgroup -S tradelab && adduser -S tradelab -G tradelab

WORKDIR /app

ENV NODE_ENV=production
ENV PORT=3000
ENV HOSTNAME=0.0.0.0

COPY --from=builder /app/.next/standalone ./
COPY --from=builder /app/.next/static ./.next/static

USER tradelab

EXPOSE 3000

CMD ["node", "server.js"]
