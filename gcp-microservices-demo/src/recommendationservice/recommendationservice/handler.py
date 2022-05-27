import json
from flask import request, Response
import rest
import http

def main():
    if request.method == "GET":   #ListRecommendations
        try:
            user_id = request.args.get("user_id")
            product_ids = []
            raw_product_ids = request.args.get("product_ids")
            if raw_product_ids != None:
                product_ids = raw_product_ids.split(",")
            req = rest.ListRecommendationsRequest(user_id, product_ids)
            resp = rest.recommendationservice.listRecommendations(req)
            resp_body = json.dumps(resp.toDict())
            return Response(response=resp_body, status=http.HTTPStatus.OK)
        except Exception as e:
            print(e)
            return Response(status=http.HTTPStatus.BAD_REQUEST)
    else:
        print("methods other than GET are not supported")
        return Response(status=http.HTTPStatus.BAD_REQUEST)