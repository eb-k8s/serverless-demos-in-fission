import json
from flask import request, Response
import rest
import http
from otel import tracer
from opentelemetry.propagate import extract
from opentelemetry.trace import SpanKind

def main():
    with tracer.start_as_current_span("handle request", context=extract(request.headers), kind=SpanKind.SERVER) as span:
        if request.method == "POST":    #AddItem
            span.add_event("invoke AddItem")
            try:
                body = request.get_data()
                req = rest.dict2AddItemRequest(json.loads(body))
                rest.cartservice.addItem(req)
                span.add_event("successfully handle request")
                return Response(status=http.HTTPStatus.OK)
            except Exception as e:
                print(e)
                span.add_event("an error occurred in AddItem")
                return Response(status=http.HTTPStatus.BAD_REQUEST)
        elif request.method == "GET":   #GetCart
            span.add_event("invoke GetCart")
            try:
                user_id = request.args.get("user_id")
                req = rest.GetCartRequest(user_id)
                resp = rest.cartservice.getCart(req)
                resp_body = json.dumps(resp.toDict())
                span.add_event("successfully handle request")
                return Response(response=resp_body, status=http.HTTPStatus.OK)
            except Exception as e:
                print(e)
                span.add_event("an error occurred in GetCart")
                return Response(status=http.HTTPStatus.BAD_REQUEST)
        elif request.method == "DELETE":    #EmptyCart
            span.add_event("invoke EmptyCart")
            try:
                body = request.get_data()
                req = rest.dict2EmptyCartRequest(json.loads(body))
                rest.cartservice.emptyCart(req)
                span.add_event("successfully handle request")
                return Response(status=http.HTTPStatus.OK)
            except Exception as e:
                print(e)
                span.add_event("an error occurred in EmptyCart")
                return Response(status=http.HTTPStatus.BAD_REQUEST)
        else:
            print("methods other than POST and GET and DELETE are not supported")
            span.add_event("methods other than POST and GET and DELETE are not supported")
            return Response(status=http.HTTPStatus.BAD_REQUEST)