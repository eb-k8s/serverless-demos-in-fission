import unittest
import rest
import random
import string

class TestCartService(unittest.TestCase):
    def test_getCartIfEmpty(self):
        user_id = ''.join(random.sample(string.ascii_letters + string.digits, 10))
        expect = rest.Cart(None, [])
        res = rest.cartservice.getCart(rest.GetCartRequest(user_id))
        self.assertDictEqual(expect.toDict(), res.toDict(), "the cart should be empty!")
    
    def test_emptyCart(self):
        user_id = ''.join(random.sample(string.ascii_letters + string.digits, 10))
        expect = rest.Cart(None, [])
        rest.cartservice.emptyCart(rest.EmptyCartRequest(user_id))
        res = rest.cartservice.getCart(rest.GetCartRequest(user_id))
        self.assertDictEqual(expect.toDict(), res.toDict(), "the cart should be empty!")
    
    def test_addItemIfNotExist(self):
        user_id = ''.join(random.sample(string.ascii_letters + string.digits, 10))
        expect = rest.Cart(user_id, [rest.CartItem("shoes", 2)])
        newItem = rest.CartItem("shoes", 2)
        rest.cartservice.addItem(rest.AddItemRequest(user_id, newItem))
        res = rest.cartservice.getCart(rest.GetCartRequest(user_id))
        self.assertDictEqual(expect.toDict(), res.toDict(), "number of shoes should be 2!")
        # clean up
        rest.cartservice.emptyCart(rest.EmptyCartRequest(user_id))

    def test_addItemIfExist(self):
        user_id = ''.join(random.sample(string.ascii_letters + string.digits, 10))
        expect = rest.Cart(user_id, [rest.CartItem("shoes", 5)])
        rest.cartservice.addItem(rest.AddItemRequest(user_id, rest.CartItem("shoes", 2)))
        existedItem = rest.CartItem("shoes", 3)
        rest.cartservice.addItem(rest.AddItemRequest(user_id, existedItem))
        res = rest.cartservice.getCart(rest.GetCartRequest(user_id))
        self.assertDictEqual(expect.toDict(), res.toDict(), "number of shoes should be 5!")
        # clean up
        rest.cartservice.emptyCart(rest.EmptyCartRequest(user_id))


if __name__ == "__main__":
    unittest.main()