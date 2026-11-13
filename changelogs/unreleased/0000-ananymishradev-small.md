HTTPProxy and Gateway API `Host` header rewrites now reject invalid literal hostnames, including IP addresses, wildcards, ports, uppercase labels, and trailing dots, instead of passing them to Envoy.
