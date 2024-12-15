# Product Management System with Asynchronous Image Processing

This project is a backend system for managing product data, built using **Golang**. It emphasizes **asynchronous image processing**, **caching**, and **high scalability** to handle real-world use cases effectively.

---

## Project Features

### 1. API Design
- **POST /products**: Accepts product details including:
  - `user_id` (reference to the user table)
  - `product_name` (string)
  - `product_description` (text)
  - `product_images` (array of image URLs)
  - `product_price` (decimal)

- **GET /products/:id**: Fetches product details by ID, including processed images.
  
- **GET /products**: Fetches all products for a specific user, with optional filters like:
  - `name` (product name)
  - `min_price` & `max_price` (price range)
  - Pagination using `page` and `limit` parameters.

- **PUT /products/:id**: Updates product details by ID.
  
- **DELETE /products/:id**: Deletes a product by ID.

---

### 2. Technologies Used
- **Programming Language**: Golang
- **Database**: PostgreSQL
- **Message Queue**: RabbitMQ
- **Cache**: Redis
- **Framework**: Gin-Gonic for REST APIs
- **Logging**: Zap for structured logging
- **Asynchronous Tasks**: RabbitMQ for queuing image URLs
- **Testing**: Postman for API testing

---

### 3. Key Features
#### **Data Storage**
- Products and users are stored in PostgreSQL with the following schema:
  - **Products**:
    - `product_name`
    - `product_description`
    - `product_images`
    - `compressed_product_images`
    - `product_price`
  - **Users**:
    - Basic user details with a one-to-many relationship with products.

#### **Asynchronous Image Processing**
- Image URLs are added to a RabbitMQ queue upon product creation.
- A microservice fetches these URLs, compresses the images, and updates the database.

#### **Caching**
- Redis is used to cache product data fetched by `GET /products/:id`.
- Cache invalidation is implemented during updates or deletions.

#### **Logging**
- Structured logs capture API requests, responses, and key events like:
  - Image processing success/failure
  - Cache hits/misses
  - Database operations

#### **Error Handling**
- Centralized error handling with descriptive messages for database errors, API issues, and queue processing failures.

#### **Testing**
- API testing done via Postman, with comprehensive scenarios covering:
  - Happy path
  - Edge cases
  - Error cases

---

## Setup Instructions

### 1. Prerequisites
- Golang installed (version >= 1.20)
- Docker installed for RabbitMQ, Redis, and PostgreSQL
- Postman for API testing

### 2. Clone the Repository
```bash
git clone https://github.com/Atharv175/product-management-system.git
cd product-management-system
3. Set Up Environment Variables
Create a .env file with the following:

env
Copy code
DB_HOST=<Your PostgreSQL Host>
DB_PORT=5432
DB_USER=<Your Database User>
DB_PASSWORD=<Your Database Password>
DB_NAME=product_management_db

REDIS_HOST=localhost
REDIS_PORT=6379

RABBITMQ_URL=amqp://guest:guest@localhost:5672/
4. Start Dependencies Using Docker
Start PostgreSQL, Redis, and RabbitMQ containers:

bash
Copy code
docker run --name postgres -e POSTGRES_USER=<Your Database User> -e POSTGRES_PASSWORD=<Your Database Password> -p 5432:5432 -d postgres
docker run --name redis -p 6379:6379 -d redis
docker run --name rabbitmq -p 5672:5672 -p 15672:15672 -d rabbitmq:management
5. Run the Application
bash
Copy code
go run main.go image_processor.go
6. Test the APIs
Open Postman and import the API collection.
Use the following base URL:
bash
Copy code
http://localhost:8080
Directory Structure
go
Copy code
product-management-system/
├── main.go
├── image_processor.go
├── routes/
│   └── routes.go
├── database/
│   └── connection.go
├── models/
│   └── product.go
├── Dockerfile
├── go.mod
├── .env
Deployment
Push the repository to GitHub.
