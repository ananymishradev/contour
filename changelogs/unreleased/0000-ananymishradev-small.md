Contour now checks literal `Host` rewrites and rejects values that are not
simple lowercase DNS names. This prevents invalid host rewrite settings in
HTTPProxy and Gateway API configurations.
