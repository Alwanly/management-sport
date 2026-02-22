-- Create "users" table
CREATE TABLE IF NOT EXISTS "users" (
	"id" character varying(36) NOT NULL,
	"username" character varying(255) NOT NULL,
	"password" character varying(255) NOT NULL,
	"role" character varying(50) NOT NULL DEFAULT 'user',
	"created_at" timestamptz NOT NULL,
	"updated_at" timestamptz NOT NULL,
	PRIMARY KEY ("id")
);

CREATE UNIQUE INDEX IF NOT EXISTS "idx_users_username" ON "users"("username");
