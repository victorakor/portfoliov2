-- Victor Akor Portfolio v3 — project cover images
--
-- blog_posts already had cover_image; projects had no image column at all, so
-- the admin project editor had nowhere to store an uploaded Cloudinary URL.

ALTER TABLE projects ADD COLUMN IF NOT EXISTS image_url TEXT;
