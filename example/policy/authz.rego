package demo.authz
import rego.v1

default allow := false

allow if {
 input.env.auth.role == "reader"
 input.env.auth.tenantId == "acme"
 input.path == "Query.customer"
 input.params.id > 0
}

allow if {
 input.env.auth.role == "reader"
 input.env.auth.tenantId == "acme"
 input.path == "Query.customer.customerItems"
 input.parent.id > 0
 input.params.limit > 0
 input.params.limit <= 100
}

allow if {
 input.env.auth.role == "reader"
 input.env.auth.tenantId == "acme"
 input.path == "Query.customers"
 input.params.limit > 0
 input.params.limit <= 100
}

allow if {
 input.env.auth.role == "reader"
 input.env.auth.tenantId == "acme"
 input.path == "Query.customers.items"
 input.parent.id > 0
 input.params.limit > 0
 input.params.limit <= 100
}

# Type round-trip demo: echoes supplied values, no table access.
allow if {
 input.env.auth.role == "reader"
 input.env.auth.tenantId == "acme"
 input.path == "Query.typeDemo"
}

allow if {
 input.env.auth.role == "reader"
 input.env.auth.tenantId == "acme"
 input.path == "Query.numericDemo"
}

allow if {
 input.env.auth.role == "reader"
 input.env.auth.tenantId == "acme"
 input.path == "Query.customerPage"
}

allow if {
 input.env.auth.role == "reader"
 input.env.auth.tenantId == "acme"
 input.path == "Query.customerPage.pageItems"
 input.parent.id > 0
}
