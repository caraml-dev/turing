# Traffic Rules

{% hint style="info" %}
This step is **optional** and the default behaviour will not discriminate between requests.
{% endhint %}

![](../../.gitbook/assets/create_router_rules.png)

It is also possible to configure your router such that each request is only dispatched to a subset of the configured routes, based on some request specific conditions. For example, you might have some models, trained on geography-specific data. So, in this scenario, you want to call `model-a` if the request contains header `"X-Region: Region-A"` or `model-b` if `X-Region` header equals `"Region-B"`. It's possible to achieve this by configuring traffic rules on your router.

Traffic rules define which routes should be "activated" for a particular request to your router. Each rule is defined by one or more request conditions, and one or more routes that would be activated if the request satisfies the conditions of the rule. Routes that are not part of any traffic rules will be called for each request.

Rules are matched against the incoming request in the order in which they are defined. This property can be used to create one ore more specific rules over a general rule. Consider the following example, where a router has rules defined such that requests are routed by the country of origin (ID, SG, etc.). In addition, if the routing logic must be altered for a certain service type in a country, this rule (the 'specific' rule) can be defined before the general rule.

![](../../.gitbook/assets/create_router_rules_priority.png)

### Conditions

Each rule should have at least one condition configured on it. If there are multiple conditions configured on the same rule, then this rule will be triggered only if each and every condition is satisfied.  Rule condition can be defined on either request header or request payload (assuming payload is a valid JSON object).  For each condition you should specify:

* **Condition source**: either `Header` or `Payload`
* **Condition key**: if condition's source is `Header` – then the name of a request Header (example: `X-Session-ID`), or else, if condition's source is `Payload` – a valid JSON path of the property from the request's JSON payload (example: `service_type.id` or `users.0.name`)
* **Condition values**: one or more values that the extracted condition key is expected to match. Condition will be satisfied if key matches at least one of the configured values.<br/>

{% hint style="warning" %}
Provided values are case-sensitive.
{% endhint %}

### Routes   

You should also select from the drop-down list one or more routes, that would be activated if this rule is triggered. It's not allowed to include the default route into any of traffic rules. As for non-default routes, they could be attached to zero or more traffic rules:

* If route is attached to some traffic rule, then Turing will only send request to this route if the request meets this rule's conditions.
* If route is not attached to any of the traffic rules, then Turing call this route with every incoming request.
* If route is attached to multiple rules and the request satisfies more than one rule, then Turing will decide what group of routes should receive this request based on the order in which the traffic rules are defined.
