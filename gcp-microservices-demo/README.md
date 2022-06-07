# gcp-microservices-demo
A web application based on https://github.com/GoogleCloudPlatform/microservices-demo

The application is a web-based e-commerce app where users can browse items, add them to the cart, and purchase them.

## Screenshots
| Home Page | Checkout Screen |
| ----------|-----------------|
| [![Screenshot of store homepage](./docs/img/gcpdemo-frontend-1.png)](./docs/img/gcpdemo-frontend-1.png) | [![Screenshot of checkout screen](./docs/img/gcpdemo-frontend-2.png)](./docs/img/gcpdemo-frontend-2.png) |

## Quickstart(local)
1. **Launch your k8s cluster and deploy the fission framework.**

2. **Deploy fission functions:**
```
cd fission-deploy && fission spec apply
```
You should see:
```
DeployUID: f18a3c0f-9c60-4b0b-8961-2499415a8dc3
Resources:
 * 9 Functions
 * 3 Environments
 * 9 Packages 
 * 9 Http Triggers 
 * 0 MessageQueue Triggers
 * 0 Time Triggers
 * 0 Kube Watchers
 * 9 ArchiveUploadSpec
Validation Successful
......
```

3. **Deploy frontend service & redis:**
```
cd kubernetes-manifests/ && kubectl apply -f frontend.yaml -f redis.yaml
```

4. **Wait for the Pods to be ready.**
```
kubectl get pods -n gcpdemo
```
After a few minutes, you should see:
```
NAME                             READY   STATUS    RESTARTS   AGE
frontend-5949958478-8s9bs        1/1     Running   0          3m5s
redis-cart-57bd646894-m4qgj      1/1     Running   0          3m6s
```

5. **Deploy loadgenerator(optional):**

Use loadgenerator without web UI:
```
cd kubernetes-manifests/ && kubectl apply -f loadgenerator.yaml
```

Use loadgenerator with web UI:
```
cd kubernetes-manifests/ && kubectl apply -f loadgenerator-with-webui.yaml
```

6. **Access the web frontend in a browser** using the frontend's nodeport service.
```
kubectl get service frontend-external -n gcpdemo | awk '{print $5}'
```
Example output - do not copy
```
PORT(S)
80:56510/TCP
```

7. [Optional]**Clean up:**
```
cd kubernetes-manifests/ && kubectl delete -f .
cd fission-deploy && fission spec destroy
```

## Architecture
**gcp-microservices-demo** is composed of 11 microservices written in different
languages that talk to each other over **http/json**. See the [API Documentation](./docs/api-documentation.md) doc for more information.

Unlike the original GoogleCloudPlatform/microservices-demo, gcp-microservices-demo transform microservices into fission functions, except for frontend and loadgenerator.

[![Architecture of
microservices](./docs/img/architecture-diagram.png)](./docs/img/architecture-diagram.png)

| Service | Language | Description | Type |
| ------- | -------- | ----------- | ---- |
| [frontend](./src/frontend) | Go | Exposes an HTTP server to serve the website. Does not require signup/login and generates session IDs for all users automatically. | k8s Deployment |
| [loadgenerator](./src/loadgenerator) | Python/Locust | Continuously sends requests imitating realistic user shopping flows to the frontend. | k8s Deployment |
| [cartservice](./src/cartservice) | Python | Stores the items in the user's shopping cart in Redis and retrieves it. | Function |
| [productcatalogservice](./src/productcatalogservice) | Go | Provides the list of products from a JSON file and ability to search products and get individual products. | Function |
| [currencyservice](./src/currencyservice) | Node.js | Converts one money amount to another currency. Uses real values fetched from European Central Bank. It's the highest QPS service. | Function |
| [paymentservice](./src/paymentservice) | Node.js | Charges the given credit card info (mock) with the given amount and returns a transaction ID. | Function |
| [shippingservice](./src/shippingservice) | Go | Gives shipping cost estimates based on the shopping cart. Ships items to the given address (mock) | Function |
| [emailservice](./src/emailservice) | Python | Sends users an order confirmation email (mock). | Function |
| [checkoutservice](./src/checkoutservice) | Go | Retrieves user cart, prepares order and orchestrates the payment, shipping and the email notification. | Function |
| [recommendationservice](./src/recommendationservice) | Python | Recommends other products based on what's given in the cart. | Function |
| [adservice](./src/adservice) | Go | Provides text ads based on given context words. | Function |
