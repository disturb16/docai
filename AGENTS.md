# AGENTS.md

## Project

DocAI — self-hosted document management app. Upload documents (PDF, DOCX, ODT,
MD, TXT, XLSX, PNG, JPG), extract content via OCR/vision AI, generate
classification metadata (title, summary, category, tags, entities, language),
browse and view results. Single user, local, no auth (MVP).

- Specs: `PRD.md` (requirements), `DESIGN.md` (design system)

## Stack

- PocketBase as a Go framework — single binary (`main.go` + `internal/`), serves
  API, files, and the SPA
- Frontend: Deno + Vite + Vue 3 (Composition API, `<script setup lang="ts">`),
  Tailwind 4, lives in `frontend/`, built to `frontend/pb_public/`
- AI backend: llama.cpp `llama-server` (`:42951`, OpenAI-compatible
  `/v1/chat/completions`) by default, or Ollama (`:11434`) — set
  `DOCAI_BACKEND=ollama`
- Models via env: `DOCAI_OLLAMA_VISION_MODEL`, `DOCAI_OLLAMA_CLASSIFY_MODEL`,
  `DOCAI_LLAMA_VISION_MODEL`, `DOCAI_LLAMA_CLASSIFY_MODEL` (shared
  `DOCAI_VISION_MODEL`/`DOCAI_CLASSIFY_MODEL` still work as fallbacks; default
  `Gemma-4-E4B`); concurrency via `DOCAI_CONCURRENCY` (default 1)
- Semantic search (optional): `DOCAI_*_EMBED_MODEL` (+ `DOCAI_*_EMBED_URL` for a
  second llama-server) points at an embedding model (e.g. EmbeddingGemma via
  `/v1/embeddings`); empty disables it
- `lit` CLI (liteparse) — PDF/DOCX/XLSX text extraction + page screenshots
  (calls libreoffice internally)

## Architecture

```
Browser → pocketbase (single Go binary, :8090)
            ├─ frontend/         Vue 3 SPA (Vite via Deno), built to frontend/pb_public/
            ├─ main.go           app bootstrap, SPA catch-all route, cron queue
            ├─ config/           env config (backend, models, concurrency)
            ├─ internal/queue    in-memory queue, bounded concurrency, boot recovery
            ├─ internal/pipeline lit CLI (os/exec) → vision/classify (net/http)
            ├─ internal/classify strict-JSON metadata parsing
            └─ pb_data/          SQLite + files (documents collection)
```

Processing lives in the PocketBase process (cron ticker + in-memory queue,
concurrency from `DOCAI_CONCURRENCY`). Status persisted in PocketBase;
`pending`/`processing` records re-queued on startup.

## Data model

`documents` collection (created via `pb_migrations/`): `file`, `filename`,
`mime`, `size`, `status` (`pending`/`processing`/`done`/`error`), `error`,
`ocr_text`, `title`, `summary`, `category`, `tags` (json), `entities` (json),
`language`, `model`, `folder` (relation to `folders`, empty = root).

`embeddings` collection: `document` (unique relation to `documents`,
cascadeDelete), `vector` (json, mean-pooled 768d; absent = keyword-only),
`model` (which model produced the vector — vectors from a different model are
ignored in search). Server-only (no API rules); written via
`pipeline.StoreEmbedding`, which deletes the row when embedding is disabled or
failed so no stale vector survives a reprocess.

`folders` collection: `name`, `parent` (self relation, empty = root). Both
relations use `cascadeDelete`, so deleting a folder removes its descendants
and their documents (UI confirms first).

## Processing pipeline

1. Convert by mime: PDF/DOCX/ODT/XLSX → `lit parse --no-ocr`; if text < ~100
   chars → `lit screenshot` (cap ~10 pages) → vision per page; MD/TXT → read
   directly; PNG/JPG → vision directly
2. OCR/vision: vision model, prompt "transcribe all text"
3. Classify: classify model, strict JSON
   `{title, summary, category, tags, entities, language}`; fallback to raw text
   on parse failure
4. Embed (optional): chunk OCR text, embed each chunk via `/v1/embeddings`
   (EmbeddingGemma query/document prefixes, added client-side), mean-pool +
   normalize; failures are non-fatal
5. Persist → `done` (or `error` with message)

## UI

- `/` — list: folder browser (breadcrumb via `?folder=<id>`, create/rename/
  delete folders), status badges, filter, upload, delete, realtime updates
- `/upload` — multi-file upload with per-file progress + destination folder
- `/activity` — processing monitor (current doc + queue)
- `/documents/:id` — detail: metadata panel + OCR text (markdown rendered for
  MD) + image/PDF preview + download original + tag editing + move to folder +
  reprocess + find similar documents
- `/search` — keyword search + semantic "Semantically similar" section
  (`/api/search`)
- Components: `DocumentItem`, `StatusBadge`
- Composables: `useDocuments` (list/upload/subscribe/remove), `useDocument`
  (single fetch/subscribe/update/find-similar), `useFolders` (folder CRUD/
  subscribe), `useSettings` (config + health + embedding backfill)
- Folder tree helpers (pure, tested): `frontend/src/lib/folders.ts`
- Upload/download/filter via native PocketBase API — custom routes are
  `/api/search`, `/api/documents/{id}/similar`, `/api/docai/settings`,
  `/api/docai/settings/health`, `/api/docai/analytics`,
  `/api/docai/embeddings/backfill`, `/api/documents/{id}/embed`, and the SPA
  catch-all

## Commands

- `go run . serve` — dev server
- `go build -o pocketbase .` — build
- `go test ./...`, `go vet ./...` — checks
- `deno task dev` / `deno task build` — frontend dev / build (run from `frontend/`)
- `deno fmt`, `deno lint`, `deno check src/main.ts` — frontend checks (run from `frontend/`)

## Tooling

- Go: never edit `go.mod`/`go.sum` by hand — use `go get`, `go mod tidy`,
  `go mod init`, etc.
- Frontend (Deno): never edit `deno.json`/`package.json` by hand — use
  `deno add`, `deno install`, `deno run`, `deno task`, etc.

## Conventions

- TypeScript everywhere; Composition API with `<script setup lang="ts">`
- Design tokens from `DESIGN.md` frontmatter (colors, typography, spacing);
  status colors: pending `#94A3B8`, processing `#3B82F6`, done `#10B981`, error
  `#EF4444`
- No comments unless they explain a deliberate shortcut (`ponytail:` prefix)
- Keep it lazy: stdlib/native first, no new deps for what a few lines can do
