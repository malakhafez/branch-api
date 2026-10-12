CREATE TABLE branches (
    id                  UUID PRIMARY KEY,
    name                VARCHAR(100)  NOT NULL UNIQUE,
    type                VARCHAR(20)   NOT NULL
                        CHECK (type IN ('branch', 'central_kitchen')),
    address             TEXT,
    official_start_time TIME          NOT NULL,
    late_rate           NUMERIC(10,2) NOT NULL
                        CHECK (late_rate >= 0),
    is_active           BOOLEAN       NOT NULL DEFAULT TRUE,
    created_at          TIMESTAMPTZ   NOT NULL,
    updated_at          TIMESTAMPTZ   NOT NULL
);