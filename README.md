# Lilac

A content management system based on technologies described by the
[Indieweb](https://indieweb.org). Lilac contains the following components:

- A [micropub](https://micropub.net) server

## Installation

1. Clone this repository.
2. Create a virtual environment using [pipenv](https://pypi.org/project/pipenv/).
   ```shell
   mkdir .venv
   pipenv shell
   ```
3. Install the dependencies.
   ```shell
   pipenv install -r requirements.txt
   pipenv install --dev
   ```
4. Copy the example configuration and adjust it to your needs.
   ```shell
   cp data/example_config.py data/config.py
   ```
5. Create a database.
   ```shell
   alembic upgrade head
   ```
6. Run the server.
   ```shell
   # This will run on port 5000, adjust to needs.
   waitress-serve --port 5000 --call "app:create_app"
   ```
   You can then connect Lilac to your domain using a reverse proxy server.

## Upgrading

1. Pull updates to the repository.
   ```shell
   git pull
   ```
2. Update dependencies.
   ```shell
   pipenv update --dev
   ```
3. Run migrations on the database.
   ```shell
   alembic upgrade head
   ```
