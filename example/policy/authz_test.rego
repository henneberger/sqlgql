package demo.authz_test
import rego.v1
import data.demo.authz.allow

test_reader_allowed if {
 allow with input as {"env":{"auth":{"role":"reader","tenantId":"acme"}},"path":"Query.customer","params":{"id":1}}
}
test_wrong_role_denied if {
 not allow with input as {"env":{"auth":{"role":"admin","tenantId":"acme"}},"path":"Query.customer","params":{"id":1}}
}
test_wrong_tenant_denied if {
 not allow with input as {"env":{"auth":{"role":"reader","tenantId":"other"}},"path":"Query.customer","params":{"id":1}}
}
test_nested_allowed if {
 allow with input as {"env":{"auth":{"role":"reader","tenantId":"acme"}},"path":"Query.customer.customerItems","parent":{"id":1},"params":{"limit":100}}
}
test_nested_limit_denied if {
 not allow with input as {"env":{"auth":{"role":"reader","tenantId":"acme"}},"path":"Query.customer.customerItems","parent":{"id":1},"params":{"limit":101}}
}
test_unknown_path_denied if {
 not allow with input as {"env":{"auth":{"role":"reader","tenantId":"acme"}},"path":"Query.unknown"}
}
test_missing_identity_denied if {
 not allow with input as {"path":"Query.customer","params":{"id":1}}
}
