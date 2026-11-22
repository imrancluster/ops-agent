# M07: RAG (Retrieval-Augmented Generation)

**Guide:** §6 step 5 (part b) · **Week:** 6 · **Depends on:** M04, M06

## Concept

The model doesn't know your docs, and you can't paste them all into every prompt. RAG fixes this:

1. **Chunk** the docs into pieces about one section long, keeping each section's title.
2. **Embed** each chunk. An embedding model turns text into a vector (a list of floats), and texts with similar meaning get vectors that point in similar directions.
3. **Store** the vectors (SQLite is fine for a few thousand chunks).
4. At question time, embed the query and find the **top-k most similar** chunks with **cosine similarity**.
5. Return them as the `search_docs` tool result, **with their source names**, so the answer can cite them.

Making RAG a *tool* (agentic RAG), rather than always stuffing chunks into the prompt, lets the model decide **when** it needs the docs and **what** to search for.

## How industry does it

- **Indexing is offline** (a separate `cmd/index` job) and **retrieval happens online**. Re-index whenever the docs change.
- **Retrieval quality is measured**, for example with recall@k on a set of question → expected-chunk pairs. Bad retrieval leads to confident wrong answers (hallucination with citations).
- **Hybrid search:** keyword search (BM25 / SQLite FTS5) plus vector search usually beats either one alone.
- **Scaling:** brute force works up to a few thousand chunks. Beyond that, use a vector index (HNSW) or pgvector.
- **Bangla warning (guide):** many embedding models are mostly trained on English. For Bangla docs, test a multilingual model such as bge-m3.

## Go you'll use

- **`[]float32` math:** dot product and norms, using `float64` accumulators for precision.
- **Encoding vectors in SQLite:** `encoding/binary` into a `BLOB`.
- **`container/heap`** or sort-then-slice for top-k.
- **Benchmarks:** `func BenchmarkSearch(b *testing.B)` measures how many chunks brute force can handle.

## Tasks

### Task 1: Cosine similarity and top-k
- **Claude scaffolds:** `internal/memory/vector.go` signatures and table tests (identical vectors give 1, orthogonal give 0, a zero vector, mismatched lengths give an error).
- **You implement:** `Cosine`, `TopK`, and a benchmark with 5,000 chunks.

### Task 2: Embedder
- **You implement:** `Embedder` interface plus `OllamaEmbedder` (`/api/embed`, `nomic-embed-text`), with an `httptest` contract test. Use an injected `*http.Client`, not `http.DefaultClient`.

### Task 3: Chunker
- **You implement:** split Markdown by headings, keeping the heading path (`Guide > Step 3 > Agent Loop`) as chunk metadata. Cap each chunk's size, with a little overlap between chunks.

### Task 4: Index job and the `search_docs` tool
- **Claude scaffolds:** `cmd/index` and the `migrations/0002_chunks.sql` signature.
- **You implement:** index `GO-AI-AGENT.md` itself, then build the `search_docs` tool, which returns the top 4 chunks with their sources.
- **Try it:** ask the agent "Why must tool_use and tool_result never be separated?"

### Task 5: Retrieval eval
- **You implement:** `evals/retrieval.json` with 10 questions → expected heading. Compute recall@4. Then try `bge-m3` (or another multilingual model) on Bangla queries and compare.

## Check yourself

1. Why does cosine similarity measure *direction* and not *length*?
2. Why expose RAG as a tool instead of always adding chunks to the prompt?
3. Retrieval returns the wrong chunk, and the model then answers confidently. Whose fault is it, and how do you detect it?
4. When would you replace SQLite brute force with pgvector?
