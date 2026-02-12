# VanguardQ - Development Log

Daily technical progress on VanguardQ. Each entry documents changes, rationale, and design decisions for future reference.

---

## Template (Copy for future steps)

```
## YYYY-MM-DD | Step X: [Feature Name] ✓

**Components:** `path/file.go`, `path/file.go`

**Changes:** Brief 2-4 line summary of what changed

**Purpose:** Why this matters in 2-4 lines

**Notes:** Any issues, blockers, or design decisions (optional)

**References:** [ProjectArchitecture.md](#), [ProjectContacts.md](#)
```

---

## References For Anything

**Architecture & Contracts:**

- [ProjectArchitecture.md](ProjectArchitecture.md) — Locked lifecycle, Redis commands, DB schema
- [ProjectContacts.md](ProjectContacts.md) — Job struct, interfaces, API contract
- [ImplementionFlow.md](ImplementionFlow.md) — 10-step build plan

**All decisions align with locked DOCS. When in doubt, defer to DOCS as source of truth.**

## Entries

### 2026-02-10

#### Step 1: Project Skeleton ✓

**Components:** `go.mod`, `README.md`

**Changes:** Initialized Go module, created internal/cmd folder structure per [ProjectSpecs.md](ProjectSpecs.md#code-structure). Added brief README linked to detailed docs.

**Purpose:** Establish clean project layout. Clear onboarding context for future reference.

**References:** [ProjectSpecs.md](ProjectSpecs.md#code-structure)

---

#### Step 2: Contracts and Models ✓

**Components:** `internal/validators/job_state.go`, `internal/validators/queue.go`, `internal/validators/payload.go`

**Changes:** Defined 8 JobStatus constants and validators: `IsValidStatus()`, `CanTransition()` (with processing→cancelled fix), `ValidatePayload()` (256KB max, valid JSON), `ValidateQueue()` (high/default/low), `IsFinalState()`.

**Purpose:** Enforce locked rules from ProjectArchitecture.md at type level. Prevent invalid states from reaching queue/storage layers.

**Notes:** Naming inconsistency noted (ValidateQueue vs IsValidPayload) — standardize to Validate\* prefix. Ready for Step 3.

**References:** [ProjectArchitecture.md](ProjectArchitecture.md#job-lifecycle-exact-states), [ProjectArchitecture.md](ProjectArchitecture.md#anything-else-locked-decisions), [ProjectContacts.md](ProjectContacts.md#job-contract-final-fields)

---

#### Step 3: Redis Queue Adapter (In Progress) 🔄

**Components:** `internal/queue/queue.go`, `internal/queue/redis.go`

**Changes:** Defined Queue interface matching ProjectContracts.md (9 operations: Enqueue, EnqueueDelayed, Reserve, Ack, FailWithRetry, FailTerminal, MoveToDead, Cancel, Get). Initialized QueueService struct with Redis client and NewRedisClient() factory with env-based config (fallback: localhost:6379).

**Purpose:** Abstract Redis operations behind clean interface. Decouple API/scheduler/worker from Redis internals. Enable testable, swappable queue implementations per architecture design.

**Notes:** Interface is locked per ProjectContracts.md. QueueService method stubs remain—need to implement Redis command sequences (HSET, LPUSH, BRPOPLPUSH, ZADD, ZREM, etc.) per ProjectArchitecture.md. Next: serialization helpers (Job ↔ Redis hash), then Enqueue/EnqueueDelayed (simplest), then Reserve (most complex with inflight atomicity). All operations must maintain at-least-once delivery guarantee and proper state transitions.

**References:** [ProjectArchitecture.md](ProjectArchitecture.md#redis-keys-and-data-model), [ProjectArchitecture.md](ProjectArchitecture.md#redis-commands-per-step), [ProjectContacts.md](ProjectContacts.md#queue-contract-redis-abstraction), [ImplementionFlow.md](ImplementionFlow.md#3-redis-queue-adapter)

**Next Steps (Tomorrow):**

1. **Implement serialization helpers** in `redis.go`: ✅
   - `jobToHash()` — Convert Job struct to Redis hash fields (HSET args) ✅
   - `hashToJob()` — Deserialize HGETALL result back to Job struct ✅
   - Handle timestamp serialization (unix ms format) ✅

2. **Implement Enqueue & EnqueueDelayed** (simplest operations):
   - Enqueue: HSET job:{id} + LPUSH queue:{name}
   - EnqueueDelayed: HSET job:{id} + ZADD delayed {scheduledAt_ts} {id}
   - Both must set status to "queued" or "delayed" respectively

3. **Implement Get** (needed for all other operations):
   - HGETALL job:{id} → deserialize to Job

4. **Implement Reserve** (most complex):
   - BRPOPLPUSH queue:{name} queue:{name}:inflight {timeout}
   - ZADD processing:{name} {deadline_ts} {job_id}
   - LREM queue:{name}:inflight 1 {job_id}
   - Update job status to "processing" in hash
   - Handle timeout calculations correctly

5. **Implement state transitions** (Ack, FailWithRetry, FailTerminal, MoveToDead, Cancel):
   - Each must follow exact Redis command sequence from ProjectArchitecture.md
   - Maintain proper status in job hash
   - Always remove from processing:{name} when transitioning out

6. **Write unit tests** for each method (can use embedded Redis or docker for testing)

7. **Verify against ProjectArchitecture.md:**
   - All Redis command sequences must match exactly
   - All state transitions must be allowed per job_state validators
   - At-least-once delivery guarantee maintained (ZADD processing before LREM inflight)

**Blockers/Notes:** No blockers. Validators already in place. Just follow the locked commands in ProjectArchitecture.md step-by-step.

```

```
