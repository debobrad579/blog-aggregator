# Gator CLI

Gator is an RSS feed aggregator written in Go.

## Features
* Add RSS feeds from across the internet to be collected
* Store the collected posts in a PostgreSQL database
* Follow and unfollow RSS feeds that other users have added
* View summaries of the aggregated posts in the terminal, with a link to the full post

## Prerequisites

Before running Gator, make sure you have the following installed:

- [Go](https://golang.org/doc/install) (version 1.20 or higher recommended)
- [PostgreSQL](https://www.postgresql.org/download/) (version 15 or higher recommended)

### Goose (Database Migrations)

Gator uses [Goose](https://github.com/pressly/goose) to manage database migrations. Install it using Go:

```bash
go install github.com/pressly/goose/v3/cmd/goose@latest
```

## Installation

### Clone the Repository
```bash
git clone https://github.com/debobrad579/terminal-chess.git
cd blog-aggregator
```

### Apply Migrations
```bash
goose -dir ./sql/schema postgres "<your_db_url>" up
```

### Build
```bash
go build
```

### Run
```bash
./blog-aggregator <command> [arguments]
```

## Configuration

Gator uses a configuration file named `.gatorconfig.json` in your home directory to store database and user settings.

1. Create a `.gatorconfig.json` file in your home directory:

   - **Linux / macOS:**
     ```bash
     nano ~/.gatorconfig.json
     ```
   - **Windows (PowerShell):**
     ```powershell
     notepad $HOME\.gatorconfig.json
     ```

2. Add your configuration in JSON format:

```json
{
  "db_url": "postgres://<username>:<password>@<host>:<port>/<database>?sslmode=disable"
}
```

## CLI Commands

Once Gator is installed and your database is set up, you can run the following commands:

| Command       | Description |
|---------------|-------------|
| `login`       | Log in as an existing user. |
| `register`    | Register a new user account. |
| `reset`       | Reset user-related data. |
| `users`       | List all users in the system. |
| `addfeed`     | Add a new feed to the system. |
| `feeds`       | List all available feeds. |
| `follow`      | Follow a feed. |
| `following`   | Show the feeds the current user is following. |
| `unfollow`    | Unfollow a feed. |
| `agg`         | Aggregate posts from feeds. |
| `browse`      | Browse posts. |
