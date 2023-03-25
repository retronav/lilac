import os
import tempfile


def test_upload(client, app):
    with tempfile.NamedTemporaryFile() as media_file:
        response = client.post("/micropub/media", data={"file": media_file.file})
        uploaded_file_url = response.headers["location"]
        uploaded_file_name = os.path.basename(uploaded_file_url)

        assert (app.config.get("MICROPUB_MEDIA_DIR") / uploaded_file_name).exists()
