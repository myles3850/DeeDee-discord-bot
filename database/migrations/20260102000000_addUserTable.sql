-- +goose Up
CREATE TABLE IF NOT EXISTS "users" (
	"id" serial NOT NULL UNIQUE,
	"discord_id" bigint NOT NULL UNIQUE,
	"discord_user" varchar(255) NOT NULL,
	PRIMARY KEY ("id")
);

CREATE TABLE IF NOT EXISTS "todo" (
	"id" serial NOT NULL UNIQUE,
	"channel_message_id" varchar(255) NOT NULL,
	"title" varchar(255) NOT NULL,
	"description" text,
	"completed" bit,
	"completed_by" bigint NOT NULL,
	PRIMARY KEY ("id")
);

CREATE TABLE IF NOT EXISTS "reminders" (
	"id" serial NOT NULL UNIQUE,
	"title" varchar(255) NOT NULL,
	"additional_info" text,
	"user_created" bigint NOT NULL,
	"users_to_tag" bigint[],
	"reminder_datetime" timestamp with time zone NOT NULL,
	PRIMARY KEY ("id")
);

DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'todo_fk5') THEN
        ALTER TABLE "todo" ADD CONSTRAINT "todo_fk5" FOREIGN KEY ("completed_by") REFERENCES "users"("id");
    END IF;
END $$;

DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'reminders_fk3') THEN
        ALTER TABLE "reminders" ADD CONSTRAINT "reminders_fk3" FOREIGN KEY ("user_created") REFERENCES "users"("id");
    END IF;
END $$;

-- +goose Down
ALTER TABLE "reminders" DROP CONSTRAINT IF EXISTS "reminders_fk3";
ALTER TABLE "todo" DROP CONSTRAINT IF EXISTS "todo_fk5";
DROP TABLE IF EXISTS "reminders";
DROP TABLE IF EXISTS "todo";
DROP TABLE IF EXISTS "users";
