FROM node:24-alpine AS builder

WORKDIR /app

ARG NEXT_PUBLIC_BUILD_RELEASE=local-dev
ARG NEXT_PUBLIC_BUILD_BRANCH=workspace
ARG NEXT_PUBLIC_BUILD_COMMIT=uncommitted
ARG NEXT_PUBLIC_BASE_PATH=
ARG NEXT_PUBLIC_API_BASE_URL=

ENV NEXT_PUBLIC_BUILD_RELEASE=$NEXT_PUBLIC_BUILD_RELEASE
ENV NEXT_PUBLIC_BUILD_BRANCH=$NEXT_PUBLIC_BUILD_BRANCH
ENV NEXT_PUBLIC_BUILD_COMMIT=$NEXT_PUBLIC_BUILD_COMMIT
ENV NEXT_PUBLIC_BASE_PATH=$NEXT_PUBLIC_BASE_PATH
ENV NEXT_PUBLIC_API_BASE_URL=$NEXT_PUBLIC_API_BASE_URL

COPY frontend/package.json frontend/package-lock.json frontend/.npmrc ./
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
