import pytest
import os
import tempfile
import datetime
from alembic.config import Config
from alembic import command
from flask.testing import FlaskClient
from app import create_app

published_ts = datetime.datetime(2023, 1, 29, 5, 30)
updated_ts = datetime.datetime(2023, 1, 30, 16, 30)


class CustomClient(FlaskClient):
    def open(self, *args, **kwargs):
        kwargs.setdefault("headers", {"Authorization": "Bearer testing"})
        return super().open(*args, **kwargs)


@pytest.fixture()
def app():
    website_dir = tempfile.TemporaryDirectory()
    os.mkdir(os.path.join(website_dir.name, "content"))
    config = {
        "DATABASE_URI": f"sqlite:///{os.path.join(website_dir.name, 'posts.db')}",
        "WEBSITE_DIR": website_dir.name,
        "PREFERRED_URL_SCHEME": "https",
        "WEBSITE_POST_DIR": "content",
        "SERVER_NAME": "domain.tld",
        "MICROPUB_ME": "https://domain.tld",
        "MICROPUB_MEDIA_URL": "/uploads",
        "MICROPUB_MEDIA_DIR": ".",
        "MICROPUB_ALLOW_ALL_TOKENS_UNSAFE": True,
        "TESTING": True,
    }

    alembic_cfg = Config("alembic.ini")
    alembic_cfg.set_main_option("sqlalchemy.url", config["DATABASE_URI"])
    command.upgrade(alembic_cfg, "head")

    app = create_app(config)
    app.test_client_class = CustomClient

    yield app

    website_dir.cleanup()


@pytest.fixture()
def client(app):
    return app.test_client()


@pytest.fixture()
def runner(app):
    return app.test_cli_runner()
