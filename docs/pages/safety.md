updated 2023-2-23

> BE WARNED that these are early days for pgpkg. I have made an attempt to
> provide a framework to enable security via Postgresql primitives (such as
> roles and schemas), but at this time you should consider any package you use to
> potentially have the ability to read and write to your database.
> 
> pgpkg is great for internal reuse of code, but an adversary would certainly
> be able to defeat the measures currently in use at this time.

doesn't check schema conformance
doesn't provide any security
