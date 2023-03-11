def test_request_example(client):
    response = client.get(
        "/micropub?q=config", headers={"authorization": "Bearer testing"}
    )
    print(response.text)
