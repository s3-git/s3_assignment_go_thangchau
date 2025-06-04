CREATE TABLE "users" (
  "id" integer PRIMARY KEY,
  "email" varchar UNIQUE
);

CREATE TABLE "friendships" (
  "id" integer PRIMARY KEY,
  "user1_id" integer,
  "user2_id" integer,

  CONSTRAINT unique_friendship UNIQUE (user1_id, user2_id),
  CONSTRAINT ordered_users CHECK (user1_id < user2_id)
);

CREATE TABLE "subscriptions" (
  "id" integer PRIMARY KEY,
  "requestor_id" integer,
  "target_id" integer
);

CREATE TABLE "blocks" (
  "id" integer PRIMARY KEY,
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

INSERT INTO users VALUES("andy@mail.com")
INSERT INTO users VALUES("alice@mail.com")
INSERT INTO users VALUES("bob@mail.com")
INSERT INTO users VALUES("jack@mail.com")
INSERT INTO users VALUES("lisa@mail.com")
