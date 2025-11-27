FROM node:18-alpine AS builder

WORKDIR /app

# Copy package files
COPY package*.json ./

# Install dependencies
RUN npm ci

# Copy source code
COPY . .

# Build the app
RUN npm run build

# Production stage
FROM node:18-alpine

WORKDIR /app

# Copy package files
COPY package*.json ./

# Install only production dependencies
RUN npm ci --only=production

# Copy built app from builder
COPY --from=builder /app/dist ./dist

# Create app user
RUN addgroup -g 1001 -S nodejs
RUN adduser -S electron -u 1001

USER electron

EXPOSE 3000

CMD ["npm", "start"]