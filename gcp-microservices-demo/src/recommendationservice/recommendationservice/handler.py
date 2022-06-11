import json
from flask import request, Response
import rest
import http
from otel import tracer
from opentelemetry.propagate import extract, inject
from opentelemetry.trace import SpanKind

def main():
    with tracer.start_as_current_span("handle request", context=extract(request.headers), kind=SpanKind.SERVER) as span:
        if request.method == "GET":   #ListRecommendations
            span.add_event("invoke ListRecommendations")
            try:
                user_id = request.args.get("user_id")
                product_ids = []
                raw_product_ids = request.args.get("product_ids")
                if raw_product_ids != None:
                    product_ids = raw_product_ids.split(",")
                req = rest.ListRecommendationsRequest(user_id, product_ids)
                headers = {}
                inject(headers)
                resp = rest.recommendationservice.listRecommendations(headers, req)
                resp_body = json.dumps(resp.toDict())
                span.add_event("successfully handle request")
                return Response(response=resp_body, status=http.HTTPStatus.OK)
            except Exception as e:
                print(e)
                span.add_event("an error occurred in ListRecommendations")
                return Response(status=http.HTTPStatus.BAD_REQUEST)
        else:
            print("methods other than GET are not supported")
            span.add_event("methods other than GET are not supported")
            return Response(status=http.HTTPStatus.BAD_REQUEST)