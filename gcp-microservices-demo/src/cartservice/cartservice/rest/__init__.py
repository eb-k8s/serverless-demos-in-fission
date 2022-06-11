from .rest import *

redisHost = "redis-cart.gcpdemo.svc"
redisPort = 6379
cartservice = CartService(redisHost, redisPort)