# Popkat v2
Popkat v2 is a high-performance CDN microservice under Purrquinox, designed to handle file storage and retrieval efficiently. Built with Golang, it leverages MinIO or SeaweedFS for scalable object storage.

## Features

- **Fast and Lightweight**: Optimized for high-speed file access and minimal resource usage.
- **Golang Backend**: Uses Go for efficient concurrency and performance.
- **MinIO Support**: S3-compatible storage API for easy cloud compatibility.
- **SeaweedFS Support**: Distributed storage for scalability and redundancy.
- **Secure Access Control**: Supports authentication and permission management.
- **API-Driven**: RESTful API for easy integration with other services.

## Installation

### Prerequisites

- Golang 1.20+
- MinIO instance or SeaweedFS instance
- PostgreSQL

### Setup

1. Clone the repository:
   ```sh
   git clone https://github.com/purrquinox/popkat-v2.git
   cd popkat-v2
   ```
2. Install dependencies:
   ```sh
   go mod tidy
   ```
3. Configure environment variables:
   ```sh
   make
   ./popkat
   cp -r config.yaml.sample config.yaml
   nano config.yaml  # Update with your credentials
   ```
4. Build and run the service:
   ```sh
   make
   ./popkat
   ```

## License

Popkat v2 is licensed under the MIT License. See `LICENSE` for details.

---

**Purrquinox** | [GitHub](https://github.com/purrquinox) | *Scalable, Fast, and Reliable CDN*

