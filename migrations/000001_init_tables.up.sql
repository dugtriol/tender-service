BEGIN;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS employee
(
    id         UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    username   VARCHAR(50) UNIQUE NOT NULL,
    first_name VARCHAR(50),
    last_name  VARCHAR(50),
    created_at TIMESTAMP        DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP        DEFAULT CURRENT_TIMESTAMP
);

DROP TYPE IF EXISTS organization_type CASCADE;
CREATE TYPE organization_type AS ENUM (
    'IE',
    'LLC',
    'JSC'
    );

CREATE TABLE IF NOT EXISTS organization
(
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name        VARCHAR(100) NOT NULL,
    description TEXT,
    type        organization_type,
    created_at  TIMESTAMP        DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP        DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS organization_responsible
(
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    organization_id UUID REFERENCES organization (id) ON DELETE CASCADE,
    user_id         UUID REFERENCES employee (id) ON DELETE CASCADE
);

DROP TYPE IF EXISTS service_type CASCADE;
CREATE TYPE service_type AS ENUM (
    'Construction',
    'Delivery',
    'Manufacture'
    );

DROP TYPE IF EXISTS tender_status CASCADE;
CREATE TYPE tender_status AS ENUM (
    'Created',
    'Published',
    'Closed'
    );

CREATE TABLE IF NOT EXISTS tender
(
    id               UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name             VARCHAR(100) NOT NULL,
    description      TEXT,
    type             service_type,
    status           tender_status DEFAULT 'Created',
    organization_id  UUID REFERENCES organization (id) ON DELETE CASCADE,
    version          INT              DEFAULT 1,
    created_at       TIMESTAMP        DEFAULT CURRENT_TIMESTAMP,
    creator_username VARCHAR(50) REFERENCES employee (username) ON DELETE CASCADE
);

DROP TYPE IF EXISTS bid_status CASCADE;
CREATE TYPE bid_status AS ENUM (
    'Created',
    'Published',
    'Canceled',
    'Approved',
    'Rejected'
    );

DROP TYPE IF EXISTS author_type CASCADE;
CREATE TYPE author_type AS ENUM (
    'Organization',
    'User'
    );

CREATE TABLE IF NOT EXISTS bid
(
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name        VARCHAR(100) NOT NULL,
    description TEXT,
    status      bid_status DEFAULT 'Created',
    tender_id   UUID REFERENCES tender (id) ON DELETE CASCADE,
    author_type author_type,
    author_id   UUID REFERENCES employee (id) ON DELETE CASCADE,
    version     INT              DEFAULT 1,
    created_at  TIMESTAMP        DEFAULT CURRENT_TIMESTAMP
);
CREATE TABLE IF NOT EXISTS bid_review
(
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    description TEXT,
    bid_id      UUID REFERENCES bid (id) ON DELETE CASCADE,
    created_at  TIMESTAMP        DEFAULT CURRENT_TIMESTAMP
);

COMMIT;