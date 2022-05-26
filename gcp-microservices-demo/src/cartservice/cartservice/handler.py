import json
from flask import request, Response
import rest
import http

def main():
    if request.method == "POST":    #AddItem
        try:
            body = request.get_data()
            req = rest.dict2AddItemRequest(json.loads(body))
            rest.cartservice.addItem(req)
            return Response(status=http.HTTPStatus.OK)
        except Exception as e:
            print(e)
            return Response(status=http.HTTPStatus.BAD_REQUEST)
    elif request.method == "GET":   #GetCart
        try:
            user_id = request.args.get("user_id")
            req = rest.GetCartRequest(user_id)
            resp = rest.cartservice.getCart(req)
            resp_body = json.dumps(resp.toDict())
            return Response(response=resp_body, status=http.HTTPStatus.OK)
        except Exception as e:
            print(e)
            return Response(status=http.HTTPStatus.BAD_REQUEST)
    elif request.method == "DELETE":    #EmptyCart
        try:
            body = request.get_data()
            req = rest.dict2EmptyCartRequest(json.loads(body))
            rest.cartservice.emptyCart(req)
            return Response(status=http.HTTPStatus.OK)
        except Exception as e:
            print(e)
            return Response(status=http.HTTPStatus.BAD_REQUEST)
    else:
        print("methods other than POST and GET and DELETE are not supported")
        return Response(status=http.HTTPStatus.BAD_REQUEST)