CREATE TABLE IF NOT EXISTS posts (
    date text,
    id text,
    type text,
    slug text
);

-- this will work because dates sort properly lexicographically
-- it is commented out because we want to load the data faster, then add the
-- index on at the end
-- CREATE INDEX posts_by_date ON posts(date, id);

CREATE TABLE IF NOT EXISTS quote_posts (
    date text,
    id text,
    body text,
    source text
);
CREATE TABLE IF NOT EXISTS photo_posts (
    date text,
    id text,
    caption text,
    link text
);
CREATE TABLE IF NOT EXISTS photo_urls (
    id text,
    url text,
    FOREIGN KEY(id) REFERENCES photo_posts
);
CREATE TABLE IF NOT EXISTS text_posts (
    date text,
    id text,
    title text,
    body text
);
CREATE TABLE IF NOT EXISTS link_posts (
    date text,
    id text,
    url text,
    text text,
    desc text
);
CREATE TABLE IF NOT EXISTS video_posts (
    date text,
    id text,
    source text,
    caption text
);
CREATE TABLE IF NOT EXISTS audio_posts (
    date text,
    id text,
    player text,
    caption text
);
CREATE TABLE IF NOT EXISTS tags (
    id text,
    tag text
);
