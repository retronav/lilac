"""add posts table triggers

Revision ID: 0f4e73092a91
Revises: 2b707be1442b
Create Date: 2023-03-02 08:52:13.870622+00:00

"""
from alembic import op
import sqlalchemy as sa


# revision identifiers, used by Alembic.
revision = "0f4e73092a91"
down_revision = "2b707be1442b"
branch_labels = None
depends_on = None


def upgrade() -> None:
    op.execute(
        """
    CREATE TRIGGER posts_update_trigger
    AFTER UPDATE ON posts
    FOR EACH ROW
    BEGIN
        UPDATE posts SET updated = CURRENT_TIMESTAMP WHERE id = NEW.id;
    END;
    """
    )

    op.execute(
        """CREATE TRIGGER posts_delete_trigger
    BEFORE DELETE ON posts
    FOR EACH ROW
    BEGIN
        INSERT INTO deleted_posts (id, kind, published)
        VALUES (OLD.id, OLD.kind, OLD.published);
    END;
    """
    )
    pass


def downgrade() -> None:
    op.execute("DROP TRIGGER posts_update_trigger;")
    op.execute("DROP TRIGGER posts_delete_trigger;")
    pass
