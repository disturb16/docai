# DocAI — Product Requirements Document (PRD)

## 1. Overview

DocAI is a self-hosted document management app. Users upload documents (PDF,
DOCX, ODT, MD, TXT, XLSX, images), the system extracts their content via
OCR/vision AI, and generates metadata (title, summary, category, tags, entities,
language) for classification. Users can browse processed documents and view
content + metadata.

**MVP goal:** upload all files → process them → view processed files with
content + generated metadata.

## 2. Users & Use Cases

Single user, local deployment (no auth in MVP).

| #    | Use case                                        | Priority |
| ---- | ----------------------------------------------- | -------- |
| UC-1 | Upload one or many documents at once            | P0       |
| UC-2 | System extracts text content from each document | P0       |
| UC-3 | System generates classification metadata via AI | P0       |
| UC-4 | Browse list of documents with processing status | P0       |
| UC-5 | View a document's content + metadata            | P0       |
| UC-6 | Download the original file                      | P1       |
| UC-7 | Filter documents by status                      | P1       |

## 3. Functional Requirements

### FR-1 Upload

- Multi-file upload via web UI.
- Accepted types: PDF, DOCX, ODT, MD, TXT, XLSX, PNG, JPG.
- Each upload creates a document record with status `pending`.

### FR-2 Processing

- Runs automatically after upload, concurrency configurable
  (`DOCAI_CONCURRENCY`, default 1).
- Pipeline per file type:
  - PDF/DOCX/ODT/XLSX: `lit parse --no-ocr`; if extracted text < ~100 chars,
    render pages to images (`lit screenshot`, cap 10 pages) → vision OCR.
  - MD/TXT: read text directly.
  - PNG/JPG: vision OCR directly.
- AI backend: local Ollama or llama.cpp `llama-server` (OpenAI-compatible),
  selected via `DOCAI_BACKEND` (default `llamacpp`).
- Vision and classification models configurable via env (`DOCAI_VISION_MODEL`,
  `DOCAI_CLASSIFY_MODEL`, default `Gemma-4-E4B`).
- Generated metadata: `title`, `summary`, `category`, `tags`, `entities`,
  `language`.
- Status transitions: `pending` → `processing` → `done` | `error`.
- On failure, record error message; document remains viewable with status
  `error`.

### FR-3 Viewing

- Routes: `/` (list), `/upload` (multi-file upload), `/activity` (processing
  monitor), `/documents/:id` (detail).
- List view: filename, size, status badge, generated title/category when
  available; filter by status; delete.
- Detail view: metadata panel, extracted text (markdown rendered for MD),
  image/PDF preview, download link for original, tag editing, reprocess.
- Lists auto-refresh via realtime subscription.

### FR-4 Persistence

- All data in PocketBase (files + metadata).
- Processing status survives app restarts (recovery: `pending`/`processing`
  records are re-queued on startup).

## 4. Non-Functional Requirements

| Area        | Requirement                                                                |
| ----------- | -------------------------------------------------------------------------- |
| Privacy     | All processing local (Ollama/llama.cpp); no data leaves the machine        |
| Performance | Processing is async; UI stays responsive; concurrency configurable (`DOCAI_CONCURRENCY`, default 1) |
| Reliability | Failed processing never loses the original file; errors surfaced in UI     |
| Scalability | MVP: single user, thousands of documents; PDF OCR capped at 10 pages/doc   |
| Tech        | Vue 3 SPA (Vite via Deno) + PocketBase (Go framework, single binary) + Ollama/llama.cpp; TypeScript |

## 5. Out of Scope (MVP)

- Authentication / multi-user
- Full-text search
- Cloud AI providers
- Persistent job queue / horizontal scaling
- PDF OCR beyond 10 pages per document
- XLSX cell-level parsing (rendered-PDF OCR covers it)
- Editing title/summary/category in UI (tags are editable)

## 6. Success Criteria

- Upload a mixed batch of files (PDF, DOCX, MD, image) → all reach `done` with
  content + metadata.
- Scanned (image-only) PDF yields readable OCR text.
- Metadata JSON parses correctly for ≥90% of documents (fallback to raw text
  otherwise).
- Restarting the app mid-processing resumes pending documents.

## 7. Future (post-MVP)

- Auth, search, full metadata editing, cloud AI option, better queue,
  full-page OCR, per-cell spreadsheet parsing.
