FROM node:18-alpine

WORKDIR /app

# Copy and install dependencies
COPY package.json yarn.lock ./
RUN yarn install

# Copy the rest of the source code
COPY . .

# Start the app using Yarn
CMD ["yarn", "start"]
