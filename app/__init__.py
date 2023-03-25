import os

from flask import Flask

from app import errors, micropub
from app.database import get_session


def create_app(test_config=None, instance_path=os.path.join(os.getcwd(), "data")):
    app = Flask(
        __name__,
        instance_path=instance_path,
        instance_relative_config=True,
    )

    if test_config is None:
        app.config.from_pyfile("config.py")
    else:
        app.config.from_mapping(test_config)

    try:
        os.makedirs(app.instance_path)
    except OSError:
        pass

    # Check required configuration properties
    required_properties = [
        "DATABASE_URI",
        "WEBSITE_DIR",
        "WEBSITE_POST_DIR",
        "MICROPUB_ME",
        "MICROPUB_MEDIA_DIR",
        "MICROPUB_MEDIA_URL",
    ]
    for property in required_properties:
        if not app.config.get(property):
            raise Exception(f"Missing configuration property {property}")

    with app.app_context():
        with get_session(app) as session:
            micropub.sync_posts_to_ssg(session)

    # Register errors
    with app.app_context():

        @app.errorhandler(errors.BaseError)
        def handle_error(error: errors.BaseError):
            app.logger.info(f"the error: {error.code}")
            return {"error": error.error, "message": error.message}, error.code

    app.register_blueprint(micropub.endpoint)

    return app
