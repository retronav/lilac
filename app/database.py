from flask import Flask
from sqlalchemy import create_engine
from sqlalchemy.orm import declarative_base, sessionmaker

Base = declarative_base()


def get_session(app: Flask):
    """
    Get an SQLAlchemy `.Session` by configuring the engine to use `DATABASE_URI`
    set in Flask app configuration.
    """
    engine = create_engine(app.config.get("DATABASE_URI"))
    session = sessionmaker(engine)
    return session()
