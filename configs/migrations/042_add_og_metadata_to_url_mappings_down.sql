ALTER TABLE url_mappings
  DROP COLUMN IF EXISTS og_title,
  DROP COLUMN IF EXISTS og_description,
  DROP COLUMN IF EXISTS og_image,
  DROP COLUMN IF EXISTS twitter_card;
