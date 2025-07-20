ALTER TABLE "tils"
  ADD COLUMN "user_id" INT,
  ADD CONSTRAINT "fk_user"
    FOREIGN KEY ("user_id") REFERENCES "users"("id");

CREATE INDEX "idx_user_id" ON "tils" ("user_id");
