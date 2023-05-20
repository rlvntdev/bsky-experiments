import io
import logging
import os
from time import time
from typing import List

import aiohttp
from fastapi import FastAPI, Request
from opentelemetry import trace
from opentelemetry.exporter.otlp.proto.http.trace_exporter import OTLPSpanExporter
from opentelemetry.instrumentation.fastapi import FastAPIInstrumentor
from opentelemetry.sdk.resources import SERVICE_NAME, Resource
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import BatchSpanProcessor
from PIL import Image
from prometheus_client import Counter
from prometheus_fastapi_instrumentator import Instrumentator
from pythonjsonlogger import jsonlogger
from starlette.middleware.base import BaseHTTPMiddleware

from .models import ImageMeta, ImageResult
from .object_detection import detect_objects

# Set up JSON logging
formatter = jsonlogger.JsonFormatter()
handler = logging.StreamHandler()

# Use OUR `formatter` to format all `logging` entries.
handler.setFormatter(formatter)
root_logger = logging.getLogger()
root_logger.addHandler(handler)
root_logger.setLevel(logging.INFO)

for _log in ["uvicorn", "uvicorn.error"]:
    # Clear the log handlers for uvicorn loggers, and enable propagation
    # so the messages are caught by our root logger and formatted correctly
    # by structlog
    logging.getLogger(_log).handlers.clear()
    logging.getLogger(_log).propagate = True

# Since we re-create the access logs ourselves, to add all information
# in the structured log, we clear the handlers and prevent the logs to propagate to
# a logger higher up in the hierarchy (effectively rendering them silent).
logging.getLogger("uvicorn.access").handlers.clear()
logging.getLogger("uvicorn.access").propagate = False

images_processed_successfully = Counter(
    "images_processed_successfully", "Number of images processed successfully"
)
images_failed = Counter("images_failed", "Number of images failed")
images_submitted = Counter("images_submitted", "Number of images submitted")


class LoggingMiddleware(BaseHTTPMiddleware):
    async def dispatch(self, request: Request, call_next):
        start_time = time()
        response = await call_next(request)
        process_time = time() - start_time
        logging.info(
            {
                "message": "request handled",
                "path": request.url.path,
                "method": request.method,
                "processing_time": process_time,
                "status_code": response.status_code,
                "query_params": request.query_params,
            },
        )
        return response


# Set up OpenTelemetry
otel_endpoint = os.getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
if otel_endpoint:
    resource = Resource(attributes={SERVICE_NAME: "bsky-object-detection"})
    trace.set_tracer_provider(TracerProvider(resource=resource))
    trace.get_tracer_provider().add_span_processor(
        BatchSpanProcessor(
            OTLPSpanExporter(
                endpoint=otel_endpoint + "v1/traces",
            )
        )
    )

app = FastAPI()

# Instrument FastAPI for OpenTelemetry
if otel_endpoint:
    FastAPIInstrumentor.instrument_app(app)

# Instrument FastAPI for Prometheus
Instrumentator().instrument(
    app,
    latency_lowr_buckets=[0.01, 0.05, 0.1, 0.2, 0.5, 1.0, 2.5, 5, 10],
).expose(app, include_in_schema=False)

# Add logging middleware
app.add_middleware(LoggingMiddleware)


@app.post("/detect_objects", response_model=List[ImageResult])
async def detect_objects_endpoint(image_metas: List[ImageMeta]):
    images_submitted.inc(len(image_metas))
    image_results: List[ImageResult] = []
    async with aiohttp.ClientSession() as session:
        for image_meta in image_metas:
            # Download the image from the URL in the payload
            async with session.get(image_meta.url) as resp:
                # If the response is not 200, log an error and continue to the next image
                if resp.status != 200:
                    logging.error(
                        f"Error fetching image from {image_meta.url} - {resp.status}"
                    )
                    image_results.append(ImageResult(meta=image_meta, results=[]))
                    continue
                imageData = await resp.read()
                pilImage = Image.open(io.BytesIO(imageData))

                # Run the object detection model
                try:
                    detection_results = detect_objects(pilImage)
                except Exception as e:
                    logging.error(f"Error running object detection model: {e}")
                    image_results.append(ImageResult(meta=image_meta, results=[]))
                    images_failed.inc()
                    continue
                images_processed_successfully.inc()
                image_results.append(
                    ImageResult(meta=image_meta, results=detection_results)
                )
        return image_results