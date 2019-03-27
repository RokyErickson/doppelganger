# SSH
Doppelganger provides support for synchronizing with filesystem locations accessible via SSH.
Doppelganger requires an OpenSSH client installation to be available on your system.
SSH synchronization endpoints can be specified to Doppelganger's `create` command
using URL syntax:

    [user@]host[:port]:path

The `user` component is optional. 

The `host` component can be any IP address, hostname, or alias understood by OpenSSH.

The `port` component is also optional.

The `path` component: 

a absoulte path
	
	user@host:/var/www

a relative path 

    user@host:path/in/home/directory

a home-directory-relative path, e.g.

    user@host:~/path/in/home/directory
-----------------------------------------


**Note** SSH support is much more stable now. Prompts function and connections to endpoints are stable. However if you do get an ssh error output doppelganger will hang and you will need to sigterm doppelganger and the ssh/doppelganger-agent process. Then, tweak your ssh till you no longer get an error output.
