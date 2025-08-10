# Use Node.js 18 as the base image
FROM node:22.13.0

# Set the working directory inside the container
WORKDIR /app

# Copy package files
COPY package.json package-lock.json ./

# Install dependencies
RUN npm install

# Copy the entire project
COPY . .

# Expose Vite's development port
EXPOSE 5173

# Enable polling for file changes
ENV CHOKIDAR_USEPOLLING=true
ENV WATCHPACK_POLLING=true

# Run Vite with polling enabled
CMD ["npm", "run", "dev"]
