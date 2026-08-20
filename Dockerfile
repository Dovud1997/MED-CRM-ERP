FROM node:22-alpine AS base
RUN corepack enable
WORKDIR /app
COPY package.json pnpm-workspace.yaml pnpm-lock.yaml* turbo.json tsconfig.base.json ./
COPY apps ./apps
COPY packages ./packages
RUN pnpm install --frozen-lockfile=false
ARG APP
ENV APP=${APP}
RUN pnpm --filter @clinicos/${APP} build
CMD ["sh","-c","pnpm --filter @clinicos/${APP} start"]
