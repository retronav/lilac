import mimetypes
import os
import posixpath
import shutil
import uuid
from datetime import datetime
from os import path
from pathlib import Path
from typing import Dict, List
from urllib.parse import urljoin, urlparse

import dateutil.parser
import requests
from flask import Blueprint, current_app, g, request
from sqlalchemy import delete, update
from sqlalchemy.exc import NoResultFound, SQLAlchemyError
from sqlalchemy.orm import Session

from app import errors, models, util
from app.database import get_session
from app.render import render_post

# TODO: map errors to correct response codes and return error in JSON as given
# in the spec.

endpoint = Blueprint("micropub", __name__)


def validate(req: Dict):
    for p in ["type", "properties"]:
        if p not in req:
            raise errors.BadRequest(f"request does not have the property '{p}'")


def form_req_to_mf2(flask_request_form) -> Dict:
    form_items = flask_request_form.items(multi=True)
    mf2 = {"type": [], "properties": {}}
    for k, v in form_items:
        if k.endswith("[]"):
            k = k[:-2]

        if k == "h":
            mf2["type"].append("h-" + v)
            continue

        if k in mf2["properties"]:
            mf2["properties"][k].append(v)
        else:
            mf2["properties"][k] = [v]

    return mf2


def url_to_id(url: str) -> str:
    me_url = current_app.config.get("MICROPUB_ME")
    if url.startswith(me_url) == False:
        raise errors.BadRequest(f"URL not from {me_url}")
    post_id = url.removeprefix(me_url).strip("/")
    return post_id


def get_post(post_id: str) -> models.Post:
    with get_session(current_app) as session:
        try:
            return session.query(models.Post).filter(models.Post.id == post_id).one()
        except NoResultFound:
            raise errors.NotFound(f"Post with id {post_id} does not exist")
        except SQLAlchemyError:
            raise errors.InternalServerError


def write_post_to_file(post: models.Post):
    rendered_post = render_post(post)

    posts_dir = current_app.config.get("WEBSITE_POST_DIR") / path.dirname(post.id)
    os.makedirs(posts_dir, exist_ok=True)

    with open(path.join(posts_dir, path.basename(post.id) + ".md"), "w+") as f:
        f.write(rendered_post)


def sync_posts_to_ssg(session: Session):
    website_post_dir = current_app.config.get("WEBSITE_POST_DIR")

    for root, dirs, files in os.walk(website_post_dir):
        if Path(root) == website_post_dir:
            continue
        files = filter(lambda f: f == "_index.md", files)
        for d in dirs:
            shutil.rmtree(website_post_dir / root / d)
        for f in files:
            os.remove(website_post_dir / root / f)

    posts: List[models.Post] = session.query(models.Post).all()

    for post in posts:
        write_post_to_file(post)


@endpoint.before_request
def check_authorization():
    authorization = request.headers.get("Authorization")
    if not authorization:
        raise errors.Unauthorized("No authorization header present")
    if not authorization.startswith("Bearer "):
        raise errors.Unauthorized("Authorization header does not start with 'Bearer '")
    token = authorization.removeprefix("Bearer ")

    # Use this for testing without making calls to endpoint
    if current_app.config.get("MICROPUB_ALLOW_ALL_TOKENS_UNSAFE"):
        current_app.logger.warning(
            "Micropub endpoint is accepting any token. DO NOT USE IN PRODUCTION."
        )
        g.token_scope = "create update delete media"
        return

    token_endpoint = current_app.config.get("MICROPUB_TOKEN_ENDPOINT")
    resp = requests.get(
        token_endpoint,
        headers={"authorization": f"Bearer {token}", "accept": "application/json"},
    ).json()
    if resp.get("error"):
        raise errors.Unauthorized("Token was not authorized")
    if (
        urlparse(resp["me"]).hostname
        != urlparse(current_app.config.get("MICROPUB_ME")).hostname
    ):
        raise errors.Unauthorized(
            "Token 'me' and 'me' set in configuration do not match"
        )
    g.token_scope = resp["scope"]


@endpoint.get("/micropub")
def micropub_query():
    args = request.args.to_dict(flat=False)
    q = request.args.get("q")
    if q == "config":
        return {
            "media-endpoint": current_app.url_for(".micropub_media", _external=True)
        }
    elif q == "source":
        url = request.args.get("url")
        assert url
        post = get_post(url_to_id(url))

        mf2 = {"type": [post.type], "properties": post.data}
        if post.published:
            mf2["properties"]["published"] = [post.published.isoformat()]
        if post.updated:
            mf2["properties"]["updated"] = [post.updated.isoformat()]

        if args.get("properties[]"):
            filtered_properties = {
                key: mf2["properties"][key]
                for key in args["properties[]"]
                if key in mf2["properties"]
            }
            return {"properties": filtered_properties}
        else:
            return mf2
    else:
        raise errors.BadRequest()


@endpoint.post("/micropub")
def micropub_crud():
    body = {}

    if request.is_json:
        body = request.json
    else:
        body = request.form

    if "action" in body and "url" in body:
        action = util.pluck_one(body["action"])
        if action not in ["update", "delete"]:
            raise errors.BadRequest("Unknown action")
        url = util.pluck_one(body["url"])
        post = get_post(url_to_id(url))
        new_data = dict(post.data)

        if action == "update":
            if "update" not in g.token_scope:
                raise errors.UnsufficientScope("Scope 'update' not present in token")

            add_spec: dict | None = body.get("add")
            replace_spec: dict | None = body.get("replace")
            delete_spec: dict | list | None = body.get("delete")

            if add_spec:
                for k, v in add_spec.items():
                    new_data[k] = v
            if replace_spec:
                for k, v in replace_spec.items():
                    new_data[k] = v
            if delete_spec:
                if type(delete_spec) == list or type(delete_spec) == tuple:
                    for k in delete_spec:
                        new_data.pop(k)
                elif type(delete_spec) == dict:
                    for k, v in delete_spec.items():
                        if k in new_data:
                            new_value = [i for i in new_data[k] if i not in v]
                            if len(new_value) == 0:
                                del new_data[k]
                            else:
                                new_data[k] = new_value
            post.updated = datetime.utcnow()
            with get_session(current_app) as session:
                session.execute(
                    update(models.Post)
                    .where(models.Post.id == post.id)
                    .values(data=new_data)
                )
                session.commit()
                sync_posts_to_ssg(session)
            return "", 200

        elif action == "delete":
            if "delete" not in g.token_scope:
                raise errors.UnsufficientScope("Scope 'delete' not present in token")

            with get_session(current_app) as session:
                session.execute(delete(models.Post).where(models.Post.id == post.id))
                session.commit()
                sync_posts_to_ssg(session)
            return "", 200

    if "create" not in g.token_scope:
        raise errors.UnsufficientScope("Scope 'create' not present in token")

    mf2 = body if request.is_json else form_req_to_mf2(body)
    properties = mf2["properties"]

    post = models.Post()
    post.type = mf2["type"][0]
    post.data = {
        key: properties[key]
        for key in properties
        if key not in ["published", "updated"]
    }
    post.kind = post.get_kind()
    post.published = (
        dateutil.parser.isoparse(properties["published"][0])
        if "published" in properties
        else datetime.utcnow()
    )
    with get_session(current_app) as session:
        post.id = post.generate_id(session)

        session.add(post)
        session.commit()
        sync_posts_to_ssg(session)

    return (
        "",
        201,
        {"Location": urljoin(current_app.config.get("MICROPUB_ME"), post.id)},
    )


@endpoint.post("/micropub/media")
def micropub_media():
    if "media" not in g.token_scope:
        raise errors.UnsufficientScope("Scope 'media' not present in token")

    file = request.files.get("file")
    if file:
        filename = str(uuid.uuid4()) + mimetypes.guess_extension(file.mimetype)
        media_dir = current_app.config.get("MICROPUB_MEDIA_DIR")
        os.makedirs(media_dir, exist_ok=True)
        file.save(media_dir / filename)

        return (
            "",
            201,
            {
                "location": urljoin(
                    current_app.config.get("MICROPUB_ME"),
                    posixpath.join(
                        current_app.config.get("MICROPUB_MEDIA_URL"),
                        filename,
                    ),
                )
            },
        )
    else:
        raise errors.BadRequest()
