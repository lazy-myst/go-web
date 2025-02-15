# Use Node.js 18 as the base image for the frontend
FROM node:18

# Set the working directory inside the container
WORKDIR /app

# Copy package.json and package-lock.json (or yarn.lock) to install dependencies
COPY package.json package-lock.json ./

# Install frontend dependencies
RUN npm install

# Copy the rest of the frontend source files
COPY . .

# Expose the default port Vite runs on (5173)
EXPOSE 5173

# Run the Vite development server
CMD ["npm", "run", "dev"]
