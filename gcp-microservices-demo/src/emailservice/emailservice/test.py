import unittest
import rest

class TestDummyEmailService(unittest.TestCase):
    def test_sendOrderConfirmation(self):
        req = rest.SendOrderConfirmationRequest(
            "abc@example.com", 
            rest.OrderResult(
                "xxx-yyy-zzz",
                "1234-5678-90",
                rest.Money("USD", 92, 990000000),
                rest.Address("A", "B", "C", "D", 100000),
                [
                    rest.OrderItem(
                        rest.CartItem("OLJCESPC7Z", 2), 
                        rest.Money("USD", 38, 990000000)
                    ),
                    rest.OrderItem(
                        rest.CartItem("66VCHSJNUP", 3), 
                        rest.Money("USD", 54, 990000000)
                    ),
                ]
            )
        )
        rest.emailservice.sendOrderConfirmation(req)

if __name__ == "__main__":
    unittest.main()