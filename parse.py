import os
import sys
import xml.etree.ElementTree as ET
import sqlite3


def script(db, q):
    c = db.cursor()
    c.executescript(q)
    db.commit()
    c.close()


def query(db, q, *args):
    c = db.cursor()
    c.execute(q, args)
    db.commit()

    data = c.fetchall()
    c.close()

    return data


def initdb(db):
    script(db, open("schema.sql").read())


def txt(elt, child):
    kid = elt.find(child)
    if kid is not None:
        return kid.text
    return ""


def dig(elt, *tags):
    for tag in tags:
        t = txt(elt, tag)
        if t:
            return t
    return ""


def handle_quote(db, post):
    id_ = post.attrib["id"]
    dt = post.attrib["date-gmt"]
    quote_text = txt(post, "quote-text")
    quote_source = dig(post, "quote-source", "tag")
    if quote_source == "":
        print("no quote source found for post:", id_)
    query(
        db,
        "INSERT INTO quote_posts VALUES(?, ?, ?, ?)",
        dt,
        id_,
        quote_text,
        quote_source,
    )


def handle_photo(db, post):
    id_ = post.attrib["id"]
    dt = post.attrib["date-gmt"]

    # verify that these are the only tags found in photo posts
    # tags = [elt.tag for elt in post]
    # for tag in tags:
    #     if tag not in ("photo-link-url", "photo-url", "photo-caption"):
    #         print(id_, tags)

    caption = txt(post, "photo-caption")
    link = txt(post, "photo-link-url")

    query(db, "INSERT INTO photo_posts VALUES(?, ?, ?, ?)", dt, id_, caption, link)

    for photo in post.findall("photo-url"):
        query(db, "INSERT INTO photo_urls VALUES(?, ?)", id_, photo.text)


def handle_text(db, post):
    id_ = post.attrib["id"]
    dt = post.attrib["date-gmt"]

    # assert that there is one child, regular-body, and an optional
    # regular-title
    kids = list(post)
    tags = [kid.tag in ("regular-body", "regular-title", "tag") for kid in kids]
    assert len(kids) < 3 and all(tags), f"{id_}, {kids}, {tags}"

    title = txt(post, "regular-title")
    text = txt(post, "regular-body")
    assert text != ""

    query(db, "INSERT INTO text_posts VALUES(?, ?, ?, ?)", dt, id_, title, text)


def handle_link(db, post):
    id_ = post.attrib["id"]
    dt = post.attrib["date-gmt"]

    kids = list(post)
    tags = [
        kid.tag in ("link-text", "link-url", "link-description", "tag") for kid in kids
    ]
    assert all(tags), f"{id_}, {kids}, {tags}"

    text = txt(post, "link-text")
    desc = txt(post, "link-description")
    url = txt(post, "link-url")
    assert url != ""

    query(db, "INSERT INTO link_posts VALUES(?, ?, ?, ?, ?)", dt, id_, url, text, desc)


def handle_video(db, post):
    id_ = post.attrib["id"]
    dt = post.attrib["date-gmt"]

    kids = list(post)
    tags = [
        kid.tag in ("video-source", "video-caption", "video-player", "tag")
        for kid in kids
    ]
    assert all(tags), f"{id_}, {kids}, {tags}"

    source = txt(post, "video-source")
    assert source != ""
    caption = txt(post, "video-caption")

    query(db, "INSERT INTO video_posts VALUES(?, ?, ?, ?)", dt, id_, source, caption)


def handle_audio(db, post):
    id_ = post.attrib["id"]
    dt = post.attrib["date-gmt"]

    kids = list(post)
    tags = [
        kid.tag in ("audio-caption", "audio-player", "audio-embed", "tag")
        for kid in kids
    ]
    assert all(tags), f"{id_}, {kids}, {tags}"

    player = txt(post, "audio-player")
    caption = txt(post, "audio-caption")

    query(db, "INSERT INTO audio_posts VALUES(?, ?, ?, ?)", dt, id_, player, caption)


handlers = {
    "quote": handle_quote,
    "photo": handle_photo,
    "regular": handle_text,
    "link": handle_link,
    "video": handle_video,
    "audio": handle_audio,
    # there is a "postcard" type but in my case it appears to be only spam
}


def main(db):
    posts = ET.parse("tumblr_backup/posts.xml").findall(".//post")

    for post in posts:
        typ = post.attrib["type"]
        if typ in handlers:
            # insert this into the generic posts table
            id_ = post.attrib["id"]
            slug = post.attrib["url-with-slug"].split("/")[-1]
            query(
                db,
                "INSERT INTO posts VALUES(?, ?, ?, ?)",
                post.attrib["date-gmt"],
                id_,
                typ,
                slug,
            )
            for tag in post.findall("tag"):
                query(db, "INSERT INTO tags VALUES(?, ?)", id_, tag.text)
            handlers[typ](db, post)
        else:
            print("no handler for:", typ, post.attrib["id"])


if __name__ == "__main__":
    f = sys.argv[1]
    with sqlite3.connect(f) as db:
        initdb(db)
        main(db)
        query(db, "CREATE INDEX posts_by_date ON posts(date, id);")
