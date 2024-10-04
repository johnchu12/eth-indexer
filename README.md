# Project

## Project Introduction

eth-indexer is a blockchain-based application designed to provide efficient data indexing and API services. The project comprises multiple services, including a database (PostgreSQL), an indexer, and an API service, all containerized using Docker for deployment. The project is developed in Go.

## Architecture Overview

### Services Composition

1. **PostgreSQL Database (`pgsql`)**
   - **Description**: Responsible for storing the core data of the application, including user information, transaction records, and more.
   - **Configuration**:
     - Uses the `postgres:15` Docker image.

2. **Indexer Service (`indexer`)**
   - **Description**: Listens to blockchain events, processes and indexes contract events (such as UniswapV2, USDC, AAVE, etc.), and stores the processed data in the PostgreSQL database.
   - **Extended Features**:
     - **Multi-Chain Support**: The indexer service can operate on multiple different blockchains, listening to various contract events and providing cross-chain data indexing capabilities.
     - **Custom Event Handling**: Users can customize the event handling logic based on their needs, allowing flexible processing of specific events from different contracts to meet various business requirements.

3. **API Service (`api`)**
   - **Description**: Provides RESTful APIs to display backend data, including user information and transaction records.

### Architecture Diagram

```
+-----------+           +-----------+
|    API    |           |  Indexer  |
+-----------+           +-----------+
      |                       |
      |                       |
      v                       v
   +-------------------------------+
   |            PostgreSQL         |
   +-------------------------------+
```

## Installation and Usage

### Prerequisites
- Golang 1.22
- Docker and Docker Compose
- Make

### Environment Setup

1. **Clone the Project**

   ```bash
   git clone https://github.com/johnchu12/eth-indexer.git
   cd eth-indexer
   ```

2. **Set Environment Variables**

   The project uses the `.env/server.env` file to configure environment variables for the API service. Ensure that `PORT` is set correctly.
   The default port is `8080`.
   ```env
   PORT=8080
   ```

### Using Makefile Commands

The Makefile provides a series of convenient commands to build and manage the project.

- **Start the Indexer Service**

  ```bash
  DATABASE_URL=postgresql://postgres:mysecretpassword@localhost:5432/pelith?sslmode=disable \
  make start
  ```

- **Start the API Service**

  ```bash
  DATABASE_URL=postgresql://postgres:mysecretpassword@localhost:5432/pelith?sslmode=disable \
  make api
  ```

  - **Build a Single Service Image**

  ```bash
  make build target={service_name}
  ```

  For example, to build the API image:

  ```bash
  make build target=api
  ```

- **Build All Services**

  Build images for all services in the project.

  ```bash
  make build-all
  ```

- **Run Tasks**

  Execute specific tasks, such as processing `usdcweth` transactions.
  - The main content of this task is located at `/cmd/task/usdcweth/main.go`

  ```bash
  DATABASE_URL=postgresql://postgres:mysecretpassword@localhost:5432/pelith?sslmode=disable \
  make task
  ```

### Using Docker Compose

The project uses `docker-compose.yml` to define and manage the operation of multiple services. You can use the following command to start all services:

```bash
make docker-compose
```

This command will build all defined service images and start the containers.

## Service Descriptions

### API Routes

| Endpoint              | Description                       |
| --------------------- | --------------------------------- |
| `/leaderboard`        | Displays the user leaderboard     |
| `/user/:id`           | Displays detailed information of a single user |
| `/user/:id/history`   | Displays the point history data of a single user |
| `/ping`               | Health check            |

### Indexer Service

- **Features**:
  - Listens to specified contract events (such as UniswapV2's Swap events, USDC's Transfer and Approval events, etc.).
  - Supports operation on multiple different blockchains, listening to various contract events and providing cross-chain data indexing capabilities.
  - Provides custom event handling, allowing users to process specific events from specific contracts based on their needs.
  - Stores the processed relevant data into the PostgreSQL database.

- **Configuration**:
  - **ABI Configuration**: Located in the `internal/indexer/abis/` directory, defines the ABIs and related events of contracts.
  - **Contract Configuration**: `internal/indexer/config.json` defines supported networks, contract addresses, and starting blocks.

## Migrations

The project utilizes [golang-migrate](https://github.com/golang-migrate/migrate) for managing database migrations. When the Indexer service starts, it automatically runs migrations to ensure the database schema is up-to-date.
