import json
import random
import requests
from otel import tracer
from opentelemetry.propagate import extract
from opentelemetry.trace import SpanKind

class ListRecommendationsRequest:
    def __init__(self, user_id, product_ids):
        self.user_id = user_id
        self.product_ids = product_ids

    def toDict(self):
        return {"user_id": self.user_id, "product_ids": self.product_ids}

def dict2ListRecommendationsRequest(dic):
    user_id, product_ids = None, None
    if "user_id" in dic:
        user_id = dic["user_id"]
    if "product_ids" in dic:
        product_ids = dic["product_ids"]
    return ListRecommendationsRequest(user_id, product_ids)

class ListRecommendationsResponse:
    def __init__(self, product_ids):
        self.product_ids = product_ids
    
    def toDict(self):
        return {"product_ids": self.product_ids}

def dict2ListRecommendationsResponse(dic):
    product_ids = None
    if "product_ids" in dic:
        product_ids = dic["product_ids"]
    return ListRecommendationsResponse(product_ids)

class Money:
    def __init__(self, currency_code, units, nanos):
        self.currency_code = currency_code
        self.units = units
        self.nanos = nanos
    
    def toDict(self):
        return {"currency_code": self.currency_code, "units": self.units, "nanos": self.nanos}

def dict2Money(dic):
    currency_code, units, nanos = None, None, None
    if "currency_code" in dic:
        currency_code = dic["currency_code"]
    if "units" in dic:
        units = dic["units"]
    if "nanos" in dic:
        nanos = dic["nanos"]
    return Money(currency_code, units, nanos)

class Product:
    def __init__(self, pid, name, description, picture, price_usd, categories):
        self.id = pid
        self.name = name
        self.description = description
        self.picture = picture
        self.price_usd = price_usd
        self.categories = categories
    
    def toDict(self):
        return {"id": self.id, "name": self.name, 
        "description": self.description, "picture": self.picture, 
        "price_usd": self.price_usd.toDict(), "categories": self.categories}

def dict2Product(dic):
    pid, name, description, picture, price_usd, categories = None, None, None, None, None, None
    if "id" in dic:
        pid = dic["id"]
    if "name" in dic:
        name = dic["name"]
    if "description" in dic:
        description = dic["description"]
    if "picture" in dic:
        picture = dic["picture"]
    if "price_usd" in dic:
        price_usd = dict2Money(dic["price_usd"])
    if "categories" in dic:
        categories = dic["categories"]
    return Product(pid, name, description, picture, price_usd, categories)

class ListProductsResponse:
    def __init__(self, products):
        self.products = products
    
    def toDict(self):
        productsArr = []
        for product in self.products:
            productsArr.append(product.toDict())
        return{"products": productsArr}

def dict2ListProductsResponse(dic):
    products = []
    if "products" in dic:
        productsArr = dic["products"]
        for product in productsArr:
            products.append(dict2Product(product))
    return ListProductsResponse(products)


class RealProductcatalogserviceClient:
    def __init__(self, productcatalogserviceHost):
        self.url = productcatalogserviceHost

    def listProducts(self, headers):
        with tracer.start_as_current_span("invoke ListProducts", context=extract(headers), kind=SpanKind.CLIENT) as span:
            span.add_event("invoke ListProducts")
            try:
                raw_resp = requests.get(self.url, headers=headers)
                resp = dict2ListProductsResponse(json.loads(raw_resp.text))
                span.add_event("successfully invoke ListProducts")
                return resp
            except Exception as e:
                print(e)
                span.add_event("an error occurred in ListProducts")
                return ListProductsResponse([])

# For tests
class FakeProductcatalogserviceClient:
    def __init__(self):
        self.products = [
            Product(
                "OLJCESPC7Z", 
                "Sunglasses", 
                "Add a modern touch to your outfits with these sleek aviator sunglasses.",
                "/static/img/products/sunglasses.jpg",
                Money("USD", 19, 990000000),
                ["accessories"]
            ),
            Product(
                "66VCHSJNUP",
                "Tank Top",
                "Perfectly cropped cotton tank, with a scooped neckline.",
                "/static/img/products/tank-top.jpg",
                Money("USD", 18, 990000000),
                ["clothing", "tops"]
            ),
            Product(
                "1YMWWN1N4O",
                "Watch",
                "This gold-tone stainless steel watch will work with most of your outfits.",
                "/static/img/products/watch.jpg",
                Money("USD", 109, 990000000),
                ["accessories"]
            ),
            Product(
                "L9ECAV7KIM",
                "Loafers",
                "A neat addition to your summer wardrobe.",
                "/static/img/products/loafers.jpg",
                Money("USD", 89, 990000000),
                ["footwear"]
            )
        ]
    
    def listProducts(self, headers):
        return ListProductsResponse(self.products)

class Recommendationservice:
    def __init__(self, productcatalogserviceHost):
        if productcatalogserviceHost == None or productcatalogserviceHost == "":
            self.productcatalogserviceClient = FakeProductcatalogserviceClient()
        else:
            self.productcatalogserviceClient = RealProductcatalogserviceClient(productcatalogserviceHost)
    
    def listRecommendations(self, headers, listRecommendationsRequest):
        print("ListRecommendations called with userId=" + listRecommendationsRequest.user_id)
        max_responses = 5
        # fetch list of products from productcatalogservice
        cat_response = self.productcatalogserviceClient.listProducts(headers)
        product_ids = [x.id for x in cat_response.products]
        filtered_products = list(set(product_ids)-set(listRecommendationsRequest.product_ids))
        num_products = len(filtered_products)
        num_return = min(max_responses, num_products)
        # sample list of indicies to return
        indices = random.sample(range(num_products), num_return)
        # fetch product ids from indices
        prod_list = [filtered_products[i] for i in indices]
        print("[Recv ListRecommendations] product_ids={}".format(prod_list))
        # build and return response
        return ListRecommendationsResponse(prod_list)