import enum
from app.database import Base
from sqlalchemy import func, text, Column, Text, Enum, DateTime, JSON
from sqlalchemy.orm import Session
from datetime import datetime


class PostKind(enum.Enum):
    note = "note"
    article = "article"
    like = "like"
    bookmark = "bookmark"
    repost = "repost"


class Post(Base):
    __tablename__ = "posts"

    id = Column(Text, primary_key=True)
    type = Column(Text, nullable=False)
    kind = Column(Enum(PostKind), nullable=False)
    published = Column(DateTime, nullable=False, default=datetime.utcnow())
    updated = Column(DateTime, server_default=text("NULL"))
    data = Column(JSON, nullable=False)

    def generate_id(self, session: Session):
        id_suffix = self.data.get("mp-slug", [""])[0]
        if not id_suffix:
            id_suffix = f"{self.posts_published_today(session)+1:02d}"

        # Generate the full ID string, eg. notes/2023/02/15/01 or notes/2023/02/15/my-slug
        id_str = f"{self.kind.value}s/{self.published:%Y/%m/%d}/{id_suffix}"

        return id_str

    def posts_published_today(self, session: Session) -> int:
        num_posts = (
            session.query(func.count(Post.id))
            .filter(
                Post.kind == self.kind,
                func.date(Post.published) == self.published.date(),
            )
            .scalar()
        )
        num_deleted_posts = (
            session.query(func.count(DeletedPost.id))
            .filter(
                DeletedPost.kind == self.kind,
                func.date(DeletedPost.published) == self.published.date(),
            )
            .scalar()
        )
        return num_posts + num_deleted_posts

    def get_kind(self) -> PostKind:
        if "like-of" in self.data:
            return PostKind.like
        elif "name" in self.data:
            return PostKind.article
        elif "bookmark-of" in self.data:
            return PostKind.bookmark
        elif "repost-of" in self.data:
            return PostKind.repost
        else:
            return PostKind.note


class DeletedPost(Base):
    __tablename__ = "deleted_posts"
    id = Column(Text, primary_key=True)
    kind = Column(Enum(PostKind), nullable=False)
    published = Column(DateTime, nullable=False)
