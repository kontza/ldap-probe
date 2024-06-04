# TL;DR
1. Set up your `~/.ldap-probe.yaml`:

    ```yaml
    base-dn: <YOUR BASE DN>
    bind-dn: <YOUR BIND DN>
    dial-url: <your LDAP URL>
    ```
1. Build this app.
2. Make sure you have a valid password in `$HOME/.adpass`. 
1. Run this app with a term to search for.
