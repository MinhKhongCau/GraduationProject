import json
from datetime import datetime, timezone

from starlette.middleware.base import BaseHTTPMiddleware
from starlette.requests import Request
from starlette.responses import JSONResponse

# Paths whose bodies must stay untouched (Swagger UI fetches the raw
# OpenAPI schema and would choke on an envelope wrapper around it).
_EXCLUDED_PATHS = {"/openapi.json"}


class ResponseEnvelopeMiddleware(BaseHTTPMiddleware):
    """Wraps every JSON response (success or error) in the shared API envelope:
    {statusCode, timestamp, method, path, result, message}."""

    async def dispatch(self, request: Request, call_next):
        response = await call_next(request)

        if request.url.path in _EXCLUDED_PATHS:
            return response

        content_type = response.headers.get("content-type", "")
        if "application/json" not in content_type:
            return response

        body = b"".join([chunk async for chunk in response.body_iterator])
        raw = json.loads(body) if body else None

        message, result = _split_body(raw, response.status_code)

        envelope = {
            "statusCode": response.status_code,
            "timestamp": datetime.now(timezone.utc).isoformat(),
            "method": request.method,
            "path": request.url.path,
            "result": result,
            "message": message,
        }

        headers = {k: v for k, v in response.headers.items() if k.lower() != "content-length"}
        return JSONResponse(content=envelope, status_code=response.status_code, headers=headers)


def _split_body(raw, status_code: int):
    if status_code >= 400:
        if isinstance(raw, dict) and "detail" in raw:
            detail = raw["detail"]
            if isinstance(detail, str):
                return detail, None
            # e.g. pydantic validation errors, returned as a list of dicts
            return "Dữ liệu không hợp lệ!", detail
        return "Có lỗi xảy ra!", raw

    if isinstance(raw, dict) and "message" in raw:
        message = raw["message"]
        rest = {k: v for k, v in raw.items() if k != "message"}
        result = rest.get("data", rest)
        if isinstance(result, dict) and not result:
            result = None
        return message, result

    return "Thành công!", raw
