# Scheduling-API

**Scheduling-API** is a simple, lightweight social media scheduling tool.  
It currently supports **Instagram**, **YouTube**, and **TikTok** platforms.

You can use it directly via the REST API or integrate it with the official [frontend interface](https://github.com/maheshrc27/schedulingapi-ui) available on GitHub.

---

## Run Locally

### Requirements:
- [Go (Golang)](https://go.dev/)
- [Redis] (https://redis.io/docs/latest/operate/redisinsight/install/install-on-desktop/)

### Steps:
1. Clone the repository
2. Copy `.env.example` to `.env` and fill in your values
3. Run the server:

```bash
go mod tidy
go run ./cmd/server
```


```bash
# Replace these variables with your actual values
API_KEY="YOUR_API_KEY"
ENDPOINT="https://api.scheduling.com/posts/create"
FILE_PATH="/path/to/your/media/file"
SCHEDULING_TIME="2025-01-05T10:00"
CAPTION="Your post caption here"
TITLE="Optional title for the post"
SELECTED_ACCOUNTS="[21]"

# Perform the API request
curl -X POST "$ENDPOINT?api_key=$API_KEY" \
  -H "Content-Type: multipart/form-data" \
  -F "files=@$FILE_PATH" \
  -F "scheduling_time=$SCHEDULING_TIME" \
  -F "caption=$CAPTION" \
  -F "title=$TITLE" \
  -F "selected_counts=$SELECTED_ACCOUNTS"

```