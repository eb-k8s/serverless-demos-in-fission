import json
from flask import request, Response
import rest
import http

def main():
    if request.method == "POST":   #SendOrderConfirmation
        try:
            body = request.get_data()
            req = rest.dict2SendOrderConfirmationRequest(json.loads(body))
            rest.emailservice.sendOrderConfirmation(req)
            return Response(status=http.HTTPStatus.OK)
        except Exception as e:
            print(e)
            return Response(status=http.HTTPStatus.BAD_REQUEST)
    else:
        print("methods other than POST are not supported")
        return Response(status=http.HTTPStatus.BAD_REQUEST)