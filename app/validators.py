from typing import Optional, List
from typing_extensions import TypedDict
from pydantic import BaseModel, Field, PastDatetime, conlist, constr
from app.models import PostKind


class PostPhoto(TypedDict):
    value: str
    alt: str


class PostProperties(BaseModel):
    h: str
    kind: PostKind
    date: PastDatetime
    updated: Optional[PastDatetime] = None
    tags: Optional[List[str]] = None
    summary: Optional[str] = None
    # Kind-specific properties
    like_of: Optional[str] = Field(alias="like-of", default=None)
    repost_of: Optional[str] = Field(alias="repost-of", default=None)
    bookmark_of: Optional[str] = Field(alias="bookmark-of", default=None)
    in_reply_to: Optional[str] = Field(alias="in-reply-to", default=None)
    photo: Optional[List[PostPhoto]] = None
    title: Optional[str] = None

    class Config:
        use_enum_values = True


class Microformats2(BaseModel):
    type: conlist(str, max_length=1)  # type: ignore
    properties: PostProperties
