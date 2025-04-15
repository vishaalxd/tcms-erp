### Step-by-Step Deployment and CI/CD Pipeline Setup

Let's break down the process of setting up a deployment and CI/CD pipeline for a React frontend and Golang backend application using Docker, Docker Compose, and AWS ECS. We'll cover each step in a straightforward manner.

### 1. Docker Basics

**Docker** is a tool designed to make it easier to create, deploy, and run applications by using containers. Containers allow a developer to package up an application with all the parts it needs, such as libraries and other dependencies, and ship it all out as one package.

#### Dockerizing Applications

**React Dockerfile (`frontend/Dockerfile`):**

1. Build the React application.
2. Serve the built files using Nginx.

```Dockerfile
# Stage 1: Build the React application
FROM node:14 AS build
WORKDIR /app
COPY package*.json ./
RUN npm install
COPY . .
RUN npm run build

# Stage 2: Serve the built files with Nginx
FROM nginx:alpine
COPY --from=build /app/build /usr/share/nginx/html
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
```

**Go Dockerfile (`backend/Dockerfile`):**

1. Build the Go application.
2. Run the built binary.

```Dockerfile
# Stage 1: Build the Go application
FROM golang:1.16 AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o main .

# Final Stage: Run the built binary
FROM alpine:latest
WORKDIR /root/
COPY --from=build /app/main .
EXPOSE 8080
CMD ["./main"]
```

### 2. Docker Compose

**Docker Compose** is a tool for defining and running multi-container Docker applications.

**docker-compose.yml:**

```yaml
version: '3'
services:
  frontend:
    image: your-docker-repo/frontend:latest
    ports:
      - "80:80"

  backend:
    image: your-docker-repo/backend:latest
    ports:
      - "8080:8080"
    environment:
      - PORT=8080
```

### 3. CI/CD with GitHub Actions

**CI/CD** (Continuous Integration and Continuous Deployment) is a method to frequently deliver apps to customers by introducing automation into the stages of app development.

**GitHub Actions** allows you to create automated workflows like building and deploying applications.

**GitHub Actions Workflow (`.github/workflows/ci-cd-pipeline.yml`):**

```yaml
name: CI/CD Pipeline

on:
  push:
    branches:
      - main

jobs:
  build-and-deploy:
    runs-on: ubuntu-latest

    steps:
    - name: Checkout code
      uses: actions/checkout@v2

    - name: Set up Docker Buildx
      uses: docker/setup-buildx-action@v1

    - name: Log in to Docker Hub
      uses: docker/login-action@v1
      with:
        username: ${{ secrets.DOCKER_USERNAME }}
        password: ${{ secrets.DOCKER_PASSWORD }}

    - name: Build and push Frontend Docker image
      uses: docker/build-push-action@v2
      with:
        context: ./frontend
        file: ./frontend/Dockerfile
        push: true
        tags: your-docker-repo/frontend:latest

    - name: Build and push Backend Docker image
      uses: docker/build-push-action@v2
      with:
        context: ./backend
        file: ./backend/Dockerfile
        push: true
        tags: your-docker-repo/backend:latest

    - name: Deploy with Docker Compose
      run: |
        docker-compose -f docker-compose.yml up -d
```

### 4. AWS ECS (Elastic Container Service)

**Amazon ECS** is a fully managed container orchestration service that makes it easy to deploy, manage, and scale containerized applications.

#### Setting Up AWS ECS

1. **Create a Cluster:**
   - Go to the ECS section of the AWS Management Console.
   - Click "Create Cluster" and follow the prompts to create a new ECS cluster.

2. **Task Definition:**
   - Create a new Task Definition which is like a blueprint for your containers.
   - Define the container specifications, including Docker images, ports, and environment variables.

3. **Service:**
   - Create a new Service which runs and maintains your desired number of copies of a task definition simultaneously in an ECS cluster.
   - Specify the number of tasks, load balancer settings, and other configurations.

#### ECS Task Definition Example:

```json
{
  "family": "task-definition-name",
  "networkMode": "awsvpc",
  "containerDefinitions": [
    {
      "name": "frontend",
      "image": "your-docker-repo/frontend:latest",
      "portMappings": [
        {
          "containerPort": 80,
          "hostPort": 80,
          "protocol": "tcp"
        }
      ]
    },
    {
      "name": "backend",
      "image": "your-docker-repo/backend:latest",
      "portMappings": [
        {
          "containerPort": 8080,
          "hostPort": 8080,
          "protocol": "tcp"
        }
      ],
      "environment": [
        {
          "name": "PORT",
          "value": "8080"
        }
      ]
    }
  ],
  "requiresCompatibilities": ["EC2"],
  "memory": "512",
  "cpu": "256"
}
```

### Step-by-Step Guide

#### Step 1: Dockerizing Applications

1. **Create Dockerfiles for both React and Golang applications.**
2. **Build Docker images locally.**

```sh
docker build -t your-docker-repo/frontend:latest -f docker/frontend/Dockerfile .
docker build -t your-docker-repo/backend:latest -f docker/backend/Dockerfile .
```

#### Step 2: Pushing Docker Images to Docker Hub

1. **Log in to Docker Hub.**

```sh
docker login --username=your-username
```

2. **Push Docker images.**

```sh
docker push your-docker-repo/frontend:latest
docker push your-docker-repo/backend:latest
```

#### Step 3: Deploying with Docker Compose

1. **Ensure `docker-compose.yml` is configured correctly.**
2. **Run Docker Compose.**

```sh
docker-compose up -d
```

#### Step 4: Setting Up GitHub Actions for CI/CD

1. **Create `.github/workflows/ci-cd-pipeline.yml`.**
2. **Configure secrets for Docker Hub in GitHub repository settings.**

#### Step 5: Deploying with AWS ECS

1. **Set up an ECS Cluster, Task Definitions, and Services in AWS ECS.**
2. **Deploy ECS Services using the provided Task Definitions.**

### Conclusion

Using Docker, Docker Compose, and AWS ECS provides a flexible, scalable, and cost-effective way to deploy and manage your applications. The CI/CD pipeline with GitHub Actions automates the process of building, testing, and deploying your applications, making the workflow efficient and reliable. By following the above steps, you can achieve a modern deployment setup with reduced costs and complexity.