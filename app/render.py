from io import StringIO
from datetime import datetime, timezone
from flask import current_app
from ruamel.yaml import YAML
from app import models
from app import util

yaml = YAML()


def represent_datetime(self, dt):
    return self.represent_scalar("tag:yaml.org,2002:timestamp", dt.isoformat())


yaml.representer.add_representer(datetime, represent_datetime)


def render_post(post: models.Post):
    metadata = {
        "h": post.type.removeprefix("h-"),
        "kind": post.kind.value,
        "published": post.published.replace(tzinfo=timezone.utc),
        "updated": post.updated.replace(tzinfo=timezone.utc) if post.updated else None,
        "tags": post.data.get("category"),
        # Kind-specific properties
        "like-of": util.pluck_one(post.data.get("like-of")),
        "repost-of": util.pluck_one(post.data.get("repost-of")),
        "bookmark-of": util.pluck_one(post.data.get("bookmark-of")),
        "in-reply-to": util.pluck_one(post.data.get("in-reply-to")),
        "photo": post.data.get("photo"),
    }

    # Post processing
    if metadata.get("photo"):
        photos = metadata.get("photo")
        for n in range(len(photos)):
            photo = photos[n]
            if type(photo) == str:
                photos[n] = {"value": photo, "alt": ""}
            elif type(photo) == dict:
                if "value" not in photo or "alt" not in photo:
                    raise Exception(f"{photo} is an invalid representation of a photo")

    metadata = {k: v for k, v in metadata.items() if v != None}

    metadata_yaml = StringIO()
    yaml.dump(metadata, metadata_yaml)

    content = util.pluck_one(post.data.get("content"))
    if type(content) == dict:
        content = content.get("html", content.get("value", ""))

    return f"---\n{metadata_yaml.getvalue()}---\n{content}"
