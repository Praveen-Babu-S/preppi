-- Question service owns the questions table
CREATE TABLE questions (
    id          BIGSERIAL   PRIMARY KEY,
    student_id  BIGINT      NOT NULL,
    assignee_id BIGINT,
    subject     VARCHAR(100) NOT NULL,
    topic       VARCHAR(100),
    description TEXT        NOT NULL,
    image_urls  TEXT,
    urgency     VARCHAR(20)  NOT NULL DEFAULT 'normal' CHECK (urgency IN ('low', 'normal', 'urgent')),
    status      VARCHAR(20)  NOT NULL DEFAULT 'open' CHECK (status IN ('open', 'assigned', 'in_progress', 'answered', 'escalated')),
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    deleted_at  TIMESTAMPTZ
);

CREATE INDEX idx_questions_student_status ON questions(student_id, status);
CREATE INDEX idx_questions_status ON questions(status);
CREATE INDEX idx_questions_subject ON questions(subject);
CREATE INDEX idx_questions_assignee ON questions(assignee_id);
CREATE INDEX idx_questions_created ON questions(created_at);
