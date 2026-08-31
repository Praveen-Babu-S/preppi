-- User service owns profiles and mentor_profiles tables
CREATE TABLE profiles (
    id          BIGSERIAL   PRIMARY KEY,
    user_id     BIGINT      NOT NULL,
    name        VARCHAR(255) NOT NULL,
    avatar_url  VARCHAR(500),
    phone       VARCHAR(20),
    school      VARCHAR(255),
    college     VARCHAR(255),
    bio         TEXT,
    role        VARCHAR(20)  NOT NULL CHECK (role IN ('student', 'mentor', 'admin')),
    online      BOOLEAN      NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    deleted_at  TIMESTAMPTZ
);

CREATE UNIQUE INDEX idx_profiles_user_id ON profiles(user_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_profiles_role ON profiles(role);
CREATE INDEX idx_profiles_online ON profiles(online);

CREATE TABLE mentor_profiles (
    id                   BIGSERIAL   PRIMARY KEY,
    user_id              BIGINT      NOT NULL,
    expertise_subjects   TEXT,
    sub_topics           TEXT,
    verification_status  VARCHAR(20) NOT NULL DEFAULT 'pending',
    rating               NUMERIC(3,2) DEFAULT 0,
    questions_answered   INTEGER     NOT NULL DEFAULT 0,
    created_at           TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at           TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX idx_mentor_profiles_user_id ON mentor_profiles(user_id);
CREATE INDEX idx_mentor_profiles_verification ON mentor_profiles(verification_status);
