Conditional Allow / Limit TO 

==========

A simple plugin to allow a semi split-DNS option, allowing you to restrict certain zones to specific IP ranges.

## Using

Example 
```
. {
    limitto 2001:db8::/32 192.0.2.0/24
    forward . 1.1.1.1
}
```
The paramaters to the directive specify which IP ranges to allow