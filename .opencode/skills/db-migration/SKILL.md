---
name: db-migration
description: Use when creating, modifying, or altering the database schema for preppi services. Trigger on requests like "add a table", "modify the users schema", "create a migration", "add a column to X". Generates forward and rollback SQL migrations following the golang-migrate convention.
---

# Database Migration

## Purpose

Create and manage database schema changes for preppi services using golang-migrate style
SQL migrations. Each service owns its own database and schema.

## When to Use

Use this skill when the user asks to:

- Add a new table
- Add/remove/modify columns
- Add indexes or constraints
- Create a new migration file
- Change the data model for a service

## Migration Location & Naming

Migrations live in `services/<name>/migrations/`. Naming convention:

```
<timestamp>_<snake_case_name>.up.sql
<timestamp>_<snake_case_name>.down.sql
```

Example:

```
1725230400_create_questions.up.sql
1725230400_create_questions.down.sql
```

The timestamp is Unix epoch seconds (of "now"). Use `date +%s` to generate it.

## Migration File Template

**UP migration** (`*_create_questions.up.sql`):

```sql
CREATE TABLE questions (
    id            BIGSERIAL PRIMARY KEY,
    student_id    BIGINT       NOT NULL REFERENCES users(id),
    assignee_id   BIGINT       REFERENCES users(id),
    subject       VARCHAR(100) NOT NULL,
    topic         VARCHAR(100),
    description   TEXT         NOT NULL,
    status        VARCHAR(20)  NOT NULL DEFAULT 'open',  -- see enum convention below
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT now(),
    deleted_at    TIMESTAMPTZ
);

CREATE INDEX idx_questions_student_id ON questions(student_id);
CREATE INDEX idx_questions_status ON questions(status);
```

**DOWN migration** (`*_create_questions.down.sql`):

```sql
DROP TABLE IF EXISTS questions;
```

## Conventions

### Table Conventions

- Always include `id BIGSERIAL PRIMARY KEY`
- Always include `created_at TIMESTAMPTZ NOT NULL DEFAULT now()`
- Always include `updated_at TIMESTAMPTZ NOT NULL DEFAULT now()`
- Soft delete: `deleted_at TIMESTAMPTZ` (NULL = active)
- Use `BIGSERIAL` for PKs (matches GORM `uint`/`uint64`)
- Use `TEXT` for long content, `VARCHAR(n)` for short strings
- Use `TIMESTAMPTZ` for all timestamps (never `TIMESTAMP`)

### Foreign Keys

- Always explicit, named constraints: `CONSTRAINT fk_questions_student FOREIGN KEY (student_id) REFERENCES users(id)`
- Prefix FK names with `fk_`
- Index every FK column

### Indexes

- Prefix index names with `idx_`
- Index all FK columns
- Composite indexes for common query patterns (e.g., `(student_id, status)`)
- Consider partial indexes for filtered queries (e.g., `WHERE status = 'open'`)

### Status Columns

Match the proto enum values (snake_case of the protobuf enum value):

- `open`, `assigned`, `in_progress`, `answered`, `escalated`
- Use `VARCHAR(20)` with `NOT NULL DEFAULT '<first_enum_value>'`

### Enums

Prefer `VARCHAR` over Postgres `ENUM` types — easier to migrate later. Only use `ENUM`
if values are immutable.

## Rules

- Every schema change requires a **paired up/down** migration
- Up and down migrations must be exact inverses
- **Never** use GORM `AutoMigrate` in production — only migrations
- Never modify an already-applied migration — create a new one
- Down migrations must be safe to run (guarded, no destructive surprises)
- Migration files are committed to the repo

## Quick Reference Templates

### Add a column

```sql
-- up
ALTER TABLE questions ADD COLUMN urgency VARCHAR(10) NOT NULL DEFAULT 'normal';
-- down
ALTER TABLE questions DROP COLUMN urgency;
```

### Add an index

```sql
-- up
CREATE INDEX idx_questions_assignee_status ON questions(assignee_id, status);
-- down
DROP INDEX IF EXISTS idx_questions_assignee_status;
```

### Add a column with FK

```sql
-- up
ALTER TABLE solutions ADD COLUMN mentor_id BIGINT NOT NULL;
ALTER TABLE solutions ADD CONSTRAINT fk_solutions_mentor FOREIGN KEY (mentor_id) REFERENCES users(id);
CREATE INDEX idx_solutions_mentor_id ON solutions(mentor_id);
-- down
ALTER TABLE solutions DROP CONSTRAINT fk_solutions_mentor;
ALTER TABLE solutions DROP COLUMN mentor_id;
```

### Create a new schema (per-service database)

Migrations in `services/<name>/migrations/` operate on that service's own database.

## After Creating

After writing migrations, remind the user to run:

```
make migrate-db
```

or the service-specific migration command.
