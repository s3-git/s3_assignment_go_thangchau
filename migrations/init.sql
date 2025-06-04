DROP TABLE IF EXISTS blocks, subscriptions, friendships, users CASCADE;

CREATE TABLE "users" (
  "id" SERIAL PRIMARY KEY,
  "email" varchar UNIQUE
);

CREATE TABLE "friendships" (
  "id" SERIAL PRIMARY KEY,
  "user1_id" integer,
  "user2_id" integer,

  CONSTRAINT unique_friendship UNIQUE (user1_id, user2_id),
  CONSTRAINT ordered_users CHECK (user1_id < user2_id)
);

CREATE TABLE "subscriptions" (
  "id" SERIAL PRIMARY KEY,
  "requestor_id" integer,
  "target_id" integer
);

CREATE TABLE "blocks" (
  "id" SERIAL PRIMARY KEY,
  "requestor_id" integer,
  "target_id" integer
);

CREATE UNIQUE INDEX ON "subscriptions" ("requestor_id", "target_id");

CREATE UNIQUE INDEX ON "blocks" ("requestor_id", "target_id");

ALTER TABLE "friendships" ADD FOREIGN KEY ("user1_id") REFERENCES "users" ("id");

ALTER TABLE "friendships" ADD FOREIGN KEY ("user2_id") REFERENCES "users" ("id");

ALTER TABLE "subscriptions" ADD FOREIGN KEY ("requestor_id") REFERENCES "users" ("id");

ALTER TABLE "subscriptions" ADD FOREIGN KEY ("target_id") REFERENCES "users" ("id");

ALTER TABLE "blocks" ADD FOREIGN KEY ("requestor_id") REFERENCES "users" ("id");

ALTER TABLE "blocks" ADD FOREIGN KEY ("target_id") REFERENCES "users" ("id");

INSERT INTO users (email) VALUES('andy@mail.com');

INSERT INTO users (email) VALUES('alice@mail.com');

INSERT INTO users (email) VALUES('bob@mail.com');

INSERT INTO users (email) VALUES('jack@mail.com');

INSERT INTO users (email) VALUES('lisa@mail.com');
