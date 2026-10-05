DROP TABLE IF EXISTS comments;
-- The `ltree` extension is intentionally left installed: dropping it in a down
-- migration is unnecessarily destructive if other schemas in a shared database
-- ever come to depend on it, and leaving it installed is harmless.
