import json
import redis

class CartItem:
    def __init__(self, product_id, quantity):
        self.product_id = product_id
        self.quantity = quantity

    def toDict(self):
        return {"product_id": self.product_id, "quantity": self.quantity}

def dict2CartItem(dic):
    product_id, quantity = None, None
    if "product_id" in dic:
        product_id = dic["product_id"]
    if "quantity" in dic:
        quantity = dic["quantity"]
    return CartItem(product_id, quantity)

class AddItemRequest:
    def __init__(self, user_id, item):
        self.user_id = user_id
        self.item = item

    def toDict(self):
        return {"user_id": self.user_id, "item": self.item.toDict()}

def dict2AddItemRequest(dic):
    user_id, item = None, None
    if "user_id" in dic:
        user_id = dic["user_id"]
    if "item" in dic:
        item = dict2CartItem(dic["item"])
    return AddItemRequest(user_id, item)

class Cart:
    def __init__(self, user_id, items):
        self.user_id = user_id
        self.items = items

    def toDict(self):
        itemsArr = []
        for item in self.items:
            itemsArr.append(item.toDict())
        return {"user_id": self.user_id, "items": itemsArr}

def dict2Cart(dic):
    user_id, items = None, []
    if "user_id" in dic:
        user_id = dic["user_id"]
    if "items" in dic:
        itemsArr = dic["items"]
        for item in itemsArr:
            items.append(dict2CartItem(item))
    return Cart(user_id, items)

class GetCartRequest:
    def __init__(self, user_id):
        self.user_id = user_id

    def toDict(self):
        return {"user_id": self.user_id}

def dict2GetCartRequest(dic):
    user_id = None
    if "user_id" in dic:
        user_id = dic["user_id"]
    return GetCartRequest(user_id)

class EmptyCartRequest:
    def __init__(self, user_id):
        self.user_id = user_id

    def toDict(self):
        return {"user_id": self.user_id}

def dict2EmptyCartRequest(dic):
    user_id = None
    if "user_id" in dic:
        user_id = dic["user_id"]
    return EmptyCartRequest(user_id)

class CartService:
    CART_FIELD_NAME = "cart"

    def __init__(self, redisHost, redisPort):
        self.redisClient = redis.Redis(host=redisHost, port=redisPort, db=0)

    def addItem(self, addItemRequest):
        print("AddItem called with userId=" + addItemRequest.user_id)
        try:
            value = self.redisClient.hget(addItemRequest.user_id, self.CART_FIELD_NAME)
            if value != None:
                cart = dict2Cart(json.loads(value))
                existingItem = None
                for item in cart.items:
                    if addItemRequest.item.product_id == item.product_id:
                        existingItem = item
                        break
                if existingItem == None:
                    cart.items.append(addItemRequest.item)
                else:
                    existingItem.quantity += addItemRequest.item.quantity
            else:
                cart = Cart(addItemRequest.user_id, [addItemRequest.item])
            self.redisClient.hset(addItemRequest.user_id, self.CART_FIELD_NAME, json.dumps(cart.toDict()))
        except redis.exceptions.ConnectionError:
            print("cannot connect redis database!")

    def getCart(self, getCartRequest):
        print("GetCart called with userId=" + getCartRequest.user_id)
        try:
            value = self.redisClient.hget(getCartRequest.user_id, self.CART_FIELD_NAME)
            if value != None:
                return dict2Cart(json.loads(value))
            else:
                # return an empty Cart, maybe frontend service cannot deal with it
                print("the cart is empty")
                return Cart(None, [])
        except redis.exceptions.ConnectionError:
            print("cannot connect redis database!")
            return Cart(None, [])

    def emptyCart(self, emptyCartRequest):
        print("EmptyCart called with userId=" + emptyCartRequest.user_id)
        try:
            self.redisClient.hdel(emptyCartRequest.user_id, self.CART_FIELD_NAME)
        except redis.exceptions.ConnectionError:
            print("cannot connect redis database!")