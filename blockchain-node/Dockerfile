FROM node:18-alpine

WORKDIR /app

# Install dependencies
COPY package.json yarn.lock ./
RUN yarn install

# Copy the rest of the files
COPY . .

# Set NODE_OPTIONS for ESM and ts-node
ENV NODE_OPTIONS="--loader ts-node/esm"

CMD ["node", "--loader", "ts-node/esm", "src/index.ts"]
