# Get Router

To get a specific router from Turing API, you will need its ID when running the following:

```python
# 3. Get the router you just created using the router_id obtained
my_router = turing.Router.get(my_router.id)
```

The return value would be a `Router` object representing the router with the corresponding ID.
