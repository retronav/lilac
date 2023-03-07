from flask import current_app


class BaseError(Exception):
    message = ""
    code = 0
    error = ""

    def __init__(self, message) -> None:
        super().__init__(message)
        self.message = message


# Error codes defined in the spec
class BadRequest(BaseError):
    error = "bad_request"
    code = 400


class Forbidden(BaseError):
    error = "forbidden"
    code = 403


class Unauthorized(BaseError):
    error = "unauthorized"
    code = 401


class UnsufficientScope(BaseError):
    error = "unsufficient_scope"
    code = 403


# Custom error codes
class InternalServerError(BaseError):
    error = "internal_server_error"
    code = 500


class NotFound(BaseError):
    error = "not_found"
    code = 404
