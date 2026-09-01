-- The doubt service owns all tables for questions, solutions, chat, and notifications.
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

CREATE TABLE chat_rooms (
    id           BIGSERIAL   PRIMARY KEY,
    question_id  BIGINT      NOT NULL,
    student_id   BIGINT      NOT NULL,
    mentor_id    BIGINT      NOT NULL,
    status       VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX idx_chat_rooms_question ON chat_rooms(question_id);
CREATE INDEX idx_chat_rooms_student ON chat_rooms(student_id);
CREATE INDEX idx_chat_rooms_mentor ON chat_rooms(mentor_id);

CREATE TABLE messages (
    id           BIGSERIAL   PRIMARY KEY,
    room_id      BIGINT      NOT NULL,
    sender_id    BIGINT      NOT NULL,
    content      TEXT        NOT NULL,
    type         VARCHAR(20) NOT NULL DEFAULT 'text',
    image_url    VARCHAR(500),
    read         BOOLEAN     NOT NULL DEFAULT FALSE,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_messages_room ON messages(room_id, created_at);
CREATE INDEX idx_messages_sender ON messages(sender_id);

CREATE TABLE notifications (
    id           BIGSERIAL   PRIMARY KEY,
    user_id      BIGINT      NOT NULL,
    type         VARCHAR(50) NOT NULL,
    title        VARCHAR(255) NOT NULL,
    body         TEXT        NOT NULL,
    data         JSONB,
    channels     TEXT,
    read         BOOLEAN     NOT NULL DEFAULT FALSE,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_notifications_user ON notifications(user_id, created_at DESC);
CREATE INDEX idx_notifications_user_read ON notifications(user_id, read);

CREATE TABLE notification_preferences (
    id              BIGSERIAL   PRIMARY KEY,
    user_id         BIGINT      NOT NULL,
    in_app_enabled  BOOLEAN     NOT NULL DEFAULT TRUE,
    push_enabled    BOOLEAN     NOT NULL DEFAULT TRUE,
    email_enabled   BOOLEAN     NOT NULL DEFAULT FALSE,
    sms_enabled     BOOLEAN     NOT NULL DEFAULT FALSE,
    digest_mode     BOOLEAN     NOT NULL DEFAULT FALSE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX idx_notification_prefs_user ON notification_preferences(user_id);
