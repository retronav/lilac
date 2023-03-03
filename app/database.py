from sqlalchemy import create_engine
from sqlalchemy.orm import sessionmaker, declarative_base
from data import config

engine = create_engine(config.DATABASE_URI)
Session = sessionmaker(engine)
Base = declarative_base()