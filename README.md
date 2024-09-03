# GoPanel

GoPanel is a minimal server management panel written in Go. It is designed to support LEMP (Linux, Nginx, MySQL/MariaDB, PHP) stacks and provides a streamlined solution for managing servers.

![image](https://github.com/user-attachments/assets/688534cf-b5e1-4f98-8d3a-c9aece815289)

## Features

- **Minimal Authentication Flow**: Very simple secure access with only PIN-token authentication (2FA but it's 1FA) to ensure that only authorized users can manage the server.
- **Simple API Framework**: Provides a straightforward API for integration and automation.
- **Portable**: The entire panel operates from a single binary, simplifying deployment and management.

## Building from source

### Prerequisites

- [Go](https://golang.org/dl/) (version 1.23 or higher) installed on your system.

1. **Clone the Repository**:

    ```bash
    git clone https://github.com/xaymup/gopanel.git
    cd gopanel
    ```

2. **Build the Binary**:

    ```bash
    go build -o gopanel cmd/gopanel/main.go
    ```

3. **Run GoPanel**:

    ```bash
    sudo ./gopanel
    ```

### Testing

To test GoPanel, you can run it directly using the following command:

```bash
sudo go run cmd/gopanel/main.go
```

### Usage

1. Access the Panel: Open your web browser and navigate to http://localhost:8888.
2. Authenticate: Use your 2FA authenticator to log in.

### Contributing

If you would like to contribute to GoPanel, please follow these steps:

1. Fork the repository.
2. Create a new branch (git checkout -b feature-branch).
3. Make your changes and commit them (git commit -am 'Add new feature').
4. Push to the branch (git push origin feature-branch).
5. Create a new Pull Request.

### License

GoPanel is licensed under the MIT License. See the LICENSE file for more information.

### Contact

For any questions or support, please reach out to hi@xaymup.me.






