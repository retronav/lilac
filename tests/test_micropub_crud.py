import time_machine
from ruamel.yaml import YAML
from pathlib import Path
from tests.conftest import published_ts

traveller = time_machine.travel(published_ts)
yaml = YAML()


def test_post_rendering(client, app):
    traveller.start()

    fixtures = [
        {
            "post": {"type": ["h-entry"], "properties": {"content": ["Hello World!"]}},
            "metadata": {
                "h": "entry",
                "kind": "note",
                "published": published_ts,
            },
        },
        {
            "post": {
                "type": ["h-entry"],
                "properties": {"like-of": ["https://some-cool.website"]},
            },
            "metadata": {
                "h": "entry",
                "kind": "like",
                "published": published_ts,
                "like-of": "https://some-cool.website",
            },
        },
        {
            "post": {
                "type": ["h-entry"],
                "properties": {
                    "photo": ["https://nice.photo/1"],
                    "content": ["What a nice flower!"],
                },
            },
            "metadata": {
                "h": "entry",
                "kind": "photo",
                "published": published_ts,
                "photo": [{"value": "https://nice.photo/1", "alt": ""}],
            },
        },
    ]

    for i, fixture in enumerate(fixtures):
        response = client.post("/micropub", json=fixture["post"])
        post_file_path = (
            str(response.headers["location"])
            .removeprefix(app.config["MICROPUB_ME"])
            .lstrip("/")
        )

        posts_dir = Path(app.config["WEBSITE_DIR"]) / app.config["WEBSITE_POST_DIR"]

        with open(posts_dir / (post_file_path + ".md"), "r") as f:
            rendered_post = f.read()
            first_fence = rendered_post.index("---") + 3
            last_fence = rendered_post.index("---", first_fence)
            metadata = yaml.load(rendered_post[first_fence:last_fence])
            assert dict(metadata) == fixture["metadata"]

            content = fixture["post"]["properties"].get("content")
            if content:
                content = content[0] if type(content) == list else content["html"]
                assert rendered_post[last_fence + 4 :] == content  # +4 for '---\n'

    traveller.stop()
