from .rest import *
import os

domain = os.getenv("DOMAIN")
recommendationservice = None
if domain == "":
    print("DOMAIN not set, skipping communicating with other functions")
    recommendationservice = Recommendationservice()
else:
    print("recommendationservice with domain: %s" % domain)
    productcatalogserviceHost = "http://" + domain + "/product" #DOMAIN value don't have "http://" prefix
    recommendationservice = Recommendationservice(productcatalogserviceHost)