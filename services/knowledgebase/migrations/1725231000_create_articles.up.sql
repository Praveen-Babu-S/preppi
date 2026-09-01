-- KnowledgeBase service owns articles and topics tables
CREATE TABLE kb_articles (
    id           BIGSERIAL   PRIMARY KEY,
    subject      VARCHAR(100) NOT NULL,
    topic        VARCHAR(100) NOT NULL,
    title        VARCHAR(255) NOT NULL,
    summary      TEXT,
    full_text    TEXT,
    upvotes      INTEGER     NOT NULL DEFAULT 0,
    tags         TEXT,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at   TIMESTAMPTZ
);

CREATE INDEX idx_kb_articles_subject ON kb_articles(subject);
CREATE INDEX idx_kb_articles_topic ON kb_articles(topic);
CREATE INDEX idx_kb_articles_search ON kb_articles USING GIN(to_tsvector('english', title || ' ' || COALESCE(summary, '')));

CREATE TABLE kb_topics (
    id           BIGSERIAL   PRIMARY KEY,
    name         VARCHAR(100) NOT NULL,
    subject      VARCHAR(100) NOT NULL,
    article_count INTEGER    NOT NULL DEFAULT 0,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX idx_kb_topics_name_subject ON kb_topics(name, subject);
