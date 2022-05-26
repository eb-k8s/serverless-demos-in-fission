from .rest import *

redisHost = "redis-cart.gcpdemo.svc.cluster.local"
redisPort = 6379
cartservice = CartService(redisHost, redisPort)
