-- Bổ sung thông tin chuyên khoa: code, slug, symptoms, location.
-- Bản ghi cũ được backfill code dạng SPEC-001 và slug sinh từ name (bỏ dấu tiếng Việt).
BEGIN;

ALTER TABLE specializations
    ADD COLUMN code     varchar(50),
    ADD COLUMN slug     varchar(255),
    ADD COLUMN symptoms jsonb NOT NULL DEFAULT '[]'::jsonb,
    ADD COLUMN location varchar(255);

WITH numbered AS (
    SELECT spec_id, row_number() OVER (ORDER BY name, spec_id) AS rn
    FROM specializations
)
UPDATE specializations s
SET code = 'SPEC-' || lpad(n.rn::text, 3, '0')
FROM numbered n
WHERE n.spec_id = s.spec_id;

WITH base AS (
    SELECT spec_id,
           COALESCE(NULLIF(trim(BOTH '-' FROM regexp_replace(
               lower(translate(name,
                   'àÀáÁảẢãÃạẠăĂằẰắẮẳẲẵẴặẶâÂầẦấẤẩẨẫẪậẬđĐèÈéÉẻẺẽẼẹẸêÊềỀếẾểỂễỄệỆìÌíÍỉỈĩĨịỊòÒóÓỏỎõÕọỌôÔồỒốỐổỔỗỖộỘơƠờỜớỚởỞỡỠợỢùÙúÚủỦũŨụỤưƯừỪứỨửỬữỮựỰỳỲýÝỷỶỹỸỵỴ',
                   'aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaddeeeeeeeeeeeeeeeeeeeeeeiiiiiiiiiioooooooooooooooooooooooooooooooooouuuuuuuuuuuuuuuuuuuuuuyyyyyyyyyy')),
               '[^a-z0-9]+', '-', 'g')), ''), 'specialization') AS slug
    FROM specializations
),
ranked AS (
    SELECT spec_id, slug, row_number() OVER (PARTITION BY slug ORDER BY spec_id) AS rn
    FROM base
)
UPDATE specializations s
SET slug = CASE WHEN r.rn = 1 THEN r.slug ELSE r.slug || '-' || r.rn END
FROM ranked r
WHERE r.spec_id = s.spec_id;

ALTER TABLE specializations
    ALTER COLUMN code SET NOT NULL,
    ALTER COLUMN slug SET NOT NULL;

CREATE UNIQUE INDEX idx_specializations_code ON specializations (code);
CREATE UNIQUE INDEX idx_specializations_slug ON specializations (slug);

COMMIT;
