## Introduction
Large language model applications currently face issues such as lack of long-term memory, insufficient domain knowledge, and limited real-time capabilities. The core of these problems is that LLMs cannot independently acquire or record certain types of knowledge, such as:
- Enterprise proprietary knowledge
- Memory data, including user behaviors, feedback, preferences, etc.
These can collectively be referred to as "knowledge." Current solutions like RAG and Memory Systems aim to supplement various types of knowledge, but integration is complex and involves significant redundant work.

To address these challenges, ino was created, introducing KAG (Knowledge-Augmented Generation) or the Unified Retrieval Framework. The core idea of this system is to treat all information sources that could benefit LLMs (external documents, conversation history, user profiles, etc.) as retrievable "knowledge," and establish a unified framework to intelligently retrieve, filter, and integrate this knowledge, ultimately injecting it into prompts in an optimized manner.

## Development Guide
> Note: Docker installation is required

### Start Dependent Services
```bash
make services-up
```

### Run ino
```bash
make dev
```

## How to Use

### Writing Knowledge

Use the `POST /api/v1/openapi/collect` endpoint to write knowledge.

**Request Example**
```bash
curl --location 'http://localhost:8080/api/v1/openapi/collect' \
--header 'user-key: <YOUR_USER_KEY>' \
--header 'Content-Type: application/json' \
--data '{
    "content": "The quick brown fox jumps over the lazy dog."
}'
```
Or write knowledge via a link:
```bash
curl --location 'http://localhost:8080/api/v1/openapi/collect' \
--header 'user-key: <YOUR_USER_KEY>' \
--header 'Content-Type: application/json' \
--data '{
    "content_link": "https://en.wikipedia.org/wiki/Fox"
}'
```

**Parameters**
- `user-key`: (Header) Optional, identifies the user.
- `collection-key`: (Header) Optional, identifies the collection.
- `content`: (Body) Knowledge content. Either `content` or `content_link` is required.
- `content_link`: (Body) Link to knowledge content.


### Reading Knowledge

Use the `POST /api/v1/openapi/search` endpoint to retrieve knowledge.

**Request Example**
```bash
curl --location 'http://localhost:8080/api/v1/openapi/search' \
--header 'user-key: <YOUR_USER_KEY>' \
--header 'Content-Type: application/json' \
--data '{
    "query": "What does the fox say?"
}'
```

**Parameters**
- `user-key`: (Header) Optional, filters by user.
- `collection-key`: (Header) Optional, filters by collection.
- `query`: (Body) Required, the query question.
- `query_strategy`: (Body) Optional, query strategy: `quick` (default) or `agent`.


> more docs your can see in `./docs`