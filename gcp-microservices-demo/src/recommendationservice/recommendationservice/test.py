import unittest
import rest
import random
import string

class TestRecommendationservice(unittest.TestCase):
    def test_IfProductsEmpty(self):
        user_id = ''.join(random.sample(string.ascii_letters + string.digits, 10))
        product_ids = []
        expect = ["OLJCESPC7Z", "66VCHSJNUP", "1YMWWN1N4O", "L9ECAV7KIM"]
        expect.sort()
        testrecommendationservice = rest.Recommendationservice("")  #use fake Productcatalogservice
        req = rest.ListRecommendationsRequest(user_id, product_ids)
        headers = None
        resp = testrecommendationservice.listRecommendations(headers, req).product_ids
        resp.sort()
        self.assertSequenceEqual(expect, resp, "the response is error!")
    
    def test_IfOneProduct(self):
        user_id = ''.join(random.sample(string.ascii_letters + string.digits, 10))
        product_ids = ["66VCHSJNUP"]
        expect = ["OLJCESPC7Z", "1YMWWN1N4O", "L9ECAV7KIM"]
        expect.sort()
        testrecommendationservice = rest.Recommendationservice("")  #use fake Productcatalogservice
        req = rest.ListRecommendationsRequest(user_id, product_ids)
        headers = None
        resp = testrecommendationservice.listRecommendations(headers, req).product_ids
        resp.sort()
        self.assertSequenceEqual(expect, resp, "the response is error!")
    
    def test_IfMultiProducts(self):
        user_id = ''.join(random.sample(string.ascii_letters + string.digits, 10))
        product_ids = ["L9ECAV7KIM", "OLJCESPC7Z"]
        expect = ["66VCHSJNUP", "1YMWWN1N4O"]
        expect.sort()
        testrecommendationservice = rest.Recommendationservice("")  #use fake Productcatalogservice
        req = rest.ListRecommendationsRequest(user_id, product_ids)
        headers = None
        resp = testrecommendationservice.listRecommendations(headers, req).product_ids
        resp.sort()
        self.assertSequenceEqual(expect, resp, "the response is error!")

if __name__ == "__main__":
    unittest.main()