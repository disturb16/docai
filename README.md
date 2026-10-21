# DocAI

Self-hosted document management. Upload documents (PDF, DOCX, ODT, MD, TXT,
XLSX, PNG, JPG), extract content via OCR/vision AI, and get generated metadata
(title, summary, category, tags, entities, language). Single user, local, no
auth (MVP).

## Requirements

- Go 1.27+
- [Deno](https://deno.com) (frontend build)
- `lit` CLI (liteparse) for PDF/DOCX/XLSX extraction
- One AI backend: llama.cpp `llama-server` on `:42951` (default) or Ollama on
  `:11434`

## Quick start

```sh
# build frontend
cd frontend && deno task build && cd ..

# run
go run . serve
```

Open http://localhost:8090.

## Configuration

| Env                      | Default       | Description                          |
| ------------------------ | ------------- | ------------------------------------ |
| `DOCAI_BACKEND`          | `llamacpp`    | `llamacpp` or `ollama`               |
| `DOCAI_VISION_MODEL`     | `Gemma-4-E4B` | Vision/OCR model                     |
| `DOCAI_CLASSIFY_MODEL`   | `Gemma-4-E4B` | Classification model                 |
| `DOCAI_EMBED_MODEL`      | *(empty)*     | Embedding model; enables semantic search |
| `DOCAI_CONCURRENCY`      | `1`           | Parallel document processing         |

Backend-specific overrides: `DOCAI_LLAMA_VISION_MODEL`,
`DOCAI_LLAMA_CLASSIFY_MODEL`, `DOCAI_LLAMA_EMBED_MODEL`,
`DOCAI_OLLAMA_VISION_MODEL`, `DOCAI_OLLAMA_CLASSIFY_MODEL`,
`DOCAI_OLLAMA_EMBED_MODEL` (take precedence over the shared vars). With
llama.cpp, `DOCAI_LLAMA_EMBED_URL` points at a second `llama-server` running the
embedding model (one model per process); Ollama serves everything from one URL.

```sh
# example: semantic search with EmbeddingGemma 2
# (raise the batch/context flags: llama-server defaults 512-token batches,
# too small for document chunks)
llama-server -hf ggml-org/embeddinggemma-2-GGUF:BF16 --embedding \
  -c 8192 -b 8192 -ub 8192 --port 42952
DOCAI_LLAMA_EMBED_MODEL=embeddinggemma-2 DOCAI_LLAMA_EMBED_URL=http://127.0.0.1:42952 go run . serve
# or with Ollama
ollama pull embeddinggemma-2
DOCAI_BACKEND=ollama DOCAI_OLLAMA_EMBED_MODEL=embeddinggemma-2 go run . serve
```

Existing documents can be embedded from the Settings page ("Backfill
Embeddings"); new uploads embed as part of processing. Switching embedding
models? Run the backfill again — vectors tagged with another model are
re-embedded. Search then returns keyword hits plus a "Semantically similar"
list, and each document page has a "Find Similar" button.

## Commands

```sh
go run . serve                 # dev server
go build -o pocketbase .       # build
go test ./... && go vet ./...  # checks

cd frontend
deno task dev                  # frontend dev server
deno task build                # build to frontend/pb_public/
deno fmt && deno lint && deno check src/main.ts
```

## Docs

- `PRD.md` — product requirements
- `DESIGN.md` — design system
- `AGENTS.md` — architecture and conventions
