import json
from flask import request, Response
import rest
import http
from otel import tracer
from opentelemetry.propagate import extract
from opentelemetry.trace import SpanKind

def main():
    with tracer.start_as_current_span("handle request", context=extract(request.headers), kind=SpanKind.SERVER) as span:
        if request.method == "POST":   #SendOrderConfirmation
            span.add_event("invoke SendOrderConfirmation")
            try:
                body = request.get_data()
                req = rest.dict2SendOrderConfirmationRequest(json.loads(body))
                rest.emailservice.sendOrderConfirmation(req)
                span.add_event("successfully handle request")
                return Response(status=http.HTTPStatus.OK)
            except Exception as e:
                print(e)
                span.add_event("an error occurred in SendOrderConfirmation")
                return Response(status=http.HTTPStatus.BAD_REQUEST)
        else:
            print("methods other than POST are not supported")
            span.add_event("methods other than POST are not supported")
            return Response(status=http.HTTPStatus.BAD_REQUEST)