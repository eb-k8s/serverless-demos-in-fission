from .rest import *

productcatalogserviceHost = "http://router.fission.svc.cluster.local/product"
recommendationservice = Recommendationservice(productcatalogserviceHost)