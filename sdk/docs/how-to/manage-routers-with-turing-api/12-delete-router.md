# Delete Router

To delete a router, call the class method `delete` while specifying the router ID of the router to be deleted:

```python
# 12. Delete this router (using its router_id)
deleted_router_id = turing.Router.delete(my_router.id)
```

The return value would be the router ID of the deleted router.