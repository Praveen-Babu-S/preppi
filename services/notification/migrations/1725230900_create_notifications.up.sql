-- Notification service owns notifications and notification_preferences tables
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
