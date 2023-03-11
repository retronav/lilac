import pytest
import os
import tempfile
from alembic.config import Config
from alembic import command
from app import create_app


@pytest.fixture()
def app():
    instance_dir = tempfile.TemporaryDirectory()
    website_dir = os.path.join(instance_dir.name, "website")
    os.mkdir(website_dir)

    config = {
        "DATABASE_URI": f"sqlite:/{os.path.join(instance_dir.name, 'posts.db')}",
        "WEBSITE_DIR": website_dir,
        "WEBSITE_POST_DIR": ".",
        "MICROPUB_ME": "https://domain.tld",
        "MICROPUB_MEDIA_URL": "/uploads",
        "MICROPUB_MEDIA_DIR": ".",
        "MICROPUB_ALLOW_ALL_TOKENS_UNSAFE": True,
    }

    alembic_cfg = Config(os.path.abspath("alembic.ini"))
    alembic_cfg.set_section_option("alembic", "sqlalchemy.url", config["DATABASE_URI"])
    command.upgrade(alembic_cfg, "head")

    app = create_app(config)

    yield app

    instance_dir.cleanup()


@pytest.fixture()
def client(app):
    return app.test_client()


@pytest.fixture()
def runner(app):
    return app.test_cli_runner()
