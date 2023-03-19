import time_machine
from tests.conftest import published_ts

traveller = time_machine.travel(published_ts)


def test_endpoint_configuration(client):
    response = client.get("/micropub?q=config").json

    assert response["media-endpoint"] == "https://domain.tld/micropub/media"


def test_source(client):
    traveller.start()
    # create test post
    response = client.post(
        "/micropub",
        data={"h": "entry", "content": "Hello World", "category[]": ["foo", "bar"]},
    )

    assert response.status_code == 201
    post_url = response.headers["location"]

    # query whole post
    response = client.get(f"/micropub?q=source&url={post_url}")
    assert response.json == {
        "type": ["h-entry"],
        "properties": {
            "content": ["Hello World"],
            "category": ["foo", "bar"],
            "published": ["2023-01-29T05:30:00"],
        },
    }
    traveller.stop()
