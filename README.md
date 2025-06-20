# Go DB Manager

This project is a Go-based API server for managing various database systems. It provides a set of HTTP endpoints to perform common database operations.

## Supported Databases

*   MySQL
*   MongoDB
*   PostgreSQL

## API Endpoints

All endpoints are of the POST method.

### Common Endpoints

*   `/server-info` (GET): Retrieves information about the server environment.

### MySQL Endpoints

*   `/mysql/create-database`: Creates a new MySQL database.
*   `/mysql/reset-credentials`: Resets credentials for a MySQL database.
*   `/mysql/rename-database`: Renames an existing MySQL database.
*   `/mysql/delete-database`: Deletes a MySQL database.
*   `/mysql/view-database-stats`: Views statistics for a MySQL database.

### MongoDB Endpoints

*   `/mongo/create-database`: Creates a new MongoDB database.
*   `/mongo/reset-credentials`: Resets credentials for a MongoDB user/database.
*   `/mongo/rename-database`: Renames an existing MongoDB database. (Note: MongoDB rename operations might have specific behaviors, like being limited to admin database or specific commands).
*   `/mongo/delete-database`: Deletes a MongoDB database.
*   `/mongo/view-database-stats`: Views statistics for a MongoDB database.

### PostgreSQL Endpoints

*   `/postgres/create-database`: Creates a new PostgreSQL database.
*   `/postgres/reset-credentials`: Resets credentials for a PostgreSQL role/database.
*   `/postgres/rename-database`: Renames an existing PostgreSQL database.
*   `/postgres/delete-database`: Deletes a PostgreSQL database.
*   `/postgres/view-database-stats`: Views statistics for a PostgreSQL database.
*   `/postgres/get-total-queries`: Retrieves the total number of queries executed against a PostgreSQL database (Note: this might require specific pg_stat_statements extension or similar).


*Further investigation is needed to document the specific request body parameters for each endpoint.*

## Configuration

Database connection details are configured via environment variables, typically loaded from a `.env` file in the project root.

**Required Environment Variables:**

*   **MySQL:**
    *   `mysql_host` (default: `localhost`)
    *   `mysql_user` (default: `root`)
    *   `mysql_password` (default: ``)
    *   `mysql_port` (default: `3306`)
*   **MongoDB:**
    *   `mongo_uri` (default: `mongodb://admin:password@localhost:27017`)
*   **PostgreSQL:**
    *   `postgres_host` (default: `localhost`)
    *   `postgres_user` (default: `postgres`)
    *   `postgres_password` (default: `password`)
    *   `postgres_port` (default: `5432`)

Create a `.env` file in the root of the project and add the necessary variables:

```env
mysql_host=your_mysql_host
mysql_user=your_mysql_user
mysql_password=your_mysql_password
mysql_port=your_mysql_port

mongo_uri=your_mongodb_uri

postgres_host=your_postgres_host
postgres_user=your_postgres_user
postgres_password=your_postgres_password
postgres_port=your_postgres_port
```

## Building and Running

1.  **Clone the repository:**
    ```bash
    git clone <repository-url>
    cd go-db-manager
    ```
2.  **Install dependencies:**
    ```bash
    go mod tidy
    ```
3.  **Create and configure your `.env` file** as described in the Configuration section.
4.  **Run the application:**
    ```bash
    go run .
    ```
    The server will start on port `8080` by default.

## Contributing

Contributions are welcome! Please feel free to:
*   Report a bug
*   Suggest a feature
*   Submit a pull request

Please make sure to update tests as appropriate.
