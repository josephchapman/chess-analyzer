# chess-analyzer

Provides an API to allow the analysis of chess games.

If `docker-compose` is installed, use it to run the container with exposed port and mounted persistent storage:
```bash
docker compose up
```

Set player name:
```bash
PLAYER="playername"
```

View available archives:
```bash
curl -X GET "http://127.0.0.1:24377/api/${PLAYER}"
```

Refresh available archives:
```bash
curl -X POST "http://127.0.0.1:24377/api/${PLAYER}"
```

View available games within the `2025-02` archive:
```bash
curl -X GET "http://127.0.0.1:24377/api/${PLAYER}/2025-02"
```

Refresh the `2025-02` archive:
```bash
curl -X POST "http://127.0.0.1:24377/api/${PLAYER}/2025-02"
```

View details of the `282ba89a-44b0-11ee-b50d-6cfe544c0428` game:
```bash
curl -X GET "http://127.0.0.1:24377/api/${PLAYER}/282ba89a-44b0-11ee-b50d-6cfe544c0428"
```

Analyze the `282ba89a-44b0-11ee-b50d-6cfe544c0428` game:
```bash
curl -X POST "http://127.0.0.1:24377/api/${PLAYER}/282ba89a-44b0-11ee-b50d-6cfe544c0428"
```

View the database files:
```bash
docker run -it --rm -v chess-analyzer_db-data:/var/lib/data ubuntu:jammy /bin/ls -hAlp /var/lib/data/
```

For `.devcontainer`, either clone or link the `contend` repository's `src/` dir to `.devcontainer/src/`.