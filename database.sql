CREATE TABLE "user" (
  "id" INT GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
  "full_name" varchar(60),
  "phone_number" varchar(16) UNIQUE
);

CREATE TABLE "password" (
  "id" INT GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
  "user_id" int,
  "password" varchar(255),
  "salt" varchar(16)
);

CREATE TABLE "login" (
  "id" INT GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
  "user_id" int,
  "success_login" int
);

ALTER TABLE "password" ADD FOREIGN KEY ("user_id") REFERENCES "user" ("id");

ALTER TABLE "login" ADD FOREIGN KEY ("user_id") REFERENCES "user" ("id");