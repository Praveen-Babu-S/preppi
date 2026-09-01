-- Solution service owns solutions and follow_ups tables
CREATE TABLE solutions (
    id           BIGSERIAL   PRIMARY KEY,
    question_id  BIGINT      NOT NULL,
    mentor_id    BIGINT      NOT NULL,
    description  TEXT        NOT NULL,
    image_urls   TEXT,
    upvotes      INTEGER     NOT NULL DEFAULT 0,
    downvotes    INTEGER     NOT NULL DEFAULT 0,
    is_accepted  BOOLEAN     NOT NULL DEFAULT FALSE,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at   TIMESTAMPTZ
);

CREATE INDEX idx_solutions_question ON solutions(question_id);
CREATE INDEX idx_solutions_mentor ON solutions(mentor_id);
CREATE INDEX idx_solutions_accepted ON solutions(is_accepted) WHERE is_accepted = TRUE;

CREATE TABLE follow_ups (
    id           BIGSERIAL   PRIMARY KEY,
    solution_id  BIGINT      NOT NULL,
    user_id      BIGINT      NOT NULL,
    message      TEXT        NOT NULL,
    image_urls   TEXT,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_follow_ups_solution ON follow_ups(solution_id);
