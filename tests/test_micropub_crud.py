import time_machine
from ruamel.yaml import YAML
from pathlib import Path
from tests.conftest import published_ts, updated_ts
from io import TextIOWrapper
from app.database import get_session
from sqlalchemy import text

traveller = time_machine.travel(published_ts)
yaml = YAML()


class Fixture:
    post: dict[any, any]
    metadata: dict[any, any]

    def __init__(self, post, metadata) -> None:
        self.post = post
        self.metadata = metadata

    def assert_content(self, file: TextIOWrapper):
        rendered_post = file.read()
        first_fence = rendered_post.index("---") + 3
        last_fence = rendered_post.index("---", first_fence)
        metadata = yaml.load(rendered_post[first_fence:last_fence])
        assert dict(metadata) == self.metadata

        content = self.post["properties"].get("content")
        if content:
            content = content[0] if type(content) == list else content["html"]
            assert rendered_post[last_fence + 4 :] == content  # +4 for '---\n'


def test_post_creation(client, app):
    posts_dir = app.config["WEBSITE_POST_DIR"]
    fixtures = [
        Fixture(
            {"type": ["h-entry"], "properties": {"content": ["Hello World!"]}},
            {
                "h": "entry",
                "kind": "note",
                "published": published_ts,
            },
        ),
        Fixture(
            {
                "type": ["h-entry"],
                "properties": {"like-of": ["https://some-cool.website"]},
            },
            {
                "h": "entry",
                "kind": "like",
                "published": published_ts,
                "like-of": "https://some-cool.website",
            },
        ),
        Fixture(
            {
                "type": ["h-entry"],
                "properties": {
                    "photo": ["https://nice.photo/1"],
                    "content": ["What a nice flower!"],
                },
            },
            {
                "h": "entry",
                "kind": "photo",
                "published": published_ts,
                "photo": [{"value": "https://nice.photo/1", "alt": ""}],
            },
        ),
        Fixture(
            {
                "type": ["h-entry"],
                "properties": {
                    "name": ["Example article"],
                    "content": {
                        "html": "<p>Did you know? This thing can <i>handle</i> HTML quite well!</p>"
                    },
                },
            },
            {
                "h": "entry",
                "kind": "article",
                "published": published_ts,
                "title": "Example article",
            },
        ),
    ]
    traveller.start()

    for fixture in fixtures:
        response = client.post("/micropub", json=fixture.post)
        post_id = (
            str(response.headers["location"])
            .removeprefix(app.config["MICROPUB_ME"])
            .lstrip("/")
        )
        fixture.assert_content(open(posts_dir / (post_id + ".md"), "r"))

    traveller.stop()


def test_updating_post(client, app):
    original_post = Fixture(
        {"type": ["h-entry"], "properties": {"content": ["Hello World!"]}},
        {
            "h": "entry",
            "kind": "note",
            "published": published_ts,
        },
    )

    updated_posts = [
        Fixture(
            {
                "type": ["h-entry"],
                "properties": {
                    "content": ["Hello World! This content was added in"],
                    "category": ["foo", "bar"],
                },
            },
            {
                "h": "entry",
                "kind": "note",
                "published": published_ts,
                "lastmod": updated_ts,
                "tags": ["foo", "bar"],
            },
        ),
        Fixture(
            {
                "type": ["h-entry"],
                "properties": {
                    "content": ["Hello World! This content was added in"],
                    "category": ["foo"],
                },
            },
            {
                "h": "entry",
                "kind": "note",
                "published": published_ts,
                "lastmod": updated_ts,
                "tags": ["foo"],
            },
        ),
        Fixture(
            {
                "type": ["h-entry"],
                "properties": {
                    "content": ["Hello World! This content was added in and changed"],
                },
            },
            {
                "h": "entry",
                "kind": "note",
                "published": published_ts,
                "lastmod": updated_ts,
            },
        ),
    ]
    updates = [
        {
            "add": {"category": ["foo", "bar"]},
            "replace": {"content": ["Hello World! This content was added in"]},
        },
        {"delete": {"category": ["bar"]}},
        {
            "replace": {
                "content": ["Hello World! This content was added in and changed"]
            },
            "delete": ["category"],
        },
    ]
    posts_dir = app.config["WEBSITE_POST_DIR"]

    traveller.start()
    response = client.post("/micropub", json=original_post.post)
    post_id = (
        str(response.headers["location"])
        .removeprefix(app.config["MICROPUB_ME"])
        .lstrip("/")
    )
    original_post.assert_content(open(posts_dir / (post_id + ".md"), "r"))
    traveller.stop()

    # modify the update trigger to use a mock timestamp
    with get_session(app) as session:
        session.execute(text("DROP TRIGGER IF EXISTS posts_update_trigger;"))
        session.execute(
            text(
                """
CREATE TRIGGER posts_update_trigger
AFTER UPDATE ON posts
FOR EACH ROW
BEGIN
    UPDATE posts SET updated = "2023-01-30 16:30:00" WHERE id = NEW.id;
END;
"""
            )
        )

    for n in range(len(updates)):
        update_schema = updates[n]
        updated_post = updated_posts[n]
        update_schema.update(
            {
                "action": "update",
                "url": f"{app.config['MICROPUB_ME']}/{post_id}",
            }
        )
        response = client.post(
            "/micropub",
            json=update_schema,
        )
        assert response.status_code == 200
        updated_post.assert_content(open(posts_dir / (post_id + ".md"), "r"))


def test_deleting_post(client, app):
    posts_dir = app.config["WEBSITE_POST_DIR"]
    post = Fixture(
        {"type": ["h-entry"], "properties": {"content": ["Hello World!"]}},
        {
            "h": "entry",
            "kind": "note",
            "published": published_ts,
        },
    )
    traveller.start()
    response = client.post("/micropub", json=post.post)
    post_id = (
        str(response.headers["location"])
        .removeprefix(app.config["MICROPUB_ME"])
        .lstrip("/")
    )
    post_file = posts_dir / (post_id + ".md")
    post.assert_content(open(post_file, "r"))

    response = client.post(
        "/micropub",
        json={"action": "delete", "url": f"{app.config['MICROPUB_ME']}/{post_id}"},
    )
    assert response.status_code == 200
    assert post_file.exists() == False

    traveller.stop()
