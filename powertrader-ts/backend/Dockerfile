FROM node:18-alpine

WORKDIR /app

# Install build dependencies for native modules (tfjs-node, sqlite3)
RUN apk add --no-cache python3 make g++

# Copy package files
COPY package*.json ./

# Install dependencies
RUN npm install

# Copy source
COPY . .

# Build TypeScript
RUN npm run build

# Expose API port
EXPOSE 3000

# Start command
CMD ["npm", "start"]
