-- 1_create_user_table.up.sql
BEGIN;

CREATE TABLE "users"
(
    "id" SERIAL PRIMARY KEY,
    "address" character(42) NOT NULL UNIQUE,
    "total_points" numeric(12, 3) NOT NULL DEFAULT 0,
    "created_at" timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE "tokens"
(
    "id" character(42) PRIMARY KEY,
    "name" character varying(255) NOT NULL,
    "symbol" character varying(10) NOT NULL,
    "decimals" integer NOT NULL,
    "created_at" timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE "swap_history"
(
    "id" SERIAL PRIMARY KEY,
    "token" character(42) NOT NULL,
    "account" character(42) NOT NULL,
    "transaction_hash" character(66) NOT NULL,
    "usd_value" numeric(20, 6) NOT NULL,
    "last_updated" timestamp with time zone NOT NULL,
    "created_at" timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP
);


CREATE TABLE "points_history"
(
    "id" SERIAL PRIMARY KEY,
    "token" character(42) NOT NULL,
    "account" character(42) NOT NULL,
    "points" numeric(12, 3) NOT NULL,
    "description" character varying(255) NOT NULL,
    "created_at" timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP
);


COMMIT;