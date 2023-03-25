from pathlib import Path

# URL for the database connection.
DATABASE_URI = "sqlite:///data/posts.db"
PREFERRED_URL_SCHEME = "https"

# Directory where the website source is.
WEBSITE_DIR = Path("../website")
# Content directory where the posts are rendered.
WEBSITE_POST_DIR = WEBSITE_DIR / "content"

MICROPUB_ME = "https://domain.tld"
# Directory where the uploaded media will be stored. Relative to WEBSITE_DIR.
MICROPUB_MEDIA_DIR = WEBSITE_DIR / "assets" / "media"
# Publicly accessible URL directory for the uploaded media. Here it will become:
# https://domain.tld/media/<filename>.
MICROPUB_MEDIA_URL = "/media"
MICROPUB_TOKEN_ENDPOINT = "https://tokens.indieauth.com/token"
# Setting this flag to true will not authenticate tokens and will consider all
# tokens as valid. Use only for testing/development.
# MICROPUB_ALLOW_ALL_TOKENS_UNSAFE = True
