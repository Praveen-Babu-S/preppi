-- Chat service owns chat_rooms and messages tables
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
